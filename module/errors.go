package module

import "errors"

// 未找到组件实例错误
var ErrNotFoundModuleInstance = errors.New("not found module instance")