package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

import "github.com/privacybydesign/irmago"

type CreateOp struct {
	sessionToken string
	listener     chan<- string
}

type NotifyOp struct {
	sessionToken string
	status       string
}

type Session struct {
	channels []chan<- string
	status   string
	created  time.Time
}

type IrmaPoll struct {
	sessions  map[string]*Session
	createOps chan CreateOp
	notifyOps chan NotifyOp
}

// Create a new IrmaPoll
func makeIrmaPoll() IrmaPoll {
	sessions := make(map[string]*Session)
	createOps := make(chan CreateOp, 10)
	notifyOps := make(chan NotifyOp, 10)
	return IrmaPoll{sessions, createOps, notifyOps}
}

// Register a new irma listener for the given session
func (poll IrmaPoll) createIrmaListener(sessionToken string, irmaStatus chan<- string) {
	poll.createOps <- CreateOp{sessionToken, irmaStatus}
}

// Try to send a status update. If the channel's buffer is full,
// the status update is discarded. This way, sending status messages
// never blocks the pollDaemon if the listener is never received from.
func (session *Session) tryNotify(status string) {
	session.status = status
	for _, channel := range session.channels {
		select {
		case channel <- status:
			// Message sent
			break;
		default:
			// Message discarded
		}
	}
}

func (session *Session) addChannel(channel chan<- string) {
	session.channels = append(session.channels, channel)
}

func (session Session) shouldPollIrma() bool {
	status := session.status
	return status == "INIT" || status == "IRMA-INITIALIZED" || status == "IRMA-CONNECTED"
}

// Notify all listeners for the given sessionToken with the status
func (poll IrmaPoll) tryNotify(sessionToken string, status string) {
	poll.notifyOps <- NotifyOp{sessionToken, status}
}

func (poll *IrmaPoll) findOrCreate(sessionToken string) *Session {
	if poll.sessions[sessionToken] == nil {
		poll.sessions[sessionToken] = &Session{
			channels: make([]chan<- string, 10),
			status:   "INIT",
			created:  time.Now(),
		}
	}
	return poll.sessions[sessionToken]
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
			for sessionToken, session := range poll.sessions {
				// Update the request server URL to include the session token.
				if session.shouldPollIrma() {
					transport.Server = cfg.IrmaServerURL + fmt.Sprintf("/session/%s/", sessionToken)
					status = "IRMA-" + pollIrmaSession(transport)
					session.tryNotify(status)
				
					err := cfg.db.updateSessionStatus(sessionToken, status)
					if err != nil {
						log.Printf("IrmaPoll failed to update session status: %#v", err);
					}
				}

				// Clean up all sessions after 2 hours regardless.
				if time.Since(session.created).Hours() > 2 {
					// Close and delete all listeners for this session.
					for _, channel := range session.channels {
						if channel != nil {
							close(channel)
						}
					}
					delete(poll.sessions, sessionToken)
				}
			}
		case op := <-poll.createOps:
			poll.findOrCreate(op.sessionToken).addChannel(op.listener)
		case op := <-poll.notifyOps:
			poll.findOrCreate(op.sessionToken).tryNotify(op.status)
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
