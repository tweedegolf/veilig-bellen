package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	irma "github.com/privacybydesign/irmago"

	flag "github.com/spf13/pflag"

	_ "github.com/lib/pq"
)

// Every backend node will ask the database to expire old sessions once every
// ExpireDelay
const ExpireDelay = time.Hour

type Configuration struct {
	PostgresAddress     string                             `json:"database,omitempty"`
	ListenAddress       string                             `json:"listen-address,omitempty"`
	InternalAddress     string                             `json:"internal-address,omitempty"`
	IrmaServerURL       string                             `json:"irma-server,omitempty"`
	ServicePhoneNumber  string                             `json:"phone-number,omitempty"`
	PurposeToAttributes map[string]irma.AttributeConDisCon `json:"purpose-map,omitempty"`
	db                  Database
	irmaPoll            IrmaPoll
	connectPoll   		ConnectPoll
}

func main() {
	var cfg Configuration

	configuration := flag.StringP("config", "c", "", `The file to read configuration from. Further options override.`)
	database := flag.String("database", "", `The address of the PostgreSQL database used for persistence and robustness.`)
	listenAddress := flag.String("listen-address", "", `The address to listen for external requests, e.g. ":8080".`)
	internalAddress := flag.String("internal-address", "", `The address to listen for internal requests such as /call. Defaults to listen-address.`)
	irmaServer := flag.String("irma-server", "", `The address of the IRMA server to use for disclosure.`)
	phoneNumber := flag.String("phone-number", "", `The service number citizens will be directed to call.`)
	purposeMap := flag.String("purpose-map", "", `The map from purposes to attribute condiscons.`)

	flag.Parse()

	if *configuration != "" {
		contents, err := ioutil.ReadFile(*configuration)
		if err != nil {
			panic(fmt.Sprintf("configuration file not found: %v", *configuration))
		}
		err = json.Unmarshal(contents, &cfg)
		if err != nil {
			panic(fmt.Sprintf("could not parse configuration file: %v", err))
		}
	}

	if *database != "" {
		cfg.PostgresAddress = *database
	}
	if *listenAddress != "" {
		cfg.ListenAddress = *listenAddress
	}
	if *internalAddress != "" {
		cfg.InternalAddress = *internalAddress
	}
	if *irmaServer != "" {
		cfg.IrmaServerURL = *irmaServer
	}
	if *phoneNumber != "" {
		cfg.ServicePhoneNumber = *phoneNumber
	}
	if *purposeMap != "" {
		err := json.Unmarshal([]byte(*purposeMap), &cfg.PurposeToAttributes)
		if err != nil {
			panic(fmt.Sprintf("could not parse purpose map: %v", err))
		}
	}

	if cfg.PostgresAddress == "" {
		panic("option required: database")
	}
	if cfg.ListenAddress == "" {
		panic("option required: listen-address")
	}
	if cfg.IrmaServerURL == "" {
		panic("option required: irma-server")
	}
	if cfg.ServicePhoneNumber == "" {
		panic("option required: phone-number")
	}
	if cfg.PurposeToAttributes == nil {
		panic("option required: purpose-map")
	}

	db, err := sql.Open("postgres", cfg.PostgresAddress)
	if err != nil {
		panic(fmt.Errorf("could not connect to database: %w", err))
	}
	cfg.db = Database{db}

	cfg.irmaPoll = makeIrmaPoll()
	cfg.connectPoll = makeConnectPoll()
	// The open call may succeed because the library seems to connect to the
	// database lazily. Expire old sessions in order to test the connection.
	err = cfg.db.expire()
	if err != nil {
		panic(fmt.Errorf("could not connect to database: %w", err))
	}

	// TODO: Fail immediately if configured Irma server
	// can't be reached before entering ListenAndServe.
	go expireDaemon(cfg)
	go irmaPollDaemon(cfg)
	go connectPollDaemon(cfg)

	externalMux := http.NewServeMux()
	externalMux.HandleFunc("/session", cfg.handleSession)
	externalMux.HandleFunc("/session/status", cfg.handleSessionStatus)
	externalMux.HandleFunc("/disclose", cfg.handleDisclose)
	externalMux.HandleFunc("/agent-feed", cfg.handleWaitlistFeed)

	if cfg.InternalAddress != "" && cfg.InternalAddress != cfg.ListenAddress {
		internalMux := http.NewServeMux()
		internalMux.HandleFunc("/call", cfg.handleCall)
		internalServer := http.Server{
			Addr:    cfg.InternalAddress,
			Handler: internalMux,
		}
		go internalServer.ListenAndServe()
	} else {
		externalMux.HandleFunc("/call", cfg.handleCall)
	}

	externalServer := http.Server{
		Addr:    cfg.ListenAddress,
		Handler: externalMux,
	}
	externalServer.ListenAndServe()
}

func expireDaemon(cfg Configuration) {
	for {
		err := cfg.db.expire()
		if err != nil {
			log.Printf("failed to expire old database entries: %v", err)
		}
		time.Sleep(ExpireDelay)
	}
}
