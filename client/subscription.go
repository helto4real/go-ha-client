package client

import (
	"context"

	"github.com/rs/zerolog/log"
)

// eventSubscription is a struct to hold the subscription to a topic
type eventSubscription struct {
	id       string
	ch       chan HaResult
	callback OnEvent
	subId    int64

	ctx         context.Context
	cancelFunc  context.CancelFunc
    haConn *haConnection
}

// unsubscribe unsubscribes from the topic and cancels the processing
func (es *eventSubscription) Unsubscribe() {
    es.cancelFunc()
}

// newEventSubscription creates a new eventSubscription instance
func newEventSubscription(ch chan HaResult, callback OnEvent, subId int64, haConn *haConnection) *eventSubscription {

    ctx, cancel := context.WithCancel(context.Background())
    es := &eventSubscription{
		ch:       ch,
		callback: callback,
		subId:    subId,
		ctx:        ctx,
		cancelFunc: cancel,
        haConn: haConn,
	}
    go es.processEvents()
    return es
}

// processEvents processes the events from the channel and calls the callback using goroutines
func (es *eventSubscription) processEvents() {
    for {
        select {
        case <-es.ctx.Done():
            return
        case result := <-es.ch:
            log.Info().Msgf("Got event: %v, my sub id: %s", result, es.id)
            if result.Id == es.subId && result.Event != nil {
                go es.callback(result.Event)
            }
        }
    }
}
