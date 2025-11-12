package handlers

import (
	"context"
	"fmt"
	"io"
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

type Handler struct {
	sessions         map[string]*s.Session
	templates        templates.TemplateCache
	cfg              *config.Config
	ctx              context.Context
	log              *slog.Logger
	containerFactory s.ContainerFactory
}

func NewHandler(
	tc templates.TemplateCache,
	cfg *config.Config,
	cf s.ContainerFactory,
	ctx context.Context,
	log *slog.Logger,
) *Handler {
	return &Handler{
		sessions:         make(map[string]*s.Session),
		templates:        tc,
		cfg:              cfg,
		ctx:              ctx,
		log:              log,
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
		h.log.Error(err.Error())
	}
}

func (h *Handler) HandleStartSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowd", http.StatusMethodNotAllowed)
		h.log.Error("Wrong Method: handleStartSession")
	}
	data, err := respond.Decode[request.CreateSessionRequest](r)
	if err != nil {
		http.Error(w, "bad json:"+err.Error(), http.StatusBadRequest)
		h.log.Error(err.Error())
	}

	sess, err := s.New(h.cfg.Container.Workdir, data.Repo, h.containerFactory, h.log)
	if err != nil {
		http.Error(w, "Failed to create Session:"+err.Error(), http.StatusBadRequest)
		h.log.Error(err.Error())
		return
	}
	h.sessions[sess.Token] = sess
	resp := map[string]string{
		"endpoint": "/sessions/" + sess.Token,
	}
	if err := respond.Encode(w, int(http.StatusOK), resp); err != nil {
		h.log.Error(err.Error())
	}
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowd", http.StatusMethodNotAllowed)
		h.log.Error("Wrong Method: handleIndex")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	temp := h.templates["index"]
	if temp == nil {
		h.handleError(w, "Inner server error", fmt.Errorf("Unknown template"), http.StatusServiceUnavailable)
		return
	}
	if err := temp.Execute(w, struct {
		Title string
	}{Title: "NvimAnywhere"}); err != nil {
		h.handleError(w, "Inner server error", fmt.Errorf("Failed to make a response: err"), http.StatusServiceUnavailable)
		return
	}
}

func (h *Handler) HandleSessionPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowd", http.StatusMethodNotAllowed)
		h.log.Error("Wrong Method: handleSessionPage")
		return
	}

	token, ok := tokenFromPath(r)
	if !ok {
		http.Error(w, "No token found", http.StatusBadRequest)
		h.log.Error("No token found: token", "token", token)
		return
	}
	sess := h.sessions[token]
	if sess == nil {
		http.Error(w, "No session found", http.StatusNotFound)
		h.log.Error("No session found: token", "token", token)
		return
	}
	temp := h.templates["session"]

	if temp == nil {
		http.Error(w, "Inner server error", http.StatusServiceUnavailable)
		h.log.Error("Unknown template")
		return
	}
	if err := temp.Execute(w, struct {
		Title string
		Token string
	}{Title: "NvimAnywhere", Token: token}); err != nil {
		http.Error(w, "Inner server error", http.StatusServiceUnavailable)
		h.log.Error("Failed to make a response: err", "err", err)
		return
	}

}

func (h *Handler) HandlePTY(w http.ResponseWriter, r *http.Request) {
	var errCh = make(chan error)
	var done = make(chan struct{})
	token := r.URL.Query().Get("token")
	sess := h.sessions[token]
	if sess == nil {
		h.handleError(w, "invalid/expired token", nil, http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		h.handleError(w, "Failed to create websocket conn", err, http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithCancel(h.ctx)
	defer cancel()

	onClose := closeAll(ctx, conn, sess, h.log)

	ptyR, ptyW, wait, err := startSession(ctx, sess)
	if err != nil {
		h.handleError(w, "", err, http.StatusInternalServerError)
		return
	}

	bridge := wsio.NewBridge(conn, ptyR, ptyW, wait, onResize(ctx, sess))

	if err := bridge.Start(ctx); err != nil {
		h.handleError(w, "", err, http.StatusInternalServerError)
		return

	}

	select {
	case <-h.ctx.Done():
		onClose("Shutdown")
		return
	case <-done:
		onClose("Finishing")
		return
	case e := <-errCh:
		onClose(e.Error())
		return
	}
}

func sendControl(ctx context.Context, ws *websocket.Conn, typ, msg string) {
	_ = ws.Write(ctx, websocket.MessageText, []byte(`{"type":"`+typ+`","reason":"`+msg+`"}`))
}

func closeAll(ctx context.Context, ws *websocket.Conn, sess *sessions.Session, log *slog.Logger) func(string) {
	return func(reason string) {
		sendControl(ctx, ws, "exit", reason)
		_ = ws.Close(websocket.StatusNormalClosure, reason)

		if err := sess.Close(); err != nil {
			log.Error(err.Error())
		}
		if reason != "" {
			log.Error(reason)
		}

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

func (h *Handler) handleError(w http.ResponseWriter, reason string, err error, status int) {
	http.Error(w, reason, status)
	if err != nil {
		h.log.Error(reason + err.Error())

	} else {
		h.log.Error(reason)

	}
}

func startSession(ctx context.Context, s *sessions.Session) (io.Reader, io.Writer, func() error, error) {

	if err := s.Start(ctx); err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to start session: %w", err)
	}
	return s.Attach(ctx)
}
