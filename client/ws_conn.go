package client

import (
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// wsConnection is an abstraction interface for the websocket connection communication
// to allow for mocking in tests
type wsConnection interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	SendCloseMessage() error
	Close() error
}

// wsConn is a wrapper for the gorilla websocket connection
type wsConn struct {
	hasSubscriptions bool
	subscribersMutex sync.Mutex
	conn             *ws.Conn
}

// ReadJSON reads a JSON message from the websocket connection
func (c *wsConn) ReadJSON(v interface{}) error {
	return c.conn.ReadJSON(v)
}

// WriteJSON writes a JSON message to the websocket connection
func (c *wsConn) WriteJSON(v interface{}) error {
	return c.conn.WriteJSON(v)
}

func (c *wsConn) Close() error {
	return c.conn.Close()
}

func (c *wsConn) SendCloseMessage() error {

	// send close message through the websocket connection
	err := c.conn.WriteMessage(ws.CloseMessage, ws.FormatCloseMessage(ws.CloseNormalClosure, ""))
	if err != nil {
		log.Error().Msgf("Error sending close message: %s", err)
		return err
	}

	// wait for the close message to be received or timeout
	select {
	case <-time.After(time.Second):
		log.Info().Msg("Timeout waiting for close message")
	default:
		// We ignore the message
		c.conn.ReadMessage()
	}
	return nil
}
