package code

import (
	"github.com/go-kratos/kratos/v2/errors"
)

// 定义一系列与用户登录相关的错误变量
var (
	// UserDoesNotExist 表示用户账号不存在的错误。
	// 当系统根据用户提供的账号未找到对应记录时，会返回此错误。
	// 使用 login.ErrorReason_USER_DOES_NOT_EXIST 作为错误原因，错误消息为“该账号不存在”。
	UserDoesNotExist = errors.New(1404, "user_does_not_exist", "user_does_not_exist")
	// IncorrectPassword 表示用户输入的密码错误的错误。
	// 当用户输入的密码与系统中存储的密码不匹配时，会返回此错误。
	// 使用 login.ErrorReason_INCORRECT_PASSWORD 作为错误原因，错误消息为“密码错误”。
	IncorrectPassword = errors.New(1405, "incorrect_password", "incorrect_password")
	// LoginError 表示用户登录过程中出现异常的错误。
	// 当登录过程中发生其他未预期的错误时，会返回此错误。
	// 使用 login.ErrorReason_LOGIN_ERROR 作为错误原因，错误消息为“登陆异常”。
	LoginError = errors.New(1406, "login_error", "login_error")
)
