package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"github.com/go-lynx/lynx/log"

	v1 "github.com/go-lynx/lynx-layout/api/login/v1"
)

type LockRunner func(ctx context.Context, key string, expiration time.Duration, fn func() error) error

// LoginService implements the v1.LoginServer interface for handling user login related RPC requests.
// This service depends on the business logic layer's LoginUseCase to complete specific login business.
type LoginService struct {
	v1.UnimplementedLoginServer                   // Embed UnimplementedLoginServer to automatically implement empty methods of the interface
	uc                          *biz.LoginUseCase // Login use case instance from business logic layer, responsible for handling core login business logic
	lockRunner                  LockRunner
}

// NewLoginService creates a new LoginService instance.
// Parameter uc is the login use case instance from the business logic layer.
// Returns a pointer to a LoginService instance.
func NewLoginService(uc *biz.LoginUseCase, lockRunner LockRunner) *LoginService {
	if lockRunner == nil {
		lockRunner = func(_ context.Context, _ string, _ time.Duration, fn func() error) error {
			if fn == nil {
				return nil
			}
			return fn()
		}
	}
	return &LoginService{
		uc:         uc,
		lockRunner: lockRunner,
	}
}

// Login handles user login RPC requests.
// Parameters: ctx is the context for controlling the request lifecycle; req is the login request sent by the client.
// Returns login response and any possible errors.
func (svc *LoginService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	log.InfofCtx(ctx, "LoginService.Login log print test 1")

	if svc == nil || svc.uc == nil {
		return nil, fmt.Errorf("login service is not initialized")
	}
	if req == nil {
		return nil, fmt.Errorf("login request is nil")
	}

	var u *bo.UserBO
	err := svc.lockRunner(ctx, buildLoginLockKey(req.GetAccount()), 10*time.Second, func() error {
		log.InfofCtx(ctx, "LoginService.Login log print test")
		var callErr error
		u, callErr = svc.uc.UserLogin(ctx, &bo.UserBO{
			Account:  req.Account,  // Get user account from request
			Password: req.Password, // Get user password from request
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}

	// Login successful, construct login response
	log.InfofCtx(ctx, "LoginService.Login log print test 2")
	return &v1.LoginReply{
		Token: u.Token, // Add the token returned by the business logic layer to the response
		User: &v1.UserInfo{
			Account:  u.Account,  // Add user account to user information
			Num:      u.Num,      // Add user number to user information
			NickName: u.Nickname, // Add user nickname to user information
			Avatar:   u.Avatar,   // Add user avatar URL to user information
		},
	}, nil
}

func buildLoginLockKey(account string) string {
	account = strings.TrimSpace(account)
	if account == "" {
		return "login"
	}
	return "login:" + account
}
