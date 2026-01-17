package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// ============================================================
// Session Endpoint Dispatcher
// ------------------------------------------------------------
// HandleSession serves a single logical endpoint:
//
//   GET /sessions/{token}
//
// It dispatches requests based on intent:
//
//   • Regular HTTP requests render the session UI
//   • WebSocket upgrade requests attach to the session PTY
//
// WebSocket intent is detected via headers only.
// Full protocol validation is performed by websocket.Accept.
// ============================================================

func (app *App) HandleSession(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		app.handleSession(w, r)
	} else {
		app.sessionUIHandler(w, r)
	}
}

// ============================================================
// Session UI Handler
// ------------------------------------------------------------
// sessionUIHandler renders the HTML UI for an existing session.
//
// This handler does NOT validate session existence.
// ============================================================

func (app *App) sessionUIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.respondError(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	tmpl := app.templates["session"]
	if tmpl == nil {
		app.respondError(w, http.StatusServiceUnavailable, "unknown template", nil)
		return
	}
	if err := tmpl.Execute(w, struct {
		Title string
	}{
		Title: "NvimAnywhere",
	}); err != nil {
		app.respondError(w, http.StatusInternalServerError, "failed to render session page", err)
	}
}

// ============================================================
// PTY WebSocket Handler
// ------------------------------------------------------------
// handleSession upgrades the connection to WebSocket and attaches
// the client to an already-running session.
//
// Lifecycle:
//   1. Parse token
//   2. Accept WebSocket upgrade
//   3. Attach session (ownership boundary)
//   4. Wait for disconnect or server shutdown
//   5. Cleanup session + connection
//
// IMPORTANT:
//   After websocket.Accept succeeds, HTTP is no longer valid.
//   All errors must be handled via WebSocket only.
// ============================================================

func (app *App) handleSession(w http.ResponseWriter, r *http.Request) {
	token, ok := tokenFromPath(r)
	if !ok {
		app.respondError(w, http.StatusBadRequest, "token is not provided", nil)
		return
	}
	conn, err := app.upgrader.Upgrade(w, r, nil)
	if err != nil {
		app.log.Error(err.Error())
		return
	}
	defer conn.Close()

	app.mu.Lock()
	sess := app.sessions[token]
	if ok {
		delete(app.sessions, token) // prevent reuse
	}
	app.mu.Unlock()
	if !ok || sess == nil {
		app.respondError(w, http.StatusNotFound, "session not found", nil)
		return
	}
	defer func() {
		if err := sess.Close(); err != nil {
			app.log.Error(err.Error())
		}
	}()

	if err := sess.Attach(conn); err != nil {
		app.log.Error(err.Error())
		sendControl(r.Context(), conn, "session failed", "Fail within session's attach")
		return
	}
}

// ============================================================
// WebSocket Intent Detection
// ------------------------------------------------------------
// isWebSocketIntent performs a cheap pre-upgrade heuristic to
// detect whether the client intends to open a WebSocket.
//
// This is NOT protocol validation.
// ============================================================

func isWebSocketIntent(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// ============================================================
// Helpers
// ============================================================

func tokenFromPath(r *http.Request) (string, bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 {
		return "", false
	}
	return parts[len(parts)-1], true
}

func sendControl(ctx context.Context, ws *websocket.Conn, typ, msg string) {
	_ = ws.WriteMessage(websocket.TextMessage,
		[]byte(`{"type":"`+typ+`","reason":"`+msg+`"}`),
	)
}
