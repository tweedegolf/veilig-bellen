package main

import (
	"log"
	"time"
)

// ConnectPoll polls Amazon connect for waitlist statistics and passes it
// to the registered listeners. Uses a Broadcast internally to manage
// listeners and sending
type ConnectPoll struct {
	bc Broadcast
}

// Make a new ConnectPoll
func makeConnectPoll() ConnectPoll {
	bc := makeBroadcast()
	return ConnectPoll{bc}
}

// Schedule registering a listener
func (poll *ConnectPoll) registerListener(listener Listener) {
	poll.bc.registerListener(listener)
}

// Schedule unregistering a listener
func (poll *ConnectPoll) unregisterListener(listener Listener, close bool) {
	poll.bc.unregisterListener(listener, close)
}

// Schedule sending a message to all listeners
func (poll *ConnectPoll) notify(message Message) {
	poll.bc.notify(message)
}

// Connect poll daemon. To be run in a separate thread.
// Polls Amazon Connect continuously and sends status updates
// to all listeners.
func connectPollDaemon(cfg Configuration) {
	poll := &cfg.connectPoll
	pollTicker := time.NewTicker(1000 * time.Millisecond)
	dbTicker := time.NewTicker(1000 * time.Millisecond)
	defer poll.bc.Close()
	go poll.bc.daemon()

	for {

		select {
		case <-pollTicker.C:
			response, err := cfg.getConnectCurrentMetrics()

			if err != nil {
				log.Printf("Connect poll could not get Amazon Connect metrics")
				continue
			}

			poll.notify(Message{"kcc", "amazon-connect", response})
		case <-dbTicker.C:
			count, err := cfg.db.activeSessionCount()
			if err != nil {
				log.Printf("Could not get active session count: %v", err)
				break
			}
			poll.notify(Message{"kcc", "active-sessions", ActiveSessionsMessage{count}})
		}
	}
}
type ActiveSessionsMessage struct {
	Count int `json:"count"`
}
