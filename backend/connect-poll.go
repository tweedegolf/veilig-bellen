package main

import (
	
	"time"
)

type ConnectPoll struct {
	bc Broadcast
}

func makeConnectPoll() ConnectPoll {
	bc := makeBroadcast()
	return ConnectPoll{bc}
}

func (poll *ConnectPoll) registerListener(listener chan string) {
	poll.bc.registerListener(listener)
}

func (poll *ConnectPoll) unregisterListener(listener chan string, close bool) {
	poll.bc.unregisterListener(listener, close)
}

func (poll *ConnectPoll) update(update string) {
	poll.bc.update(update)
}

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
			// log.Printf("tick %v ", status)
			poll.bc.update(status)
		}
	}
}

func pollConnect() string {
	return "{\"DataSnapshotTime\":\"2020-04-15T20:49:11Z\",\"MetricResults\":[{\"Collections\":[{\"Metric\":{\"Name\":\"AGENTS_ONLINE\",\"Unit\":\"COUNT\"},\"Value\":1},{\"Metric\":{\"Name\":\"AGENTS_AVAILABLE\",\"Unit\":\"COUNT\"},\"Value\":1},{\"Metric\":{\"Name\":\"AGENTS_ON_CALL\",\"Unit\":\"COUNT\"},\"Value\":0},{\"Metric\":{\"Name\":\"CONTACTS_IN_QUEUE\",\"Unit\":\"COUNT\"},\"Value\":0}],\"Dimensions\":null}],\"NextToken\":null}"
}
