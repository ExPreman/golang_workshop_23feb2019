package storage

import (
	"context"
	"testing"

	"github.com/win-t/golang_workshop_23feb2019/internal/config"
	"github.com/win-t/golang_workshop_23feb2019/internal/logger"
)

func TestSomething(t *testing.T) {
	ctx := context.Background()
	user := "testuser"
	pass := "testpass"

	s, err := Open(config.PG_URL(), logger.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.DoMigration(ctx); err != nil {
		t.Fatal(err)
	}

	if err := s.AddUser(ctx, user, pass); err != nil {
		t.Fatal(err)
	}

	counter, err := s.GetCounter(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if counter != 0 {
		t.Fatal("counter is not 0")
	}

	if err := s.IncCounter(ctx, user); err != nil {
		t.Fatal(err)
	}

	counter, err = s.GetCounter(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if counter != 1 {
		t.Fatal("counter is not 1")
	}

	if err := s.ResetCounter(ctx); err != nil {
		t.Fatal(err)
	}

	counter, err = s.GetCounter(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if counter != 0 {
		t.Fatal("counter is not 0")
	}
}
