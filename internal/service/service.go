package service

import (
	"github.com/google/wire"
)

// ProviderSet 是 Google Wire 的提供器集合，用于依赖注入。
// 该集合包含了 NewLoginService 函数，这意味着在使用 Google Wire 进行依赖注入时，
// 可以自动创建 LoginService 实例。通过将该函数添加到集合中，
// 可以方便地在项目中管理和初始化 LoginService 相关的依赖关系。
var ProviderSet = wire.NewSet(NewLoginService)
