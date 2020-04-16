package main

import (
	
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
func (poll *ConnectPoll) registerListener(listener chan string) {
	poll.bc.registerListener(listener)
}

// Schedule unregistering a listener
func (poll *ConnectPoll) unregisterListener(listener chan string, close bool) {
	poll.bc.unregisterListener(listener, close)
}

// Schedule sending a message to all listeners
func (poll *ConnectPoll) update(update string) {
	poll.bc.update(update)
}

// Connect poll daemon. To be run in a separate thread.
// Polls Amazon Connect continuously and sends status updates
// to all listeners.
func connectPollDaemon(cfg Configuration) {
	poll := &cfg.connectPoll
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer poll.bc.Close()
	go poll.bc.daemon()

	var status string
	for {
		select {
		case <-ticker.C:
			status = pollConnect()
			poll.bc.update(status)
		}
	}
}


// TODO get status from amazon connect
func pollConnect() string {
	return "{\"DataSnapshotTime\":\"2020-04-15T20:49:11Z\",\"MetricResults\":[{\"Collections\":[{\"Metric\":{\"Name\":\"AGENTS_ONLINE\",\"Unit\":\"COUNT\"},\"Value\":1},{\"Metric\":{\"Name\":\"AGENTS_AVAILABLE\",\"Unit\":\"COUNT\"},\"Value\":1},{\"Metric\":{\"Name\":\"AGENTS_ON_CALL\",\"Unit\":\"COUNT\"},\"Value\":0},{\"Metric\":{\"Name\":\"CONTACTS_IN_QUEUE\",\"Unit\":\"COUNT\"},\"Value\":0}],\"Dimensions\":null}],\"NextToken\":null}"
}
