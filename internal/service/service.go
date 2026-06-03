// Package service implements the application-service layer of the user service.
// It translates gRPC/HTTP requests into calls to the business-logic layer and
// handles cross-cutting concerns such as distributed login locking.
package service

import (
	"github.com/google/wire"
)

// ProviderSet is a Google Wire provider set used for dependency injection.
// This set contains the NewLoginService function, which means that when using Google Wire for dependency injection,
// LoginService instances can be automatically created. By adding this function to the set,
// it becomes convenient to manage and initialize LoginService related dependencies in the project.
var ProviderSet = wire.NewSet(NewLoginService)
