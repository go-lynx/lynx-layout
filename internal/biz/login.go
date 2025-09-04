package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-lynx/lynx-layout/api/login/code"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	"github.com/go-lynx/lynx/app/utils/auth"
)

// LoginUseCase defines user login related use cases, responsible for coordinating login business logic.
type LoginUseCase struct {
	repo LoginRepo // Repository interface that implements login-related data operations
}

// NewLoginUseCase creates a new LoginUseCase instance.
// Parameter repo is a repository instance that implements the LoginRepo interface, logger is the logger.
// Returns a pointer to a LoginUseCase instance.
func NewLoginUseCase(repo LoginRepo) *LoginUseCase {
	return &LoginUseCase{
		repo: repo,
	}
}

// LoginRepo defines the data operation interface required for login business.
type LoginRepo interface {
	// FindUserByAccount finds user information based on user account.
	// Parameters: ctx is the context, account is the user account.
	// Returns user business object pointer and any possible errors.
	FindUserByAccount(context.Context, string) (*bo.UserBO, error)
	// UpdateUserLastLoginTime updates the user's last login time.
	// Parameters: ctx is the context, user is the user business object.
	// Returns any possible errors.
	UpdateUserLastLoginTime(context.Context, *bo.UserBO) error
	// LoginAuth performs user login authentication and generates authentication token.
	// Parameters: ctx is the context, user is the user business object.
	// Returns authentication token string and any possible errors.
	LoginAuth(context.Context, *bo.UserBO) (string, error)
}

// UserLogin handles user login logic.
// Parameters: ctx is the context, bo is the business object containing user login information.
// Returns a business object pointer containing user information and authentication token, and any possible errors.
func (uc *LoginUseCase) UserLogin(ctx context.Context, bo *bo.UserBO) (*bo.UserBO, error) {
	// Find user information based on user account
	u, err := uc.repo.FindUserByAccount(ctx, bo.Account)
	var notFoundError *ent.NotFoundError
	// Check if user is not found
	if err != nil && errors.As(err, &notFoundError) {
		return nil, code.UserDoesNotExist
	}
	if err != nil {
		return nil, err
	}
	// Verify if the user's input password is correct
	if !auth.CheckPassword(u.Password, bo.Password) {
		return nil, code.IncorrectPassword
	}
	// Update user's last login time
	err = uc.repo.UpdateUserLastLoginTime(ctx, u)
	if err != nil {
		return nil, err
	}
	// Perform login authentication and get authentication token
	auth, err := uc.repo.LoginAuth(ctx, u)
	if err != nil {
		return nil, err
	}
	// Assign authentication token to user business object
	u.Token = auth
	return u, nil
}
