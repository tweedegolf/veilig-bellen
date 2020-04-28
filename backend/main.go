package main

import "database/sql"
import "encoding/json"
import "fmt"
import "io/ioutil"
import "log"
import "net/http"
import "time"
import "os"

import "github.com/privacybydesign/irmago"
import flag "github.com/spf13/pflag"
import _ "github.com/lib/pq"

// Every backend node will ask the database to expire old sessions once every
// ExpireDelay
const ExpireDelay = time.Hour

type ConnectConfiguration struct {
	id           string                                    `json:"id,omitempty"`
	secret       string                                    `json:"secret,omitempty"`
	instanceId   string                                    `json:"instance,omitempty"`
	queue        string                                    `json:"queue,omitempty"`
	region       string                                    `json:"region,omitempty"`
}

type Configuration struct {
	PostgresAddress     string                             `json:"database,omitempty"`
	ListenAddress       string                             `json:"listen-address,omitempty"`
	InternalAddress     string                             `json:"internal-address,omitempty"`
	IrmaServerURL       string                             `json:"irma-server,omitempty"`
	ServicePhoneNumber  string                             `json:"phone-number,omitempty"`
	PurposeToAttributes map[string]irma.AttributeConDisCon `json:"purpose-map,omitempty"`
	db                  Database
	irmaPoll            IrmaPoll
	connect             ConnectConfiguration
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

	connectId := os.Getenv("CONNECT_ID") // Amazon endpoint user identifier
	connectSecret := os.Getenv("CONNECT_SECRET") // Amazon endpoint user secret

	connectInstanceId := flag.String("connect-instance-id", "", `Identifier of the Amazon Connect instance`)
	connectQueue := flag.String("connect-queue", "", `Identifier of the Amazon Connect queue to show the metrics for`)
	connectRegion := flag.String("connect-region", "", `The Amazon Connect region to use (i.e. eu-central-1)`)

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

	if connectId != "" && connectSecret != "" {
		cfg.connect = ConnectConfiguration{
			id: connectId,
			secret: connectSecret,
			instanceId: *connectInstanceId,
			queue: *connectQueue,
			region: *connectRegion,
		}
	} else {
		log.Printf("warning: Amazon Connect credentials not provided")
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
	// The open call may succeed because the library seems to connect to the
	// database lazily. Expire old sessions in order to test the connection.
	err = cfg.db.expire()
	if err != nil {
		panic(fmt.Errorf("could not connect to database: %w", err))
	}

	// TODO: Fail immediately if configured Irma server
	// can't be reached before entering ListenAndServe.
	go expireDaemon(cfg)
	go pollDaemon(cfg)

	externalMux := http.NewServeMux()
	externalMux.HandleFunc("/session", cfg.handleSession)
	externalMux.HandleFunc("/disclose", cfg.handleDisclose)
	externalMux.HandleFunc("/session/status", cfg.handleSessionStatus)
	externalMux.HandleFunc("/metrics", cfg.handleMetrics)
	externalMux.HandleFunc("/session/update", cfg.handleSessionUpdate)

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
