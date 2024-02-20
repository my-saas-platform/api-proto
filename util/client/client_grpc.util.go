package clientutil

import (
	pingservicev1 "github.com/ikaiguang/go-srv-kit/api/ping/v1/services"
	setuputil "github.com/my-saas-platform/saas-api-proto/util/setup"
)

// NewPingGRPCClient ...
func NewPingGRPCClient(engineHandler setuputil.Engine, serviceName ServiceName) (pingservicev1.SrvPingClient, error) {
	conn, err := NewGRPCConnection(engineHandler, serviceName)
	if err != nil {
		return nil, err
	}
	return pingservicev1.NewSrvPingClient(conn), nil
}

// NewUserGRPCClient ...
//func NewUserGRPCClient(engineHandler setuputil.Engine, serviceName ServiceName) (userservicev1.SrvUserClient, error) {
//	conn, err := NewGRPCConnection(engineHandler, serviceName)
//	if err != nil {
//		return nil, err
//	}
//	return userservicev1.NewSrvUserClient(conn), nil
//}

// NewAdminGRPCClient ...
//func NewAdminGRPCClient(engineHandler setuputil.Engine, serviceName ServiceName) (adminservicev1.SrvAdminClient, error) {
//	conn, err := NewGRPCConnection(engineHandler, serviceName)
//	if err != nil {
//		return nil, err
//	}
//	return adminservicev1.NewSrvAdminClient(conn), nil
//}
