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

// IrmaPoll Irma polling facade type
type IrmaPoll struct {
	listeners map[string][]chan<- string
	createOps chan createOp
}

// Create a new IrmaPoll
func makeIrmaPoll() IrmaPoll {
	listeners := make(map[string][]chan<- string)
	createOps := make(chan createOp, 10)
	return IrmaPoll{listeners, createOps}
}

// Register a new irma listener for the given session
func (poll IrmaPoll) createIrmaListener(sessionToken string, irmaStatus chan<- string) {
	poll.createOps <- createOp{sessionToken, irmaStatus}
}

// Try to send a status update. If the channel's buffer is full,
// the status update is discarded. This way, sending status messages
// never blocks the pollDaemon if the listener is never received from.
func tryNotifyListener(listener chan<- string, status string) {
	select {
	case listener <- status:
	default:
		// Message discarded
	}
}

// Polls irma server continuously. Each registered sessionToken is polled once
// every second.
// Handles listener creation, destroy, and notification operation
// messages.
func irmaPollDaemon(cfg Configuration) {
	transport := irma.NewHTTPTransport("")
	ticker := time.NewTicker(1000 * time.Millisecond)
	poll := &cfg.irmaPoll

	var status string

	for {
		select {
		case <-ticker.C:
			for sessionToken, statusChannels := range poll.listeners {
				// Update the request server URL to include the session token.
				transport.Server = cfg.IrmaServerURL + fmt.Sprintf("/session/%s/", sessionToken)
				status = pollIrmaSession(transport)
				// Notify all channels
				for _, irmaStatus := range statusChannels {
					tryNotifyListener(irmaStatus, status)
				}
				if shouldStopPolling(status) {
					// Close and delete all listeners for this channel
					for _, irmaStatus := range statusChannels {
						close(irmaStatus)
					}
					delete(poll.listeners, sessionToken)
				}
			}
		case op := <-poll.createOps:
			poll.listeners[op.sessionToken] = append(poll.listeners[op.sessionToken], op.listener)
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
