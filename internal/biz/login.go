package biz

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"github.com/go-lynx/lynx-layout/internal/code"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	"github.com/go-lynx/lynx/util"
)

type LoginUseCase struct {
	repo LoginRepo
	log  *log.Helper
}

func NewLoginUseCase(repo LoginRepo, logger log.Logger) *LoginUseCase {
	return &LoginUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

type LoginRepo interface {
	FindUserByAccount(context.Context, string) (*bo.UserBO, error)
	UpdateUserLastLoginTime(context.Context, *bo.UserBO) error
	LoginAuth(context.Context, *bo.UserBO) (string, error)
}

func (uc *LoginUseCase) UserLogin(ctx context.Context, bo *bo.UserBO) (*bo.UserBO, error) {
	u, err := uc.repo.FindUserByAccount(ctx, bo.Account)
	var notFoundError *ent.NotFoundError
	if err != nil && errors.As(err, &notFoundError) {
		return nil, code.UserDoesNotExist
	}
	if !util.CheckCiphertext(bo.Password, u.Password) {
		return nil, code.IncorrectPassword
	}
	err = uc.repo.UpdateUserLastLoginTime(ctx, u)
	if err != nil {
		return nil, err
	}
	auth, err := uc.repo.LoginAuth(ctx, u)
	if err != nil {
		return nil, err
	}
	u.Token = auth
	return u, nil
}
