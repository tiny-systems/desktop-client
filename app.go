package main

import (
	"context"
	"github.com/go-logr/logr"
)

// App struct
type App struct {
	ctx    context.Context
	logger logr.Logger
}

// NewApp creates a new App application struct
func NewApp(l logr.Logger) *App {
	return &App{
		logger: l,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

}

func (a *App) shutdown(ctx context.Context) {
}
