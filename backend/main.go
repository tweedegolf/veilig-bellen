package main

import "database/sql"
import "encoding/json"
import "fmt"
import "io/ioutil"
import "log"
import "net/http"
import "time"

import "github.com/privacybydesign/irmago"
import flag "github.com/spf13/pflag"
import _ "github.com/lib/pq"

// Every backend node will ask the database to expire old sessions once every
// ExpireDelay
const ExpireDelay = time.Hour

type Configuration struct {
	PostgresAddress     string                             `json:"database,omitempty"`
	ListenAddress       string                             `json:"listen-address,omitempty"`
	IrmaServerURL       string                             `json:"irma-server,omitempty"`
	ServicePhoneNumber  string                             `json:"phone-number,omitempty"`
	PurposeToAttributes map[string]irma.AttributeConDisCon `json:"purpose-map,omitempty"`
	db                  Database
}

func main() {
	var cfg Configuration

	configuration := flag.StringP("config", "c", "", `The file to read configuration from. Further options override.`)
	database := flag.String("database", "", `The address of the PostgreSQL database used for persistence and robustness.`)
	listenAddress := flag.String("listen-address", "", `The address to listen for external requests, e.g. ":8080".`)
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
	// TODO: The pq driver doesn't fail until we try to use the database.
	if err != nil {
		panic("could not connect to database")
	}
	cfg.db = Database{db}

	// TODO: Fail immediately if configured Irma server or configured database
	// can't be reached before entering ListenAndServe.
	go expireDaemon(cfg)

	http.HandleFunc("/call", cfg.handleCall)
	http.HandleFunc("/session", cfg.handleSession)
	http.HandleFunc("/disclose", cfg.handleDisclose)
	http.ListenAndServe(cfg.ListenAddress, nil)
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
