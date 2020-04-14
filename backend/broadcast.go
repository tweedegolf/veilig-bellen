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
}

func (bc Broadcast) registerListener(listener chan string) {
	bc.registerOps <- registerOp{listener}
}

func (bc Broadcast) unregisterListener(listener chan string, close bool) {
	bc.unregisterOps <- unregisterOp{listener, close}
}

func (bc Broadcast) update(update string) {
	bc.updates <- update
}

func makeBroadcast() Broadcast {
	listeners := make([]chan string, 10)
	registerOps := make(chan registerOp)
	unregisterOps := make(chan unregisterOp)
	updates := make(chan string, 20)
	return Broadcast{listeners, registerOps, unregisterOps, updates}
}

func (bc Broadcast) removeListener(listener chan string, closeChan bool) {
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

func (bc Broadcast) nextUnitOfWork() {
	select {
	case update := <-bc.updates:
		for _, listener := range bc.listeners {
			select {
			case listener <- update:
			default:
				// Discard message if listener's buffer is full
			}
		}
	case op := <-bc.registerOps:
		bc.listeners = append(bc.listeners, op.listener)
	case op := <-bc.unregisterOps:
		bc.removeListener(op.listener, op.close)
	}
}

func (bc Broadcast) daemon() {
	for {
		bc.nextUnitOfWork()
	}
}
