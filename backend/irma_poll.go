package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	irma "github.com/privacybydesign/irmago"
)

// A citizen has started an IRMA session and we're waiting for them to finish.
// When the IRMA session finishes, we store the disclosed attributes in the
// database.
func (cfg Configuration) pollIrmaSessionDaemon(sessionToken string) {
	sessionUpdates := make(chan Message)
	cfg.broadcaster.Subscribe(sessionToken, sessionUpdates)
	defer cfg.broadcaster.Unsubscribe(sessionToken, sessionUpdates)

	var status string
	ticker := time.NewTicker(time.Second)
	transport := irma.NewHTTPTransport(
		fmt.Sprintf("%s/session/%s/", cfg.IrmaServerURL, sessionToken))

	for {
		select {
		case update := <-sessionUpdates:
			if update.Key == "backend_id" && update.Value != cfg.db.backendIdentity {
				// Polling responsibility was taken by another
				// backend.
				return
			}

		case <-ticker.C:
			var new_status string
			err := transport.Get("status", &new_status)
			if err != nil {
				log.Printf("failed to get irma session status: %v", err)
				new_status = "UNREACHABLE"
			}

			new_status = strings.Trim(new_status, `"`)
			new_status = "IRMA-" + new_status
			if new_status == status {
				continue
			}

			// Store disclosed attributes in database
			// *before* DONE notification
			if new_status == "DONE" {
				cfg.cacheDisclosedAttributes(sessionToken)
			}

			err = cfg.db.setStatus(sessionToken, new_status)
			if err != nil {
				log.Printf("failed to store irma session status: %v", err)
			}

			status = new_status
			if IrmaStatusIsFinal(status) {
				return
			}
		}
	}
}

// Decides whether we should stop polling based on a returned
// irma status message
func IrmaStatusIsFinal(status string) bool {
	return status != "IRMA-INITIALIZED" && status != "IRMA-CONNECTED"
}
