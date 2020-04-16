package main

type registerOp struct {
	listener chan string
}

type unregisterOp struct {
	listener chan string
	close    bool
}

type Broadcast struct {
	listeners     []chan string
	registerOps   chan registerOp
	unregisterOps chan unregisterOp
	updates       chan string
	stopped       chan interface{}
}

// Schedule registering a listener
func (bc *Broadcast) registerListener(listener chan string) {
	bc.registerOps <- registerOp{listener}
}

// Schedule unregistering a listener
func (bc *Broadcast) unregisterListener(listener chan string, close bool) {
	bc.unregisterOps <- unregisterOp{listener, close}
}

// Schedule sending a message to all listeners
func (bc *Broadcast) update(update string) {
	bc.updates <- update
}

func makeBroadcast() Broadcast {
	listeners := make([]chan string, 0)
	registerOps := make(chan registerOp)
	unregisterOps := make(chan unregisterOp)
	updates := make(chan string, 20)
	stopped := make(chan interface{})
	return Broadcast{listeners, registerOps, unregisterOps, updates, stopped}
}

// Finds and remove a listener
func (bc *Broadcast) removeListener(listener chan string, closeChan bool) {
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
	case update := <-bc.updates:
		for _, listener := range bc.listeners {
			select {
			case listener <-  update:
			default:
				// Discard message if listener's buffer is full
			}
		}
	case op := <-bc.registerOps:
		bc.listeners = append(bc.listeners, op.listener)
	case op := <-bc.unregisterOps:
		
		bc.removeListener(op.listener, op.close)
	case _ = <-bc.stopped:
		return false
	}
	return true
}

// Signal the Broadcast that it shoulds stop waiting for work
func (bc *Broadcast) Close() {
	bc.stopped <- nil
}

// Continuously run commands from each of the queues whenever they
// become available.
func (bc *Broadcast) daemon() {
	for bc.nextUnitOfWork() {}
}
