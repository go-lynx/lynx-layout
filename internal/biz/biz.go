package biz

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set used for dependency injection.
// This set contains the provider functions required to create LoginUseCase instances.
// When using Google Wire for dependency injection, this set can be used to automatically assemble required dependencies.
var ProviderSet = wire.NewSet(NewLoginUseCase)
