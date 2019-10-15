package module

import (
	"fmt"
	"sync"
	"web_crawler/errors"
)

// 组件注册器接口
type Registrar interface {
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

func NewRegistrar() Registrar {
	return &myRegistrar{
		moduleTypeMap: map[Type]map[MID]Module{},
	}
}

func (registrar *myRegistrar) Register(module Module) (bool, error) {
	if module == nil {
		return false, errors.NewIllegalParameterError("nil module instance")
	}
	mid := module.ID()
	parts, _ := SplitMID(mid)

	moduleType := legalLetterTypeMap[parts[0]]
	if !CheckType(moduleType, module) {
		return false, errors.NewIllegalParameterError(
			fmt.Sprintf("incorrect module type: %s", moduleType))
	}

	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()

	modules := registrar.moduleTypeMap[moduleType]
	if modules == nil {
		modules = map[MID]Module{}
	}

	if _,ok := modules[mid]; ok {
		return false, nil
	}
	modules[mid] = module
	registrar.moduleTypeMap[moduleType] = modules
	return true, nil
}

// 注销组件实例
func (registrar *myRegistrar) Unregister(mid MID) (bool, error) {
	parts, err := SplitMID(mid)
	if err != nil {
		return false, err
	}

	moduleType := legalLetterTypeMap[parts[0]]

	var deleted bool
	registrar.rwlock.Lock()
	defer registrar.rwlock.Lock()
	if modules, ok := registrar.moduleTypeMap[moduleType]; ok {
		if _, ok := modules[mid]; ok {
			delete(modules, mid)
			deleted = true
		}
	}
	return deleted, nil
}

// 获取一个指定类型的组件实例，
// 该函数基于负载均衡策略返回实例(通过组件评分实现均衡策略)
func (registrar *myRegistrar) Get(moduleType Type) (Module, error) {
	modules, err := registrar.GetAllByType(moduleType)
	if err != nil {
		return nil, err
	}

	// 返回评分最低的组件实例
	minScore := uint64(0)
	var selectedModule Module
	for _, module := range modules {
		SetScore(module)
		score := module.Score()
		if minScore == 0 || score < minScore {
			selectedModule = module
			minScore = score
		}
	}
	return selectedModule, nil
}

// 获取指定类型的所有组件实例
func (registrar *myRegistrar) GetAllByType(moduleType Type) (map[MID]Module, error) {
	if !LegalType(moduleType) {
		return nil, errors.NewIllegalParameterError(
			fmt.Sprintf("illegal module type: %s", moduleType))
	}
	registrar.rwlock.RLock()
	defer registrar.rwlock.RUnlock()
	modules := registrar.moduleTypeMap[moduleType]
	if len(modules) == 0 {
		return nil, ErrNotFoundModuleInstance
	}
	return modules,nil
}

// 获取所有实例
func (registrar *myRegistrar) GetAll() map[MID]Module {
	result := map[MID]Module{}
	registrar.rwlock.RLock()
	defer registrar.rwlock.RUnlock()
	for _, modules := range registrar.moduleTypeMap {
		for mid, module := range modules {
			result[mid] = module
		}
	}
	return result
}

// 清除所有组件注册记录
func (registrar *myRegistrar) Clear() {
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	registrar.moduleTypeMap = map[Type]map[MID]Module{}
}