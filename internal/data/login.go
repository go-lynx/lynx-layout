package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"github.com/go-lynx/lynx-layout/internal/code"
	"github.com/go-lynx/lynx-layout/internal/data/ent/user"
	"time"
)

type loginRepo struct {
	data *Data
	log  *log.Helper
}

func NewLoginRepo(data *Data, logger log.Logger) biz.LoginRepo {
	return &loginRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r loginRepo) FindUserByAccount(ctx context.Context, account string) (*bo.UserBO, error) {
	u, err := r.data.db.User.
		Query().
		Where(user.AccountEQ(account)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
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

func (r loginRepo) UpdateUserLastLoginTime(ctx context.Context, bo *bo.UserBO) error {
	rows, err := r.data.db.User.
		Update().
		SetLastLoginAt(time.Now()).
		Where(user.IDEQ(bo.Id)).
		Save(ctx)
	if rows != 1 {
		return code.LoginError
	}
	if err != nil {
		return err
	}
	return nil
}

func (r loginRepo) LoginAuth(ctx context.Context, bo *bo.UserBO) (string, error) {
	// TODO Remote invocation of other microservices via gRPC
	return "", nil
}
