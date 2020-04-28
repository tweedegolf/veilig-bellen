package main

import "strings"

import "github.com/lib/pq"

type Message struct {
	Key   string
	Value string
}

type subscription struct {
	feed     string
	listener chan<- Message
}

// Broadcaster Irma polling facade type
type Broadcaster struct {
	registerOps   chan subscription
	unregisterOps chan subscription
}

// Create a new Broadcaster
func makeBroadcaster() Broadcaster {
	registerOps := make(chan subscription, 10)
	unregisterOps := make(chan subscription, 10)
	return Broadcaster{registerOps, unregisterOps}
}

// Subscribe a go channel to an update feed
func (b Broadcaster) Subscribe(feed string, ch chan<- Message) {
	b.registerOps <- subscription{feed, ch}
}

// Schedule unregistering a listener
func (b Broadcaster) Unsubscribe(feed string, ch chan<- Message) {
	b.unregisterOps <- subscription{feed, ch}
}

func splitNotification(n *pq.Notification) (string, Message) {
	if n == nil {
		return "meta", Message{"reconnect", ""}
	}

	var key, value string
	session := n.Channel
	key = strings.TrimSpace(n.Extra)
	spaceIndex := strings.IndexByte(key, byte(' '))
	if spaceIndex >= 0 {
		value = key[spaceIndex+1:]
		key = key[:spaceIndex]
	}

	return session, Message{key, value}

}

// The notifyDaemon forwards notifications incoming from postgres to all
// interested channels.
func notifyDaemon(cfg Configuration) {
	channels := make(map[string][]chan<- Message)
	notificationChannel := cfg.db.listener.NotificationChannel()

	for {
		select {
		case notification := <-notificationChannel:
			session, message := splitNotification(notification)
			channels := channels[session]
			for _, listener := range channels {
				select {
				case listener <- message:
				default:
					// Message discarded
				}
			}

		case op := <-cfg.broadcaster.registerOps:
			cfg.db.listener.Listen(op.feed)
			channels[op.feed] = append(channels[op.feed], op.listener)

		case op := <-cfg.broadcaster.unregisterOps:
			subscribers := channels[op.feed]
			for i, l := range subscribers {
				if l == op.listener {
					subscribers[i] = subscribers[len(channels)-1]
					subscribers = subscribers[:len(channels)-1]
					channels[op.feed] = subscribers
					if len(subscribers) == 0 {
						cfg.db.listener.Unlisten(op.feed)
					}
					break
				}
			}
		}
	}
}
