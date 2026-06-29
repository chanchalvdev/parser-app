package services

import (
	"context"
	"errors"
	"testing"
)

type fakeHealthRepository struct {
	pingErr error
}

func (f *fakeHealthRepository) Ping(_ context.Context) error {
	return f.pingErr
}

func TestHealthServiceReady(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		service := NewHealthService(&fakeHealthRepository{})
		if err := service.Ready(context.Background()); err != nil {
			t.Fatalf("expected ready, got error: %v", err)
		}
	})

	t.Run("not ready", func(t *testing.T) {
		repoErr := errors.New("database unavailable")
		service := NewHealthService(&fakeHealthRepository{pingErr: repoErr})
		if err := service.Ready(context.Background()); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

