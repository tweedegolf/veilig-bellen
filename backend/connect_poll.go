package main

import (
	"encoding/json"
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

	var status ConnectStatusResponse
	for {

		select {
		case <-pollTicker.C:
			status = pollConnect()
			poll.notify(Message{"kcc", "amazon-connect", status})
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

// TODO get status from amazon connect
func pollConnect() ConnectStatusResponse {
	status := "{\"DataSnapshotTime\":\"2020-04-15T20:49:11Z\",\"MetricResults\":[{\"Collections\":[{\"Metric\":{\"Name\":\"AGENTS_ONLINE\",\"Unit\":\"COUNT\"},\"Value\":1},{\"Metric\":{\"Name\":\"AGENTS_AVAILABLE\",\"Unit\":\"COUNT\"},\"Value\":1},{\"Metric\":{\"Name\":\"AGENTS_ON_CALL\",\"Unit\":\"COUNT\"},\"Value\":0},{\"Metric\":{\"Name\":\"CONTACTS_IN_QUEUE\",\"Unit\":\"COUNT\"},\"Value\":0}],\"Dimensions\":null}],\"NextToken\":null}"
	var message ConnectStatusResponse
	err := json.Unmarshal([]byte(status), &message)
	if err != nil {
		log.Printf("Could not encode connect status message %#v", err)
	}
	return message
}

type Metric struct {
	Name string
	Unit string
}

type MetricCollection struct {
	Metric Metric
	Value  interface{}
}

type MetricResult struct {
	Collections []MetricCollection
}

type ConnectStatusResponse struct {
	DataSnapshotTime string
	MetricResults    []MetricResult
}

type ActiveSessionsMessage struct {
	Count int `json:"count"`
}