package biz

import (
	"github.com/google/wire"
)

// ProviderSet 是一个 Wire 提供器集合，用于依赖注入。
// 该集合包含了创建 LoginUseCase 实例所需的提供器函数。
// 在使用 Google Wire 进行依赖注入时，可通过该集合来自动装配所需的依赖项。
var ProviderSet = wire.NewSet(NewLoginUseCase)
