package clientutil

import (
	pingservicev1 "github.com/ikaiguang/go-srv-kit/api/ping/v1/services"
	setuputil "github.com/my-saas-platform/saas-api-proto/util/setup"
)

// NewPingHTTPClient ...
func NewPingHTTPClient(engineHandler setuputil.Engine, serviceName ServiceName) (pingservicev1.SrvPingHTTPClient, error) {
	conn, err := NewHTTPConnection(engineHandler, serviceName)
	if err != nil {
		return nil, err
	}
	return pingservicev1.NewSrvPingHTTPClient(conn), nil
}

// NewUserHTTPClient ...
//func NewUserHTTPClient(engineHandler setuputil.Engine, serviceName ServiceName) (userservicev1.SrvUserHTTPClient, error) {
//	conn, err := NewHTTPConnection(engineHandler, serviceName)
//	if err != nil {
//		return nil, err
//	}
//	return userservicev1.NewSrvUserHTTPClient(conn), nil
//}

// NewAdminHTTPClient ...
//func NewAdminHTTPClient(engineHandler setuputil.Engine, serviceName ServiceName) (adminservicev1.SrvAdminHTTPClient, error) {
//	conn, err := NewHTTPConnection(engineHandler, serviceName)
//	if err != nil {
//		return nil, err
//	}
//	return adminservicev1.NewSrvAdminHTTPClient(conn), nil
//}
