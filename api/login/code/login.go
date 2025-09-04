package code

import (
	"github.com/go-kratos/kratos/v2/errors"
)

// Define a series of error variables related to user login
var (
	// UserDoesNotExist represents the error when user account does not exist.
	// This error is returned when the system cannot find a corresponding record based on the account provided by the user.
	// Uses login.ErrorReason_USER_DOES_NOT_EXIST as the error reason, with error message "user_does_not_exist".
	UserDoesNotExist = errors.New(1404, "user_does_not_exist", "user_does_not_exist")
	// IncorrectPassword represents the error when user enters incorrect password.
	// This error is returned when the password entered by the user does not match the password stored in the system.
	// Uses login.ErrorReason_INCORRECT_PASSWORD as the error reason, with error message "incorrect_password".
	IncorrectPassword = errors.New(1405, "incorrect_password", "incorrect_password")
	// LoginError represents the error when an exception occurs during user login process.
	// This error is returned when other unexpected errors occur during the login process.
	// Uses login.ErrorReason_LOGIN_ERROR as the error reason, with error message "login_error".
	LoginError = errors.New(1406, "login_error", "login_error")
)
