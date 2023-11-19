package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/bo"

	v1 "github.com/go-lynx/lynx-layout/api/login/v1"
)

type LoginService struct {
	v1.UnimplementedLoginServer
	log *log.Helper
	uc  *biz.LoginUseCase
}

func NewLoginService(uc *biz.LoginUseCase) *LoginService {
	return &LoginService{uc: uc}
}

func (svc *LoginService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	u, err := svc.uc.UserLogin(ctx, &bo.UserBO{
		Account:  req.Account,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		Token: u.Token,
		User: &v1.UserInfo{
			Account:  u.Account,
			Num:      u.Num,
			NickName: u.Nickname,
			Avatar:   u.Avatar,
		},
	}, nil
}
