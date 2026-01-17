package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"nvimanywhere/internal/httpjson"
	"nvimanywhere/internal/sessions"
	"time"
)

func (app *App) HandleStartSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.respondError(w, http.StatusMethodNotAllowed, "Method not allowd", nil)
		return
	}
	type request struct {
		Repo string `json:"repo"`
	}

	data, err := httpjson.Decode[request](r)
	if err != nil {
		app.respondError(w, http.StatusBadRequest, "Failed to process request", err)
		return
	}
	token, err := createToken()
	if err != nil {
		app.respondError(w, 500, "Failed to create token", err)
		return
	}

	s, err := sessions.StartNewSession(app.ctx, app.cfg.SessionRuntime, data.Repo, token)
	if err != nil {
		app.respondError(w, 500, "Failed to creat session", err)
		return
	}
	app.mu.Lock()
	app.sessions[token] = s
	app.mu.Unlock()

	endpoint := "sessions/" + token
	if err := httpjson.Encode(w, 201, map[string]string{"endpoint": endpoint}); err != nil {
		app.respondError(w, 500, "Failed to respond", err)
		return
	}
	app.log.Info(fmt.Sprintf("Sesion Started at: %v", time.Now()))
}

func createToken() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
