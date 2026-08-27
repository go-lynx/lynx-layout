package data

import (
	"context"
	"errors"
	"testing"

	"github.com/go-lynx/lynx-layout/api/login/code"
	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	"github.com/go-lynx/lynx/pkg/auth"
)

func TestNewData_FallsBackToMemoryWithoutEntProvider(t *testing.T) {
	d, err := NewData(nil, nil)
	if err != nil {
		t.Fatalf("expected in-memory fallback, got error: %v", err)
	}
	if !d.InMemory() {
		t.Fatal("expected data layer to report in-memory mode")
	}
	if _, err := d.entClient(); err == nil {
		t.Fatal("expected entClient to fail in memory mode")
	}

	repo := NewLoginRepo(d, nil)
	if _, ok := repo.(*memoryLoginRepo); !ok {
		t.Fatalf("expected memory login repo, got %T", repo)
	}
}

func TestMemoryLoginRepo_LoginFlowAndNotFound(t *testing.T) {
	d, err := NewData(nil, nil)
	if err != nil {
		t.Fatalf("NewData: %v", err)
	}
	repo := NewLoginRepo(d, nil)
	ctx := context.Background()

	_, err = repo.FindUserByAccount(ctx, "nobody")
	var notFound *ent.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected ent.NotFoundError for unknown account, got %v", err)
	}

	u, err := repo.FindUserByAccount(ctx, memoryDemoAccount)
	if err != nil {
		t.Fatalf("expected seeded demo user: %v", err)
	}
	if !auth.CheckPassword(u.Password, memoryDemoPassword) {
		t.Fatal("seeded demo password hash does not verify")
	}
	if err := repo.UpdateUserLastLoginTime(ctx, u); err != nil {
		t.Fatalf("UpdateUserLastLoginTime: %v", err)
	}
	if err := repo.UpdateUserLastLoginTime(ctx, &bo.UserBO{Id: 9999}); !errors.Is(err, code.LoginError) {
		t.Fatalf("expected LoginError for unknown id, got %v", err)
	}

	token, err := repo.LoginAuth(ctx, u)
	if err != nil {
		t.Fatalf("LoginAuth: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if id, ok := d.memory.lookupToken(token); !ok || id != u.Id {
		t.Fatalf("token not stored for user %d (ok=%v id=%d)", u.Id, ok, id)
	}

	// Full use-case round trip through biz to make sure the not-found mapping still works.
	uc := biz.NewLoginUseCase(repo)
	if _, err := uc.UserLogin(ctx, &bo.UserBO{Account: "nobody", Password: "x"}); !errors.Is(err, code.UserDoesNotExist) {
		t.Fatalf("expected UserDoesNotExist, got %v", err)
	}
	if _, err := uc.UserLogin(ctx, &bo.UserBO{Account: memoryDemoAccount, Password: "wrong"}); !errors.Is(err, code.IncorrectPassword) {
		t.Fatalf("expected IncorrectPassword, got %v", err)
	}
	logged, err := uc.UserLogin(ctx, &bo.UserBO{Account: memoryDemoAccount, Password: memoryDemoPassword})
	if err != nil {
		t.Fatalf("expected successful login: %v", err)
	}
	if logged.Token == "" {
		t.Fatal("expected token on successful login")
	}
}
