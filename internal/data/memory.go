package data

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/go-lynx/lynx-layout/api/login/code"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	"github.com/go-lynx/lynx/log"
	"github.com/go-lynx/lynx/pkg/auth"
)

// Demo credentials seeded into the in-memory store so the login example is usable
// out of the box. They only exist when no MySQL plugin is loaded (local development).
const (
	memoryDemoAccount  = "admin"
	memoryDemoPassword = "admin123"
)

// memoryStore is a mutex-guarded, map-backed replacement for the MySQL user table and
// the Redis-backed token storage. It is selected at runtime when the mysql client plugin
// is not loaded (see NewData) and is meant for local development and smoke tests only.
type memoryStore struct {
	mu     sync.RWMutex
	nextID int64
	users  map[string]*memoryUser // keyed by account
	tokens map[string]int64       // token -> user id
}

type memoryUser struct {
	bo          bo.UserBO
	lastLoginAt time.Time
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users:  make(map[string]*memoryUser),
		tokens: make(map[string]int64),
	}
}

// newSeededMemoryStore creates a memory store containing the demo user.
func newSeededMemoryStore() (*memoryStore, error) {
	store := newMemoryStore()
	hash, err := auth.HashPassword(memoryDemoPassword, 0)
	if err != nil {
		return nil, fmt.Errorf("hash demo password: %w", err)
	}
	if _, err := store.addUser(bo.UserBO{
		Account:        memoryDemoAccount,
		Password:       hash,
		Nickname:       "Demo Admin",
		RegisterSource: 1,
		Stats:          1,
	}); err != nil {
		return nil, err
	}
	return store, nil
}

// addUser inserts a user and assigns id/num when they are empty.
func (s *memoryStore) addUser(u bo.UserBO) (*bo.UserBO, error) {
	if u.Account == "" {
		return nil, fmt.Errorf("account must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[u.Account]; exists {
		return nil, fmt.Errorf("account %q already exists", u.Account)
	}
	s.nextID++
	if u.Id == 0 {
		u.Id = s.nextID
	}
	if u.Num == "" {
		u.Num = fmt.Sprintf("U%06d", u.Id)
	}
	s.users[u.Account] = &memoryUser{bo: u}
	copied := u
	return &copied, nil
}

func (s *memoryStore) findByAccount(account string) (*bo.UserBO, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[account]
	if !ok {
		// Reuse ent's not-found error type so biz keeps a single mapping to code.UserDoesNotExist.
		return nil, &ent.NotFoundError{}
	}
	copied := u.bo
	return &copied, nil
}

func (s *memoryStore) touchLastLogin(id int64, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.bo.Id == id {
			u.lastLoginAt = at
			return nil
		}
	}
	return code.LoginError
}

func (s *memoryStore) issueToken(id int64) (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(raw)
	s.mu.Lock()
	s.tokens[token] = id
	s.mu.Unlock()
	return token, nil
}

// lookupToken resolves the user id behind a token issued by issueToken.
func (s *memoryStore) lookupToken(token string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.tokens[token]
	return id, ok
}

// memoryLoginRepo implements biz.LoginRepo on top of memoryStore.
type memoryLoginRepo struct {
	store *memoryStore
}

func newMemoryLoginRepo(store *memoryStore) *memoryLoginRepo {
	return &memoryLoginRepo{store: store}
}

func (r *memoryLoginRepo) FindUserByAccount(_ context.Context, account string) (*bo.UserBO, error) {
	return r.store.findByAccount(account)
}

func (r *memoryLoginRepo) UpdateUserLastLoginTime(_ context.Context, u *bo.UserBO) error {
	if u == nil {
		return code.LoginError
	}
	return r.store.touchLastLogin(u.Id, time.Now())
}

// LoginAuth issues a locally generated random token; the external gRPC auth service is
// only consulted when the service runs against the real data plugins.
func (r *memoryLoginRepo) LoginAuth(ctx context.Context, u *bo.UserBO) (string, error) {
	if u == nil {
		return "", fmt.Errorf("user must not be nil")
	}
	token, err := r.store.issueToken(u.Id)
	if err != nil {
		return "", err
	}
	log.DebugfCtx(ctx, "in-memory login token issued for account %s", u.Account)
	return token, nil
}
