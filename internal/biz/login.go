package biz

import (
	"context"
	"errors"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"github.com/go-lynx/lynx-layout/internal/code"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	"github.com/go-lynx/lynx/app/util"
)

// LoginUseCase 定义了用户登录相关的用例，负责协调登录业务逻辑。
type LoginUseCase struct {
	repo LoginRepo // 实现登录相关数据操作的仓库接口
}

// NewLoginUseCase 创建一个新的 LoginUseCase 实例。
// 参数 repo 是实现 LoginRepo 接口的仓库实例，logger 是日志记录器。
// 返回一个指向 LoginUseCase 实例的指针。
func NewLoginUseCase(repo LoginRepo) *LoginUseCase {
	return &LoginUseCase{
		repo: repo,
	}
}

// LoginRepo 定义了登录业务所需的数据操作接口。
type LoginRepo interface {
	// FindUserByAccount 根据用户账号查找用户信息。
	// 参数 ctx 是上下文，account 是用户账号。
	// 返回用户业务对象指针和可能的错误。
	FindUserByAccount(context.Context, string) (*bo.UserBO, error)
	// UpdateUserLastLoginTime 更新用户的最后登录时间。
	// 参数 ctx 是上下文，user 是用户业务对象。
	// 返回可能的错误。
	UpdateUserLastLoginTime(context.Context, *bo.UserBO) error
	// LoginAuth 进行用户登录认证并生成认证令牌。
	// 参数 ctx 是上下文，user 是用户业务对象。
	// 返回认证令牌字符串和可能的错误。
	LoginAuth(context.Context, *bo.UserBO) (string, error)
}

// UserLogin 处理用户登录逻辑。
// 参数 ctx 是上下文，bo 是包含用户登录信息的业务对象。
// 返回包含用户信息和认证令牌的业务对象指针，以及可能的错误。
func (uc *LoginUseCase) UserLogin(ctx context.Context, bo *bo.UserBO) (*bo.UserBO, error) {
	// 根据用户账号查找用户信息
	u, err := uc.repo.FindUserByAccount(ctx, bo.Account)
	var notFoundError *ent.NotFoundError
	// 检查是否未找到用户
	if err != nil && errors.As(err, &notFoundError) {
		return nil, code.UserDoesNotExist
	}
	// 验证用户输入的密码是否正确
	if !util.CheckCiphertext(bo.Password, u.Password) {
		return nil, code.IncorrectPassword
	}
	// 更新用户的最后登录时间
	err = uc.repo.UpdateUserLastLoginTime(ctx, u)
	if err != nil {
		return nil, err
	}
	// 进行登录认证并获取认证令牌
	auth, err := uc.repo.LoginAuth(ctx, u)
	if err != nil {
		return nil, err
	}
	// 将认证令牌赋值给用户业务对象
	u.Token = auth
	return u, nil
}
