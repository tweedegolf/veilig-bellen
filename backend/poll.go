package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

import "github.com/privacybydesign/irmago"

var listeners = make(map[string][]chan<- string)

// Register a new irma listener for the given session
func createIrmaListener(sessionToken string, irmaStatus chan<- string) {
	if listeners[sessionToken] == nil {
		listeners[sessionToken] = make([]chan<- string, 0)
	}
	listeners[sessionToken] = append(listeners[sessionToken], irmaStatus)
}

/// Close and drop all listeners for the given sessionToken
func destroyAllIrmaListeners(sessionToken string) {
	if listeners[sessionToken] == nil {
		return
	}

	for _, irmaStatus := range listeners[sessionToken] {
		close(irmaStatus)
	}
	delete(listeners, sessionToken)
}

// Notify all listeners for the given sessionToken with the status
func notifyIrmaListeners(sessionToken string, status string) {
	if listeners[sessionToken] == nil {
		return
	}
	for _, irmaStatus := range listeners[sessionToken] {
		irmaStatus <- status
	}
}

// Polls irma server continuously
//
func pollDaemon(cfg Configuration) {
	transport := irma.NewHTTPTransport("")
	ticker := time.NewTicker(1000 * time.Millisecond)

	var status string
	for range ticker.C {
		for sessionToken := range listeners {
			// Update the request server URL to include the session token.
			transport.Server = cfg.IrmaServerURL + fmt.Sprintf("/session/%s/", sessionToken)
			status = pollIrmaSession(transport)
			
			if status == "" {
				status = "ERR"
			}

			notifyIrmaListeners(sessionToken, status)
			if shouldStopPolling(status) {
				destroyAllIrmaListeners(sessionToken)
			}

			time.Sleep(100)
		}

		time.Sleep(1000)
	}
}

// Poll the irma session
func pollIrmaSession(transport *irma.HTTPTransport) string {
	var status string

	err := transport.Get("status", &status)
	if err != nil {
		log.Printf("failed to get irma session status: %v", err)
		time.Sleep(time.Second)
		return ""
	}
	return strings.Trim(status, `"`)
}

// Decides whether we should stop polling based on a returned
// irma status message
func shouldStopPolling(status string) bool {
	return status == "DONE" || status == "TIMEOUT" || status == "CANCELLED" || status == "ERR"
}
