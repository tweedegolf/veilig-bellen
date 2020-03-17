package main

import "log"
import "net/http"
import "time"

import "github.com/privacybydesign/irmago"

const ExpireDelay = time.Hour

func main() {
	// TODO: Read from configuration file.
	cfg := Configuration{
		ListenAddress:      ":8080",
		IrmaServerURL:      "http://irma:8088",
		ServicePhoneNumber: "+31123456789",
		PurposeToAttributes: map[string]irma.AttributeConDisCon{
			"foo": {{{
				irma.NewAttributeRequest("irma-demo.MijnOverheid.root.BSN"),
			}}},
		},
	}

	// TODO: Fail immediately if configured Irma server or configured database
	// can't be reached before entering ListenAndServe.
	// go expireDaemon(cfg)

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
