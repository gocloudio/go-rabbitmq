package dispatcher

import (
	"log"
	"sync"
	"time"
)

// Dispatcher -
type Dispatcher struct {
	subscribers   map[*dispatchSubscriber]struct{}
	subscribersMu *sync.Mutex
}

type dispatchSubscriber struct {
	notifyCancelOrCloseChan chan error
	closeCh                 <-chan struct{}
}

// NewDispatcher -
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		subscribers:   make(map[*dispatchSubscriber]struct{}),
		subscribersMu: &sync.Mutex{},
	}
}

// Dispatch -
func (d *Dispatcher) Dispatch(err error) error {
	d.subscribersMu.Lock()
	defer d.subscribersMu.Unlock()
	for subscriber := range d.subscribers {
		select {
		case <-time.After(time.Second * 5):
			log.Println("Unexpected rabbitmq error: timeout in dispatch")
		case subscriber.notifyCancelOrCloseChan <- err:
		}
	}
	return nil
}

// AddSubscriber -
func (d *Dispatcher) AddSubscriber() (<-chan error, chan<- struct{}) {
	closeCh := make(chan struct{})
	notifyCancelOrCloseChan := make(chan error)
	subscriber := &dispatchSubscriber{
		notifyCancelOrCloseChan: notifyCancelOrCloseChan,
		closeCh:                 closeCh,
	}

	d.subscribersMu.Lock()
	d.subscribers[subscriber] = struct{}{}
	d.subscribersMu.Unlock()

	go func() {
		<-closeCh
		d.subscribersMu.Lock()
		defer d.subscribersMu.Unlock()
		close(subscriber.notifyCancelOrCloseChan)
		delete(d.subscribers, subscriber)
	}()
	return notifyCancelOrCloseChan, closeCh
}
