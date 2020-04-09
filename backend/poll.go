package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

import "github.com/privacybydesign/irmago"

// Register a new irma listener for the given session
func (cfg Configuration) createIrmaListener(sessionToken string, irmaStatus chan<- string) {
	cfg.irmaListeners[sessionToken] = append(cfg.irmaListeners[sessionToken], irmaStatus)
}

/// Close and drop all listeners for the given sessionToken
func (cfg Configuration) destroyAllIrmaListeners(sessionToken string) {
	for _, irmaStatus := range cfg.irmaListeners[sessionToken] {
		close(irmaStatus)
	}
	delete(cfg.irmaListeners, sessionToken)
}

// Notify all listeners for the given sessionToken with the status
func (cfg Configuration) notifyIrmaListeners(sessionToken string, status string) {
	for _, irmaStatus := range cfg.irmaListeners[sessionToken] {
		irmaStatus <- status
	}
}

// Polls irma server continuously
func pollDaemon(cfg Configuration) {
	transport := irma.NewHTTPTransport("")
	ticker := time.NewTicker(1000 * time.Millisecond)

	var status string
	for range ticker.C {
		for sessionToken := range cfg.irmaListeners {
			go func() {
				// Update the request server URL to include the session token.
				transport.Server = cfg.IrmaServerURL + fmt.Sprintf("/session/%s/", sessionToken)
				status = pollIrmaSession(transport)

				cfg.notifyIrmaListeners(sessionToken, status)
				if shouldStopPolling(status) {
					cfg.destroyAllIrmaListeners(sessionToken)
				}
			}()
			time.Sleep(10)
		}
	}
}

// Poll the irma session
func pollIrmaSession(transport *irma.HTTPTransport) string {
	var status string

	err := transport.Get("status", &status)
	if err != nil {
		log.Printf("failed to get irma session status: %v", err)
		return "UNREACHABLE"
	}
	return strings.Trim(status, `"`)
}

// Decides whether we should stop polling based on a returned
// irma status message
func shouldStopPolling(status string) bool {
	return status == "DONE" || status == "TIMEOUT" || status == "CANCELLED" || status == "UNREACHABLE"
}
