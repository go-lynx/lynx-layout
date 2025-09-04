package data

import (
	"context"
	"time"

	"github.com/go-lynx/lynx-layout/api/login/code"
	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"github.com/go-lynx/lynx-layout/internal/data/ent/user"
)

// loginRepo implements the biz.LoginRepo interface for handling user login related data operations.
// Contains data access instance and logging helper tools.
type loginRepo struct {
	data *Data // Data access instance containing database and Redis clients
}

// NewLoginRepo creates a new loginRepo instance that implements the biz.LoginRepo interface.
// Parameter data is the data access instance, logger is the logger.
// Returns a pointer that implements the biz.LoginRepo interface.
func NewLoginRepo(data *Data) biz.LoginRepo {
	return &loginRepo{
		data: data,
	}
}

// FindUserByAccount finds user information based on user account.
// Parameters: ctx is the context for controlling the request lifecycle; account is the user's login account.
// Returns user business object pointer and any possible errors.
func (r *loginRepo) FindUserByAccount(ctx context.Context, account string) (*bo.UserBO, error) {
	// Use ent client to query database and find unique user based on account
	u, err := r.data.db.User.
		Query().
		Where(user.AccountEQ(account)).
		Only(ctx)
	if err != nil {
		// Query failed, return nil and error information
		return nil, err
	}
	// Convert database query result to business object
	return &bo.UserBO{
		Id:       u.ID,
		Account:  u.Account,
		Password: u.Password,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Num:      u.Num,
		Stats:    u.Stats,
	}, nil
}

// UpdateUserLastLoginTime updates the user's last login time.
// Parameters: ctx is the context for controlling the request lifecycle; bo is the user business object.
// Returns any possible errors.
func (r *loginRepo) UpdateUserLastLoginTime(ctx context.Context, bo *bo.UserBO) error {
	// Use ent client to update user's last login time in database
	rows, err := r.data.db.User.
		Update().
		SetLastLoginAt(time.Now()).
		Where(user.IDEQ(bo.Id)).
		Save(ctx)
	if rows != 1 {
		// Failed to update one record, return login error
		return code.LoginError
	}
	if err != nil {
		// Error occurred during update process, return error information
		return err
	}
	// Update successful, return nil
	return nil
}

// LoginAuth performs user login authentication and generates authentication token.
// Currently this method is in TODO status, planned to be implemented via gRPC remote calls to other microservices.
// Parameters: ctx is the context for controlling the request lifecycle; bo is the user business object.
// Returns authentication token string and any possible errors.
func (r *loginRepo) LoginAuth(ctx context.Context, bo *bo.UserBO) (string, error) {
	// TODO Remote invocation of other microservices via gRPC
	return "", nil
}
