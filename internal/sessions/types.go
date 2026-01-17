package sessions

import (
	"context"
	"errors"
	"nvimanywhere/internal/config"
	"sync"
	"time"
)

var (
	SessionIsNotFound = errors.New("Session is not found")
	SessionIsNotReady = errors.New("Session is not in ready state")
	SessionIsFailed   = errors.New("Session is failed")
	SessionIsClosed   = errors.New("Session is closed")
)

type Session struct {
	ctx    context.Context
	cancel context.CancelFunc

	createdAt time.Time
	repoUrl   string
	cfg       *config.SessionRuntime
	rootPath   string
	runtimeId string

	errOnce   sync.Once
	lastError error
}
