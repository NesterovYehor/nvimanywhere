package wsio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/coder/websocket"
)

type Bridge struct {
	Conn    *websocket.Conn
	PTYR    io.Reader
	PTYW    io.Writer
	SetSize func(rows, cols int) error
}

func NewBridge(conn *websocket.Conn, r io.Reader, w io.Writer, onResize func(rows, cols int) error) *Bridge {
	return &Bridge{
		Conn:    conn,
		PTYR:    r,
		PTYW:    w,
		SetSize: onResize,
	}
}

func (b *Bridge) WSToPTY(ctx context.Context) error {
	for {
		t, data, err := b.Conn.Read(ctx)
		if err != nil {

			return fmt.Errorf("Read from WS: %v", err)
		}
		switch t {
		case websocket.MessageBinary:
			if _, err := b.PTYW.Write(data); err != nil {
				return fmt.Errorf("pty write failed")
			}

		case websocket.MessageText:
			var m struct {
				Type string `json:"type"`
				Cols int    `json:"cols"`
				Rows int    `json:"rows"`
			}
			if json.Unmarshal(data, &m) == nil && m.Cols > 0 && m.Rows > 0 {
				if err := b.SetSize(m.Cols, m.Rows); err != nil {
					return fmt.Errorf("resize failed: ")
				}
			}
		}
	}
}

func (b *Bridge) WatchWait(wait func() error) error {
	if err := wait(); err != nil {
		return err
	}
	return fmt.Errorf("process exited")
}

func (b *Bridge) PTYToWS(ctx context.Context) error {
	data := make([]byte, 32*1024)
	for {
		n, err := b.PTYR.Read(data)
		if err != nil {
			return fmt.Errorf("Failed to read data from container: %v", err)
		}
		if n != 0 {
			if err := b.Conn.Write(ctx, websocket.MessageBinary, data[:n]); err != nil {
				return fmt.Errorf("Failed to write to WS: %v", err)
			}
		}
	}
}
