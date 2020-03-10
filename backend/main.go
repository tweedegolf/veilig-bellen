package main

import "net/http"

import "github.com/privacybydesign/irmago"

func main() {
	// TODO: Read from configuration file.
	cfg := Configuration{
		ListenAddress:      ":8080",
		IrmaServerURL:      "http://localhost:8088",
		ServicePhoneNumber: "+31123456789",
		PurposeToAttributes: map[string]irma.AttributeConDisCon{
			"foo": {{{
				irma.NewAttributeRequest("irma-demo.MijnOverheid.root.BSN"),
			}}},
		},
	}

	http.HandleFunc("/session", cfg.handleNewSession)
	http.ListenAndServe(cfg.ListenAddress, nil)
}
