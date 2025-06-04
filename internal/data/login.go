package data

import (
	"context"
	"github.com/go-lynx/lynx-layout/api/login/code"
	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"github.com/go-lynx/lynx-layout/internal/data/ent/user"
	"time"
)

// loginRepo 实现了 biz.LoginRepo 接口，用于处理用户登录相关的数据操作。
// 包含数据访问实例和日志辅助工具。
type loginRepo struct {
	data *Data // 数据访问实例，包含数据库和 Redis 客户端
}

// NewLoginRepo 创建一个新的 loginRepo 实例，实现 biz.LoginRepo 接口。
// 参数 data 是数据访问实例，logger 是日志记录器。
// 返回实现了 biz.LoginRepo 接口的指针。
func NewLoginRepo(data *Data) biz.LoginRepo {
	return &loginRepo{
		data: data,
	}
}

// FindUserByAccount 根据用户账号查找用户信息。
// 参数 ctx 是上下文，用于控制请求的生命周期；account 是用户的登录账号。
// 返回用户业务对象指针和可能出现的错误。
func (r *loginRepo) FindUserByAccount(ctx context.Context, account string) (*bo.UserBO, error) {
	// 使用 ent 客户端查询数据库，根据账号查找唯一用户
	u, err := r.data.db.User.
		Query().
		Where(user.AccountEQ(account)).
		Only(ctx)
	if err != nil {
		// 查询失败，返回 nil 和错误信息
		return nil, err
	}
	// 将数据库查询结果转换为业务对象
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

// UpdateUserLastLoginTime 更新用户的最后登录时间。
// 参数 ctx 是上下文，用于控制请求的生命周期；bo 是用户业务对象。
// 返回可能出现的错误。
func (r *loginRepo) UpdateUserLastLoginTime(ctx context.Context, bo *bo.UserBO) error {
	// 使用 ent 客户端更新数据库中用户的最后登录时间
	rows, err := r.data.db.User.
		Update().
		SetLastLoginAt(time.Now()).
		Where(user.IDEQ(bo.Id)).
		Save(ctx)
	if rows != 1 {
		// 未成功更新一行记录，返回登录错误
		return code.LoginError
	}
	if err != nil {
		// 更新过程中出现错误，返回错误信息
		return err
	}
	// 更新成功，返回 nil
	return nil
}

// LoginAuth 进行用户登录认证并生成认证令牌。
// 目前该方法为 TODO 状态，计划通过 gRPC 远程调用其他微服务实现。
// 参数 ctx 是上下文，用于控制请求的生命周期；bo 是用户业务对象。
// 返回认证令牌字符串和可能出现的错误。
func (r *loginRepo) LoginAuth(ctx context.Context, bo *bo.UserBO) (string, error) {
	// TODO Remote invocation of other microservices via gRPC
	return "", nil
}
