package module

import "sync"

// 组件注册器接口
type registrar interface {
	// 注册组件实例
	Register(module Module) (bool, error)
	// 注销组件实例
	Unregister(mid MID) (bool, error)
	// 获取一个指定类型的组件实例，
	// 该函数基于负载均衡策略返回实例(通过组件评分实现均衡策略)
	Get(moduleType Type) (Module, error)
	// 获取指定类型的所有组件实例
	GetAllByType(moduleType Type) (map[MID]Module, error)
	// 获取所有实例
	GetAll() map[MID]Module
	// 清除所有组件注册记录
	Clear()
}

type myRegistrar struct {
	// 组件类型与对应组件实例映射表 (双层的字典结构)
	moduleTypeMap map[Type]map[MID]Module
	// 组件注册专用读写锁
	rwlock sync.RWMutex
}

func (registrar *myRegistrar) Register(module Module) (bool, error) {

}

// 注销组件实例
func (registrar *myRegistrar) Unregister(mid MID) (bool, error) {

}

// 获取一个指定类型的组件实例，
// 该函数基于负载均衡策略返回实例(通过组件评分实现均衡策略)
func (registrar *myRegistrar) Get(moduleType Type) (Module, error) {

}

// 获取指定类型的所有组件实例
func (registrar *myRegistrar) GetAllByType(moduleType Type) (map[MID]Module, error) {

}

// 获取所有实例
func (registrar *myRegistrar) GetAll() map[MID]Module {

}

// 清除所有组件注册记录
func (registrar *myRegistrar) Clear() {

}