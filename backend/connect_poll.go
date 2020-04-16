package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Connect poll daemon. To be run in a separate thread.
// Polls Amazon Connect continuously and sends status updates
// to all listeners.
func connectPollDaemon(cfg Configuration) {
	pollTicker := time.NewTicker(1000 * time.Millisecond)
	dbTicker := time.NewTicker(1000 * time.Millisecond)

	for {

		select {
		case <-pollTicker.C:
			response, err := cfg.getConnectCurrentMetrics()

			if err != nil {
				log.Printf("Connect poll could not get Amazon Connect metrics")
				continue
			}

			msg, err := json.Marshal(response)

			if err != nil {
				log.Printf("failed to marshal amazon connect metrics: %v", err)
				continue
			}

			cfg.db.Notify("kcc", "amazon-connect", string(msg))
		case <-dbTicker.C:
			count, err := cfg.db.activeSessionCount()
			if err != nil {
				log.Printf("Could not get active session count: %v", err)
				break
			}
			cfg.db.Notify("kcc", "active-sessions", fmt.Sprintf("%d", count))
		}
	}
}
