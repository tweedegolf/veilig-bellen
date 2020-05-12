package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	irma "github.com/privacybydesign/irmago"

	"github.com/kelseyhightower/envconfig"
	"github.com/lib/pq"
	flag "github.com/spf13/pflag"
)

// Every backend node will ask the database to expire old sessions once every
// ExpireDelay
const ExpireDelay = time.Hour

type ConnectConfiguration struct {
	// Amazon endpoint user identifier
	Id string `json:"id,omitempty"`
	// Amazon endpoint user secret
	Secret string `json:"secret,omitempty"`
	// Amazon Connect instance identifier
	InstanceId string `json:"instance,omitempty"`
	Queue      string `json:"queue,omitempty"`
	Region     string `json:"region,omitempty"`
}

type BaseConfiguration struct {
	Configuration   string               `json:"configuration,omitempty"`
	Database        string               `json:"database,omitempty"`
	ListenAddress   string               `json:"listen-address,omitempty"`
	InternalAddress string               `json:"internal-address,omitempty"`
	IrmaServer      string               `json:"irma-server,omitempty"`
	IrmaHeaderKey   string               `json:"irma-header-key,omitempty"`
	IrmaHeaderValue string               `json:"irma-header-value,omitempty"`
	IrmaExternalURL string               `json:"irma-external-url,omitempty"`
	PhoneNumber     string               `json:"phone-number,omitempty"`
	PurposeMap      string               `json:"purpose-map,omitempty"`
	Connect         ConnectConfiguration `json:"connect,omitempty"`
}

type Configuration struct {
	Database        string
	ListenAddress   string
	InternalAddress string
	IrmaServer      string
	IrmaHeaderKey   string
	IrmaHeaderValue string
	IrmaExternalURL string
	PhoneNumber     string
	PurposeMap      map[string]irma.AttributeConDisCon
	Connect         ConnectConfiguration
	db              Database
	broadcaster     Broadcaster
}

func resolveConfiguration(base BaseConfiguration) Configuration {
	var cfg Configuration

	cfg.Database = base.Database
	cfg.ListenAddress = base.ListenAddress
	cfg.InternalAddress = base.InternalAddress
	cfg.IrmaServer = base.IrmaServer
	cfg.IrmaHeaderKey = base.IrmaHeaderKey
	cfg.IrmaHeaderValue = base.IrmaHeaderValue
	cfg.IrmaExternalURL = base.IrmaExternalURL
	cfg.PhoneNumber = base.PhoneNumber
	if base.PurposeMap != "" {
		err := json.Unmarshal([]byte(base.PurposeMap), &cfg.PurposeMap)
		if err != nil {
			panic(fmt.Sprintf("could not parse purpose map: %v", err))
		}
	}
	cfg.Connect = base.Connect

	return cfg
}

func main() {
	var baseCfg BaseConfiguration

	configuration := flag.StringP("config", "c", "", `The file to read configuration from. Further options override.`)
	database := flag.String("database", "", `The address of the PostgreSQL database used for persistence and robustness.`)
	listenAddress := flag.String("listen-address", "", `The address to listen for external requests, e.g. ":8080".`)
	internalAddress := flag.String("internal-address", "", `The address to listen for internal requests such as /call. Defaults to listen-address.`)
	irmaServer := flag.String("irma-server", "", `The address of the IRMA server to use for disclosure.`)
	irmaHeaderKey := flag.String("irma-header-key", "", `The header key to send with IRMA server session requests, i.e. Authorization. Defaults to Authorization. Will only be used if value is also set.`)
	irmaHeaderValue := flag.String("irma-header-value", "", `The header value to send with IRMA server session requests, i.e. the token passphrase. Will be sent as Authorization if key is not set. No header will be added if not set.`)
	irmaExternalURL := flag.String("irma-external-url", "", `The IRMA base url as shown to users in the app`)
	phoneNumber := flag.String("phone-number", "", `The service number citizens will be directed to call.`)
	purposeMap := flag.String("purpose-map", "", `The map from purposes to attribute condiscons.`)

	// Note: we do not provide the option to set the ID & Secret using CLI.
	connectInstanceId := flag.String("connect-instance-id", "", `Identifier of the Amazon Connect instance`)
	connectQueue := flag.String("connect-queue", "", `Identifier of the Amazon Connect queue to show the metrics for`)
	connectRegion := flag.String("connect-region", "", `The Amazon Connect region to use (i.e. eu-central-1)`)

	flag.Parse()

	err := envconfig.Process("BACKEND", &baseCfg)
	if err != nil {
		panic(fmt.Sprintf("could not parse environment: %v", err))
	}

	if *configuration != "" {
		baseCfg.Configuration = *configuration
	}

	if baseCfg.Configuration != "" {
		contents, err := ioutil.ReadFile(baseCfg.Configuration)
		if err != nil {
			panic(fmt.Sprintf("configuration file not found: %v", baseCfg.Configuration))
		}
		err = json.Unmarshal(contents, &baseCfg)
		if err != nil {
			panic(fmt.Sprintf("could not parse configuration file: %v", err))
		}
	}

	if *database != "" {
		baseCfg.Database = *database
	}
	if *listenAddress != "" {
		baseCfg.ListenAddress = *listenAddress
	}
	if *internalAddress != "" {
		baseCfg.InternalAddress = *internalAddress
	}
	if *irmaServer != "" {
		baseCfg.IrmaServer = *irmaServer
	}
	if *irmaHeaderKey != "" {
		baseCfg.IrmaHeaderKey = *irmaHeaderKey
	}
	if *irmaHeaderValue != "" {
		baseCfg.IrmaHeaderValue = *irmaHeaderValue
	}
	if *irmaExternalURL != "" {
		baseCfg.IrmaExternalURL = *irmaExternalURL
	}
	if *phoneNumber != "" {
		baseCfg.PhoneNumber = *phoneNumber
	}
	if *purposeMap != "" {
		baseCfg.PurposeMap = *purposeMap
	}
	if *connectInstanceId != "" {
		baseCfg.Connect.InstanceId = *connectInstanceId
	}
	if *connectQueue != "" {
		baseCfg.Connect.Queue = *connectQueue
	}
	if *connectRegion != "" {
		baseCfg.Connect.Region = *connectRegion
	}

	cfg := resolveConfiguration(baseCfg)

	if cfg.Database == "" {
		panic("option required: database")
	}
	if cfg.ListenAddress == "" {
		panic("option required: listen-address")
	}
	if cfg.IrmaServer == "" {
		panic("option required: irma-server")
	}
	if cfg.PhoneNumber == "" {
		panic("option required: phone-number")
	}
	if cfg.PurposeMap == nil {
		panic("option required: purpose-map")
	}
	if cfg.IrmaHeaderKey != "" && cfg.IrmaHeaderValue == "" {
		panic("irma-header-value is required when setting irma-header-key")
	}

	log.Printf("Successfully parsed configuration")

	var identity string
	hostname, err := os.Hostname()
	if err != nil {
		identity = fmt.Sprintf(":%v", os.Getpid())
	} else {
		identity = fmt.Sprintf("%v:%v", hostname, os.Getpid())
	}

	db, err := sql.Open("postgres", cfg.Database)
	if err != nil {
		panic(fmt.Errorf("could not connect to database: %w", err))
	}
	listener := pq.NewListener(cfg.Database, 100*time.Millisecond, 60*time.Second, nil)
	cfg.db = Database{
		db:              db,
		listener:        listener,
		backendIdentity: identity,
	}

	log.Printf("Connected to database")

	// The open call may succeed because the library seems to connect to the
	// database lazily. Expire old sessions in order to test the connection.
	err = cfg.db.expire()
	if err != nil {
		panic(fmt.Errorf("could not connect to database: %w", err))
	}

	cfg.broadcaster = makeBroadcaster()

	if cfg.Connect.Id != "" && cfg.Connect.Secret != "" {
		// This is expected to fail for every backend but the first to
		// call it because the feed will already exist. We leave it to
		// the adoptDaemon to start the connectPollDaemon.
		cfg.db.NewFeed("kcc")
	} else {
		log.Printf("warning: Amazon Connect credentials not provided")
	}

	// TODO: Fail immediately if configured Irma server
	// can't be reached before entering ListenAndServe.
	go adoptDaemon(cfg)
	go expireDaemon(cfg)
	go notifyDaemon(cfg)

	log.Printf("Registered polling processes")

	externalMux := http.NewServeMux()
	externalMux.HandleFunc("/", cfg.handleStatus)
	externalMux.HandleFunc("/session", cfg.handleSession)
	externalMux.HandleFunc("/session/status", cfg.handleSessionStatus)
	externalMux.HandleFunc("/metrics", cfg.handleMetrics)
	externalMux.HandleFunc("/session/update", cfg.handleSessionUpdate)
	externalMux.HandleFunc("/disclose", cfg.handleDisclose)
	externalMux.HandleFunc("/agent-feed", cfg.handleAgentFeed)

	if cfg.InternalAddress != "" && cfg.InternalAddress != cfg.ListenAddress {
		internalMux := http.NewServeMux()
		internalMux.HandleFunc("/call", cfg.handleCall)
		internalServer := http.Server{
			Addr:    cfg.InternalAddress,
			Handler: internalMux,
		}
		go internalServer.ListenAndServe()
		log.Printf("Started internal HTTP server on %v", cfg.InternalAddress)
	} else {
		externalMux.HandleFunc("/call", cfg.handleCall)
	}

	externalServer := http.Server{
		Addr:    cfg.ListenAddress,
		Handler: externalMux,
	}
	log.Printf("Starting external HTTP server on %v", cfg.ListenAddress)
	log.Fatal(externalServer.ListenAndServe())
}

func adoptDaemon(cfg Configuration) {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		feeds, err := cfg.db.AdoptOrphans()
		if err != nil {
			log.Printf("failed to adopt orphans: %v", err)
			continue
		}

		for _, feed := range feeds {
			if feed == "kcc" {
				if cfg.Connect.Id != "" && cfg.Connect.Secret != "" {
					go connectPollDaemon(cfg)
				} else {
					// Secret not available, disabling feed.
					log.Printf("no connect credentials available, disabling kcc feed")
					cfg.db.DeleteFeed("kcc")
				}
			} else {
				go cfg.pollIrmaSessionDaemon(feed)
			}
		}
	}
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
