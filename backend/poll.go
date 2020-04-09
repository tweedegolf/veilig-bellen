package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

import "github.com/privacybydesign/irmago"

type createOp struct {
	sessionToken string
	listener     chan<- string
}

type destroyOp struct {
	sessionToken string
}

type notifyOp struct {
	sessionToken string
	status       string
}

// Irma polling facade type
type IrmaPoll struct {
	listeners  map[string][]chan<- string
	createOps  chan createOp
	destroyOps chan destroyOp
	notifyOps  chan notifyOp
}

// Create a new IrmaPoll
func makeIrmaPoll() IrmaPoll {
	listeners := make(map[string][]chan<- string)
	createOps := make(chan createOp)
	destroyOps := make(chan destroyOp)
	notifyOps := make(chan notifyOp)
	return IrmaPoll{listeners, createOps, destroyOps, notifyOps}
}

// Register a new irma listener for the given session
func (poll IrmaPoll) createIrmaListener(sessionToken string, irmaStatus chan<- string) {
	poll.createOps <- createOp{sessionToken, irmaStatus}
}

/// Close and drop all listeners for the given sessionToken
func (poll IrmaPoll) destroyAllIrmaListeners(sessionToken string) {
	poll.destroyOps <- destroyOp{sessionToken}
}

// Notify all listeners for the given sessionToken with the status
func (poll IrmaPoll) notifyListeners(sessionToken string, status string) {
	poll.notifyOps <- notifyOp{sessionToken, status}
}

// Polls irma server continuously. Each registered sessionToken is polled once
// every second.
// Handles listener creation, destroy, and notification operation
// messages.
func pollDaemon(cfg Configuration) {
	transport := irma.NewHTTPTransport("")
	ticker := time.NewTicker(1000 * time.Millisecond)
	poll := &cfg.irmaPoll

	var status string

	for {
		select {
		case <-ticker.C:
			// Do the polling in the background in case this takes a long time.
			// This guarantees sessions are polled every second
			go func() {
				for sessionToken := range poll.listeners {
					// Update the request server URL to include the session token.
					transport.Server = cfg.IrmaServerURL + fmt.Sprintf("/session/%s/", sessionToken)
					status = pollIrmaSession(transport)

					poll.notifyListeners(sessionToken, status)
					if shouldStopPolling(status) {
						poll.destroyAllIrmaListeners(sessionToken)
					}
				}
			}()
		case op := <-poll.createOps:
			poll.listeners[op.sessionToken] = append(poll.listeners[op.sessionToken], op.listener)
		case op := <-poll.destroyOps:
			for _, irmaStatus := range poll.listeners[op.sessionToken] {
				close(irmaStatus)
			}
			delete(poll.listeners, op.sessionToken)
		case op := <-poll.notifyOps:
			for _, irmaStatus := range poll.listeners[op.sessionToken] {
				irmaStatus <- status
			}
		}
	}
	log.Printf("Stopped polling Irma server")
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
