package service

import (
	"context"
	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/bo"

	v1 "github.com/go-lynx/lynx-layout/api/login/v1"
)

// LoginService 实现了 v1.LoginServer 接口，用于处理用户登录相关的 RPC 请求。
// 该服务依赖业务逻辑层的 LoginUseCase 来完成具体的登录业务。
type LoginService struct {
	v1.UnimplementedLoginServer                   // 嵌入 UnimplementedLoginServer，自动实现接口的空方法
	uc                          *biz.LoginUseCase // 业务逻辑层的登录用例实例，负责处理登录的核心业务逻辑
}

// NewLoginService 创建一个新的 LoginService 实例。
// 参数 uc 是业务逻辑层的登录用例实例。
// 返回一个指向 LoginService 实例的指针。
func NewLoginService(uc *biz.LoginUseCase) *LoginService {
	return &LoginService{uc: uc}
}

// Login 处理用户登录的 RPC 请求。
// 参数 ctx 是上下文，用于控制请求的生命周期；req 是客户端发送的登录请求。
// 返回登录响应和可能出现的错误。
func (svc *LoginService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	// 调用业务逻辑层的 UserLogin 方法进行用户登录操作
	u, err := svc.uc.UserLogin(ctx, &bo.UserBO{
		Account:  req.Account,  // 从请求中获取用户账号
		Password: req.Password, // 从请求中获取用户密码
	})
	if err != nil {
		// 登录过程中出现错误，返回 nil 和错误信息
		return nil, err
	}
	// 登录成功，构造登录响应
	return &v1.LoginReply{
		Token: u.Token, // 将业务逻辑层返回的令牌添加到响应中
		User: &v1.UserInfo{
			Account:  u.Account,  // 将用户账号添加到用户信息中
			Num:      u.Num,      // 将用户编号添加到用户信息中
			NickName: u.Nickname, // 将用户昵称添加到用户信息中
			Avatar:   u.Avatar,   // 将用户头像 URL 添加到用户信息中
		},
	}, nil
}
