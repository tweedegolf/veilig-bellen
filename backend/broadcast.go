package main

import "log"

type Message struct {
	Session string  	`json:"session"`
	Key     string  	`json:"key"`
	Value   interface{}	`json:"value"`
}

type Listener chan Message

type registerOp struct {
	listener Listener
}

type unregisterOp struct {
	listener Listener
	close    bool
}

type Broadcast struct {
	listeners     []Listener
	registerOps   chan registerOp
	unregisterOps chan unregisterOp
	messages      Listener
	stopped       chan struct{}
}

// Schedule registering a listener
func (bc *Broadcast) registerListener(listener Listener) {
	bc.registerOps <- registerOp{listener}
}

// Schedule unregistering a listener
func (bc *Broadcast) unregisterListener(listener Listener, close bool) {
	bc.unregisterOps <- unregisterOp{listener, close}
}

// Schedule sending a message to all listeners
func (bc *Broadcast) notify(message Message) {
	bc.messages <- message
}

// Make a new Broadcast object
func makeBroadcast() Broadcast {
	listeners := make([]Listener, 0)
	registerOps := make(chan registerOp)
	unregisterOps := make(chan unregisterOp)
	messages := make(Listener, 20)
	stopped := make(chan struct{})
	return Broadcast{listeners, registerOps, unregisterOps, messages, stopped}
}

// Finds and remove a listener
func (bc *Broadcast) removeListener(listener Listener, closeChan bool) {
	index := -1
	for i, l := range bc.listeners {
		if l == listener {
			index = i
			break
		}
	}
	if index != -1 {
		if closeChan {
			close(bc.listeners[index])
		}
		bc.listeners[index] = bc.listeners[len(bc.listeners)-1]
		bc.listeners = bc.listeners[:len(bc.listeners)-1]
	}
}

// Runs a command from one of the queues.
// Blocks if none is available until Broadcast.Close is called.
// Returns whether this method should be run again in a loop
func (bc *Broadcast) nextUnitOfWork() bool {
	select {
	case message := <-bc.messages:
		for _, listener := range bc.listeners {
			select {
			case listener <- message:
			default:
				log.Printf("Broadcast dropped message")
				// Discard message if listener's buffer is full
			}
		}
	case op := <-bc.registerOps:
		bc.listeners = append(bc.listeners, op.listener)
	case op := <-bc.unregisterOps:
		bc.removeListener(op.listener, op.close)
	case <-bc.stopped:
		return false
	}
	return true
}

// Signal the Broadcast that it should stop waiting for work
func (bc *Broadcast) Close() {
	bc.stopped <- struct{}{}
}

// Continuously run commands from each of the queues whenever they
// become available.
func (bc *Broadcast) daemon() {
	for bc.nextUnitOfWork() {
	}
}
