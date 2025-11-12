package wsio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/coder/websocket"
	"golang.org/x/sync/errgroup"
)

type size struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

type Bridge struct {
	conn    *websocket.Conn
	r       io.Reader
	w       io.Writer
	out     chan []byte
	setSize func(rows, cols int) error
	wait    func() error
	opts    *BridgeOption
	timer   *time.Timer
	start   time.Time
	counter int
}

type BridgeOption struct {
	MaxFrameBytes    int
	MaxCoalesceWait  time.Duration
	OutQueueCapacity int
}

func NewBridge(
	conn *websocket.Conn,
	r io.Reader,
	w io.Writer,
	wait func() error,
	onResize func(rows, cols int) error,
) *Bridge {
	opts := &BridgeOption{
		MaxFrameBytes:    1024 * 4,
		MaxCoalesceWait:  10 * time.Millisecond,
		OutQueueCapacity: 1024,
	}
	return &Bridge{
		conn:    conn,
		r:       r,
		w:       w,
		wait:    wait,
		opts:    opts,
		setSize: onResize,
		timer:   time.NewTimer(time.Millisecond * 10),
		out:     make(chan []byte, opts.OutQueueCapacity),
		start:   time.Now(),
		counter: 0,
	}
}

func (b *Bridge) Start(ctx context.Context) error {
	grp, gctx := errgroup.WithContext(ctx)
	grp.Go(func() error { return b.pumpWSToPTY(gctx) })
	grp.Go(func() error { return b.pumpPTYToQueue(gctx) })
	grp.Go(func() error { return b.pumpQueueToWS(gctx) })
	grp.Go(func() error { return b.pumpWaitExit(gctx) })

	err := grp.Wait()

	b.conn.Close(websocket.StatusNormalClosure, "closing")

	switch {
	case errors.Is(err, io.EOF):
		return nil
	default:
		return err
	}

}

func (b *Bridge) pumpWaitExit(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := b.wait(); err != nil {
				return err
			}
		}
	}
}

func (b *Bridge) pumpWSToPTY(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			t, data, err := b.conn.Read(ctx)
			if err != nil {
				return fmt.Errorf("Read from WS: %v", err)
			}
			b.handlerWSMessage(t, data)
		}
	}
}

func (b *Bridge) handlerWSMessage(t websocket.MessageType, data []byte) error {
	switch t {
	case websocket.MessageBinary:
		if _, err := b.w.Write(data); err != nil {
			return fmt.Errorf("pty write failed")
		}

	case websocket.MessageText:
		var m size
		if json.Unmarshal(data, &m) == nil && m.Cols > 0 && m.Rows > 0 {
			if err := b.setSize(m.Cols, m.Rows); err != nil {
				return fmt.Errorf("resize failed: ")
			}
		}
	}
	return nil
}

func (b *Bridge) pumpPTYToQueue(ctx context.Context) error {
	data := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			n, err := b.r.Read(data)
			if err != nil {
				return fmt.Errorf("Failed to read data from container: %v", err)
			}
			if n != 0 {
				chunk := append([]byte(nil), data[:n]...)
				b.out <- chunk
			}
		}
	}
}

func (b *Bridge) pumpQueueToWS(ctx context.Context) error {
	payload := make([]byte, 0, 32*1024)

	flush := func() error {
		b.timer.Reset(time.Millisecond * 10)
		if len(payload) == 0 {
			return nil
		}
		b.counter++
		t := time.Since(b.start)
		b.start = time.Now()
		log.Printf("After: %v send pocket: %d, IN queue: %d  Totla count: %d \n", t, len(payload), len(b.out), b.counter)
		err := b.conn.Write(ctx, websocket.MessageBinary, payload)
		payload = payload[:0]
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case chunk, ok := <-b.out:
			if !ok {
				return flush()
			}
			for len(chunk) > 0 {
				space := b.opts.MaxFrameBytes - len(payload)
				if space == 0 {
					if err := flush(); err != nil {
						return err
					}
					space = b.opts.MaxFrameBytes
				}
				take := space
				take = min(len(chunk), take)
				payload = append(payload, chunk[:take]...)
				chunk = chunk[take:]
			}
			if len(payload) >= b.opts.MaxFrameBytes {
				if err := flush(); err != nil {
					return err
				}
			}
		case <-b.timer.C:
			if err := flush(); err != nil {
				return err
			}
		}
	}
}
