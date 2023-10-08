package main

import (
	"testing"

	"github.com/chamrasilva89/reservationWeb/internal/config"
	"github.com/go-chi/chi"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)
	switch v := mux.(type) {
	case *chi.Mux:
	default:
		t.Errorf("Expected a chi Mux, got %T", v)
	}
}
