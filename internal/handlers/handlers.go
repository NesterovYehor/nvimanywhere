package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"nvimanywhere/internal/config"
	"nvimanywhere/internal/request"
	"nvimanywhere/internal/respond"
	"nvimanywhere/internal/sessions"
	s "nvimanywhere/internal/sessions"
	"nvimanywhere/internal/templates"
	"nvimanywhere/internal/wsio"
	"strings"
	"time"

	"github.com/coder/websocket"
)

var done = make(chan struct{})
var errCh = make(chan error)

type Handler struct {
	sessions         map[string]*s.Session
	templates        templates.TemplateCache
	cfg              *config.Config
	containerFactory s.ContainerFactory
}

func NewHandler(tc templates.TemplateCache, cfg *config.Config, cf s.ContainerFactory) *Handler {
	return &Handler{
		sessions:         make(map[string]*s.Session),
		templates:        tc,
		cfg:              cfg,
		containerFactory: cf,
	}
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte("ok")); err != nil {
		slog.Error(err.Error())
	}
}

func (h *Handler) HandleStartSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowd", http.StatusMethodNotAllowed)
		slog.Error("Wrong Method: handleStartSession")
	}
	data, err := respond.Decode[request.CreateSessionRequest](r)
	if err != nil {
		http.Error(w, "bad json:"+err.Error(), http.StatusBadRequest)
		slog.Error(err.Error())
	}

	sess, err := s.New(h.cfg.BaseDir, data.Repo, h.containerFactory)
	if err != nil {
		http.Error(w, "Failed to create Session:"+err.Error(), http.StatusBadRequest)
		slog.Error(err.Error())
		return
	}
	h.sessions[sess.Token] = sess
	resp := map[string]string{
		"url": "http://localhost:8080/sessions/" + sess.Token,
	}
	if err := respond.Encode(w, int(http.StatusOK), resp); err != nil {
		slog.Error(err.Error())
	}
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowd", http.StatusMethodNotAllowed)
		slog.Error("Wrong Method: handleIndex")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	temp := h.templates["index"]
	if temp == nil {
		http.Error(w, "Inner server error", http.StatusServiceUnavailable)
		slog.Error("Unknown template")
		return
	}
	if err := temp.Execute(w, struct {
		Title string
	}{Title: "NvimAnywhere"}); err != nil {
		http.Error(w, "Inner server error", http.StatusServiceUnavailable)
		slog.Error("Failed to make a response: err", "err", err)
		return
	}
}

func (h *Handler) HandleSessionPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowd", http.StatusMethodNotAllowed)
		slog.Error("Wrong Method: handleSessionPage")
		return
	}

	token, ok := tokenFromPath(r)
	if !ok {
		http.Error(w, "No token found", http.StatusBadRequest)
		slog.Error("No session found: token", "token", token)
		return
	}
	sess := h.sessions[token]
	if sess == nil {
		http.Error(w, "No session found", http.StatusNotFound)
		slog.Error("No session found")
		return
	}
	temp := h.templates["session"]

	if temp == nil {
		http.Error(w, "Inner server error", http.StatusServiceUnavailable)
		slog.Error("Unknown template")
		return
	}
	if err := temp.Execute(w, struct {
		Title string
		Token string
	}{Title: "NvimAnywhere", Token: token}); err != nil {
		http.Error(w, "Inner server error", http.StatusServiceUnavailable)
		slog.Error("Failed to make a response: err", "err", err)
		return
	}

}

func (h *Handler) HandlePTY(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	sess := h.sessions[token]
	if sess == nil {
		handleError(w, "invalid/expired token", nil, http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		handleError(w, "Failed to create websocket conn", err, http.StatusInternalServerError)
		return
	}
	slog.Info("Conn to WS created")
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if err := sess.Start(ctx); err != nil {
		handleError(w, "Failed to start session", err, http.StatusServiceUnavailable)

	}
	onClose := closeAll(ctx, conn, sess)

	ptyR, ptyW, wait, err := sess.Attach(ctx)
	if err != nil {
		onClose("attache failed: " + err.Error())
		return
	}
	bridge := wsio.NewBridge(conn, ptyR, ptyW, onResize(ctx, sess))

	go bridge.PTYToWS(ctx)

	go bridge.WSToPTY(ctx)

	go bridge.WatchWait(wait)
	select {
	case <-done:
		onClose("Finishing")
		return
	case <-errCh:
		err = <-errCh
		handleError(w, "", err, http.StatusInternalServerError)
		onClose(err.Error())
		return
	}
}

func sendControl(ctx context.Context, ws *websocket.Conn, typ, msg string) {
	_ = ws.Write(ctx, websocket.MessageText, []byte(`{"type":"`+typ+`","reason":"`+msg+`"}`))
}
func closeAll(ctx context.Context, ws *websocket.Conn, sess *sessions.Session) func(string) {
	return func(reason string) {
		sendControl(ctx, ws, "exit", reason)
		_ = ws.Close(websocket.StatusNormalClosure, reason)

		if err := sess.Close(ctx); err != nil {
			slog.Error(err.Error())

		}
		slog.Error(reason)
		close(done)

	}
}

func tokenFromPath(r *http.Request) (string, bool) {
	const prefix = "/sessions/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		return "", false
	}
	tok := strings.TrimPrefix(r.URL.Path, prefix)
	tok = strings.TrimSuffix(tok, "/")
	if tok == "" || strings.Contains(tok, "/") {
		return "", false
	}
	return tok, true
}

func onResize(ctx context.Context, s *sessions.Session) func(int, int) error {
	return func(cols, rows int) error {
		if cols <= 0 || rows <= 0 {
			return nil // ignore bogus sizes
		}
		const maxCols, maxRows = 500, 200
		if cols > maxCols {
			cols = maxCols
		}
		if rows > maxRows {
			rows = maxRows
		}

		rc, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return s.ResizePTY(rc, cols, rows)
	}
}

func handleError(w http.ResponseWriter, reason string, err error, status int) {
	http.Error(w, reason, status)
	if err != nil {
		slog.Error(reason + err.Error())

	} else {
		slog.Error(reason)

	}
}
