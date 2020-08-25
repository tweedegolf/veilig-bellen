package main

import (
	"encoding/json"
	"log"
	"time"
)

// Keeps track of whether we have encountered
// an error with receiving Amazon Connect Metrics before
var metricError = false;

// Connect poll daemon. To be run in a separate thread.
// Polls Amazon Connect continuously and sends status updates
// to all listeners.
func connectPollDaemon(cfg Configuration) {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		ours, err := cfg.db.TouchFeed("kcc")
		if err != nil {
			log.Printf("failed to update feed in database: %v", err)
			return
		} else if !ours {
			// This feed is no longer our responsibility.
			return
		}

		response, err := cfg.getConnectCurrentMetrics()
		if !metricError && err != nil {
			log.Printf("Connect poll could not get Amazon Connect metrics")
			metricError = true;
			continue
		}

		msg, err := json.Marshal(response)
		if err != nil {
			log.Printf("failed to marshal amazon connect metrics: %v", err)
			continue
		}

		cfg.db.Notify("kcc", "amazon-connect", string(msg))
	}
}
