package clientutil

import (
	errorpkg "github.com/ikaiguang/go-srv-kit/kratos/error"
	configs "github.com/my-saas-platform/api-proto/api/config"
	setuputil "github.com/my-saas-platform/api-proto/util/setup"
)

// ServiceName ...
type ServiceName string

func (s ServiceName) String() string {
	return string(s)
}

const (
	PingService      ServiceName = "ping-service"
	SnowflakeService ServiceName = "snowflake-service"
	UserService      ServiceName = "user-service"
	AdminService     ServiceName = "admin-service"

	FeishuApi      ServiceName = "feishu-openapi"
	DingtalkApi    ServiceName = "dingtalk-openapi"
	DingtalkApiOld ServiceName = "dingtalk-openapi-old"
)

// ClientApiEndpoint ...
type ClientApiEndpoint struct {
	Name         string
	RegistryName string
	HttpHost     string
	GrpcHost     string
}

func (s *ClientApiEndpoint) SetByPbClientApiEndpoint(cfg *configs.ClientApi_Endpoint) {
	s.Name = cfg.Name
	s.RegistryName = cfg.RegistryName
	s.HttpHost = cfg.HttpHost
	s.GrpcHost = cfg.GrpcHost
}

// getClientApiConfig ...
func getClientApiConfig(engineHandler setuputil.Engine, serviceName ServiceName) (*ClientApiEndpoint, error) {
	apiConfig := engineHandler.ClientApiConfig()
	if apiConfig == nil {
		msg := "请先配置: client_api"
		e := errorpkg.ErrorInvalidParameter(msg)
		return nil, errorpkg.WithStack(e)
	}
	for _, cfg := range apiConfig.ClusterService {
		if cfg.Name == serviceName.String() {
			res := new(ClientApiEndpoint)
			res.SetByPbClientApiEndpoint(cfg)
			return res, nil
		}
	}
	for _, cfg := range apiConfig.ThirdParty {
		if cfg.Name == serviceName.String() {
			res := new(ClientApiEndpoint)
			res.SetByPbClientApiEndpoint(cfg)
			return res, nil
		}
	}
	msg := "请先配置: client_api.xxx.name: " + serviceName.String()
	e := errorpkg.ErrorInvalidParameter(msg)
	return nil, errorpkg.WithStack(e)
}
