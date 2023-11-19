package code

import (
	"github.com/go-kratos/kratos/v2/errors"
	login "github.com/go-lynx/lynx-layout/api/login/v1"
)

var (
	UserDoesNotExist  = errors.InternalServer(login.ErrorReason_USER_DOES_NOT_EXIST.String(), "该账号不存在")
	IncorrectPassword = errors.InternalServer(login.ErrorReason_INCORRECT_PASSWORD.String(), "密码错误")
	LoginError        = errors.InternalServer(login.ErrorReason_LOGIN_ERROR.String(), "登陆异常")
)
