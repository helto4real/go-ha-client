package client

import (
	"context"
	"net/url"
	"strconv"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// haConnection is a struct to hold the connection to Home Assistant
type haConnection struct {
    // IsConnected is a flag to indicate if the connection is connected to Home Assistant
    IsConnected bool
    // msgNr is used to keep track of the message numbering for Home Assistant messages
	msgNr       int64
    // Folloing is used to prevent multiple close calls
    closeMutex  *sync.Mutex
	isClosing   bool
    // abstracted websocket connection
	wsConn      wsConnection
    // context and cancel function to be able to cancel goroutines
	ctx         context.Context
	cancelFunc  context.CancelFunc
    // event bus to handle events subscriptions
	bus         *eventBus[HaResult]
    // channel to write messages to the websocket connection in an async way
	writeCh     chan interface{}
}

// ConnectHomeAssistant connects to Home Assistant using the provided host, port, ssl and access token
func ConnectHomeAssistant(host string, port int, ssl bool, token string) (*haConnection, error) {
	scheme := "wss"
	if !ssl {
		scheme = "ws"
	}
	uri := url.URL{Scheme: scheme, Host: host + ":" + strconv.Itoa(port), Path: "/api/websocket"}

	c, _, err := ws.DefaultDialer.Dial(uri.String(), nil)

	if err != nil {
		return nil, err
	}

    // create a new context and cancel function to be able to cancel goroutines
	ctx, cancel := context.WithCancel(context.Background())

	haConn := &haConnection{
        IsConnected: true,

		isClosing:   false,
		wsConn: &wsConn{
			conn: c,
		},
		ctx:        ctx,
		cancelFunc: cancel,
		closeMutex: &sync.Mutex{},
		bus:        NewEventBus[HaResult](),
		writeCh:    make(chan interface{}, 20),
	}

	err = haConn.authorize(token)
	if err != nil {
		c.Close()
		return nil, err
	}

	go haConn.startEventLoop()
	go haConn.startWriteLoop()
	return haConn, nil
}


func (c *haConnection) Close() {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	if c.isClosing || !c.IsConnected {
		return
	}
	c.isClosing = true

	c.cancelFunc()

	c.wsConn.SendCloseMessage()
	c.wsConn.Close()

	c.IsConnected = false
}

func (c *haConnection) WriteJSON(v interface{}) {
	c.writeCh <- v
}

type OnEvent func(event *HaEvent)

func (c *haConnection) startWriteLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.writeCh:
			c.wsConn.WriteJSON(msg)
		}
	}
}

func (c *haConnection) startEventLoop() {
	defer c.Close()
	// read all json messages from the websocket connection
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg := HaResult{}
			err := c.wsConn.ReadJSON(&msg)
			if err != nil {
				if c.isClosing == false {
					log.Error().Msgf("Error reading from websocket: %v %v %v", err, c.isClosing, c.IsConnected)
				}
				return
			}
			if msg.Error != nil {
				log.Error().Msgf("Error received: %s", msg.Error.Message)
				continue
			}
            // publish the message to the event bus to be handled by the subscriptions
			c.bus.Publish("events", msg)
		}
	}
}

// authorize gets the initial message from Home Assistant and
// sends the auth message with the access token
// to authorize the connection if auth_required is received
func (c *haConnection) authorize(token string) error {

	// Initial expected message from Home Assistant on connection
	msg := HaConnectMsg{}
	err := c.wsConn.ReadJSON(&msg)
	if err != nil {
		return err
	}

	log.Info().Msgf("Connected to Home Assistant (%s)", msg.HaVersion)

	// If the message is not an auth_required message, we don't need to authorize
	if msg.MessageType != "auth_required" {
		return nil
	}

	// Send the auth message to Home Assistant
	c.wsConn.WriteJSON(HaAuthMsg{AccessToken: token, MessageType: "auth"})
	res := HaResult{}
	err = c.wsConn.ReadJSON(&res)
	if err != nil {
		return err
	}

	// If the message type is an auth_ok message, we are authorized
	if res.MessageType == "auth_ok" {
		log.Info().Msg("Successfully authorized to Home Assistant")
		return nil
	}

	log.Error().Msg("Failed to authorize to Home Assistant!")
	return errors.New(res.Error.Message)
}

// SubscribeEvents subscribes to events from Home Assistant with the provided event type or '*' for all events
func (c *haConnection) SubscribeEvents(eventType string, callback OnEvent) (*eventSubscription, error) {
	ch := c.bus.Subscribe("events")
	c.msgNr++
	subCmd := HaSubscribeEventsCommand{
		Id:          c.msgNr,
		MessageType: "subscribe_events",
		EventType:   eventType,
	}
	c.WriteJSON(subCmd)
	sub := newEventSubscription(
		ch,
		callback,
		c.msgNr,
		c,
	)
	return sub, nil
}
