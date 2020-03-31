package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

import "github.com/privacybydesign/irmago"

/**
one thread keeps going over all currently active sessions, wich are in the db

	-> polls irma server for status
	-> sends status through corresponding channel

Requests can register an irma listener
*/

var listeners = make(map[string][]chan<- string)

func createIrmaListener(sessionToken string, irmaStatus chan<- string) {
	if listeners[sessionToken] == nil {
		listeners[sessionToken] = make([]chan<- string, 0)
	}
	listeners[sessionToken] = append(listeners[sessionToken], irmaStatus)
}

func destroyIrmaListener(sessionToken string) {
	if listeners[sessionToken] == nil {
		return
	}

	for _, irmaStatus := range listeners[sessionToken] {
		close(irmaStatus)
	}
	listeners[sessionToken] = nil
}

func notifyIrmaListeners(sessionToken string, status string) {

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
			transport.Server = cfg.IrmaServerURL + transport.Server + fmt.Sprintf("session/%s/", sessionToken)
			status = pollIrmaSession(transport)

			if status == "" {
				destroyIrmaListener(sessionToken)
			} else {
				notifyIrmaListeners(sessionToken, status)
				if status == "DONE" {
					destroyIrmaListener(sessionToken)
				}
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
