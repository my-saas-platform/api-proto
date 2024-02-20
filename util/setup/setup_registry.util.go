package setuputil

import registrypkg "github.com/ikaiguang/go-srv-kit/kratos/registry"

// SetRegistryType 设置 服务注册类型
func (s *engines) SetRegistryType(rt registrypkg.RegistryType) {
	s.registryType = rt
}

// GetRegistryType 服务注册类型
func (s *engines) GetRegistryType() registrypkg.RegistryType {
	return s.registryType
}
