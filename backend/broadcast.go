package main

type Update struct {
	msg string
}

type registerOp struct {
	listener chan Update
}

type unregisterOp struct {
	listener chan Update
	close    bool
}

type Broadcast struct {
	listeners     []chan Update
	registerOps   chan registerOp
	unregisterOps chan unregisterOp
	updates       chan Update
}

func (bc Broadcast) registerListener(listener chan Update) {
	bc.registerOps <- registerOp{listener}
}

func (bc Broadcast) unregisterListener(listener chan Update, close bool) {
	bc.unregisterOps <- unregisterOp{listener, close}
}

func (bc Broadcast) update(update Update) {
	bc.updates <- update
}

func makeBroadcast() Broadcast {
	listeners := make([]chan Update, 10)
	registerOps := make(chan registerOp)
	unregisterOps := make(chan unregisterOp)
	updates := make(chan Update, 20)
	return Broadcast{listeners, registerOps, unregisterOps, updates}
}

func (bc Broadcast) removeListener(listener chan Update, closeChan bool) {
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

func (bc Broadcast) daemon() {
	for {
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
}
