// Package ping implements WebSocket PING messages.
package ping

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gorilla/websocket"
)

// SendTicks sends the current ticks as a ping message.
func SendTicks(conn *websocket.Conn, start time.Time, deadline time.Time) error {
	var ticks int64 = time.Since(start).Nanoseconds()
	data, err := json.Marshal(ticks)
	if err == nil {
		err = conn.WriteControl(websocket.PingMessage, data, deadline)
	}
	return err
}

func ParseTicks(s string, start time.Time) (elapsed time.Duration, d time.Duration, err error) {
	elapsed = time.Since(start)
	var prev int64
	err = json.Unmarshal([]byte(s), &prev)
	if err != nil {
		return
	}
	if 0 <= prev && prev <= elapsed.Nanoseconds() {
		d = time.Duration(elapsed.Nanoseconds() - prev)
	} else {
		err = errors.New("RTT is negative")
	}
	return
}
