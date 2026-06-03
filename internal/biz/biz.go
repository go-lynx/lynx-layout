// Package biz contains the business-logic (use-case) layer of the user service.
// It defines domain interfaces and orchestrates data access via the LoginRepo interface.
package biz

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set used for dependency injection.
// This set contains the provider functions required to create LoginUseCase instances.
// When using Google Wire for dependency injection, this set can be used to automatically assemble required dependencies.
var ProviderSet = wire.NewSet(NewLoginUseCase)
