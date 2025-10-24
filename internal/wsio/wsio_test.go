package wsio_test

import(
	"github.com/coder/websocket"
	"io"
	"nvimanywhere/internal/wsio"
	"testing"
)

func TestNewBridge(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		conn     websocket.Conn
		r        io.Reader
		w        io.Writer
		onResize func(rows int, cols int) error
		want     *wsio.Bridge
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wsio.NewBridge(tt.conn, tt.r, tt.w, tt.onResize)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewBridge() = %v, want %v", got, tt.want)
			}
		})
	}
}

