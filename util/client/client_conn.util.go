package clientutil

import (
	"context"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	curlpkg "github.com/ikaiguang/go-srv-kit/kit/curl"
	clientpkg "github.com/ikaiguang/go-srv-kit/kratos/client"
	errorpkg "github.com/ikaiguang/go-srv-kit/kratos/error"
	logpkg "github.com/ikaiguang/go-srv-kit/kratos/log"
	middlewarepkg "github.com/ikaiguang/go-srv-kit/kratos/middleware"
	registrypkg "github.com/ikaiguang/go-srv-kit/kratos/registry"
	configs "github.com/my-saas-platform/saas-api-proto/api/config"
	apputil "github.com/my-saas-platform/saas-api-proto/util/app"
	setuputil "github.com/my-saas-platform/saas-api-proto/util/setup"
	pkgerrors "github.com/pkg/errors"
	stdgrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
	defaultTimeout = curlpkg.DefaultTimeout
)

var (
	_httpConnections = sync.Map{}
	_grpcConnections = sync.Map{}
)

// NewGRPCConnection grpc 链接
func NewGRPCConnection(engineHandler setuputil.Engine, serviceName ServiceName, otherOpts ...grpc.ClientOption) (*stdgrpc.ClientConn, error) {
	cc, ok := _grpcConnections.Load(serviceName)
	if ok {
		if conn, ok := cc.(*stdgrpc.ClientConn); ok {
			return conn, nil
		}
	}

	conn, err := newGRPCConnection(engineHandler, serviceName, otherOpts...)
	if err != nil {
		return nil, err
	}
	_grpcConnections.Store(serviceName, conn)

	return conn, nil
}

// NewHTTPConnection http 链接
func NewHTTPConnection(engineHandler setuputil.Engine, serviceName ServiceName, otherOpts ...http.ClientOption) (*http.Client, error) {
	cc, ok := _httpConnections.Load(serviceName)
	if ok {
		if conn, ok := cc.(*http.Client); ok {
			return conn, nil
		}
	}

	conn, err := newHTTPConnection(engineHandler, serviceName, otherOpts...)
	if err != nil {
		return nil, err
	}
	_httpConnections.Store(serviceName, conn)

	return conn, nil
}

// newGRPCConnection grpc 链接
func newGRPCConnection(engineHandler setuputil.Engine, serviceName ServiceName, otherOpts ...grpc.ClientOption) (*stdgrpc.ClientConn, error) {
	var opts = []grpc.ClientOption{
		grpc.WithTimeout(defaultTimeout),
	}

	// 服务端点
	endpointOpts, err := getGRPCEndpoint(engineHandler, serviceName)
	if err != nil {
		return nil, err
	}
	opts = append(opts, endpointOpts...)

	// 中间件
	logger, _, err := engineHandler.Logger()
	if err != nil {
		return nil, err
	}
	logHelper := log.NewHelper(logger)
	opts = append(opts, grpc.WithMiddleware(middlewarepkg.DefaultClientMiddlewares(logHelper)...))
	// 其他
	opts = append(opts, otherOpts...)

	// grpc 链接
	conn, err := grpc.DialInsecure(
		context.Background(),
		opts...,
	)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}
	return conn, nil
}

// newHTTPConnection http 链接
func newHTTPConnection(engineHandler setuputil.Engine, serviceName ServiceName, otherOpts ...http.ClientOption) (*http.Client, error) {
	var opts = []http.ClientOption{
		http.WithTimeout(defaultTimeout),
	}
	opts = append(opts, apputil.ClientDecoderEncoder()...)

	// 服务端点
	endpointOpts, err := getHTTPEndpoint(engineHandler, serviceName)
	if err != nil {
		return nil, err
	}
	opts = append(opts, endpointOpts...)

	// 中间件
	logger, _, err := engineHandler.Logger()
	if err != nil {
		return nil, err
	}
	logHelper := log.NewHelper(logger)
	opts = append(opts, http.WithMiddleware(middlewarepkg.DefaultClientMiddlewares(logHelper)...))
	// 其他
	opts = append(opts, otherOpts...)

	// http 链接
	conn, err := clientpkg.NewHTTPClient(context.Background(), opts...)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}
	return conn, nil
}

// getHTTPEndpoint 获取服务端点
func getHTTPEndpoint(engineHandler setuputil.Engine, serviceName ServiceName) ([]http.ClientOption, error) {
	endpointInfo, err := getClientApiConfig(engineHandler, serviceName)
	if err != nil {
		return nil, err
	}

	var (
		clientKind   = transport.KindHTTP
		opts         []http.ClientOption
		registryType = engineHandler.GetRegistryType()
		discovery    registry.Discovery
		endpoint     string
	)
	switch registryType {
	case registrypkg.RegistryTypeConsul:
		discovery, endpoint, err = getRegistryAndServerEndpoint(engineHandler, serviceName, endpointInfo.RegistryName)
		if err != nil {
			return nil, err
		}
		opts = append(opts, http.WithDiscovery(discovery))
	default:
		endpoint = endpointInfo.HttpHost
	}
	logpkg.Infow(
		"client.kind", clientKind,
		"client.serviceName", serviceName,
		"client.registryType", registryType,
		"client.endpoint", endpoint,
	)
	opts = append(opts, http.WithEndpoint(endpoint))
	return opts, nil
}

// getGRPCEndpoint 获取服务端点
func getGRPCEndpoint(engineHandler setuputil.Engine, serviceName ServiceName) ([]grpc.ClientOption, error) {
	endpointInfo, err := getClientApiConfig(engineHandler, serviceName)
	if err != nil {
		return nil, err
	}

	var (
		clientKind   = transport.KindGRPC
		opts         []grpc.ClientOption
		registryType = engineHandler.GetRegistryType()
		discovery    registry.Discovery
		endpoint     string
	)
	switch registryType {
	case registrypkg.RegistryTypeConsul:
		discovery, endpoint, err = getRegistryAndServerEndpoint(engineHandler, serviceName, endpointInfo.RegistryName)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithDiscovery(discovery))
	default:
		endpoint = endpointInfo.GrpcHost
	}
	logpkg.Infow(
		"client.kind", clientKind,
		"client.serviceName", serviceName,
		"client.registryType", registryType,
		"client.endpoint", endpoint,
	)
	opts = append(opts, grpc.WithEndpoint(endpoint))
	return opts, nil
}

// getRegistryAndServerEndpoint ...
func getRegistryAndServerEndpoint(engineHandler setuputil.Engine, serviceName ServiceName, registryName string) (*consul.Registry, string, error) {
	if registryName = strings.TrimSpace(registryName); registryName == "" {
		msg := "请先配置: client_api.xxx.name." + serviceName.String() + ".registry_name"
		e := errorpkg.ErrorInvalidParameter(msg)
		return nil, "", errorpkg.WithStack(e)
	}

	consulClient, err := engineHandler.GetConsulClient()
	if err != nil {
		return nil, "", err
	}
	r, err := registrypkg.NewConsulRegistry(consulClient)
	if err != nil {
		return nil, "", err
	}
	// 端点 "discovery:///${registry_endpoint}"
	// registry_endpoint = ${app.belong_to}/${app.env}/${app.env_branch}/${app.version}/${app.name}
	// 例子：registry_endpoint = go-srv-saas/DEVELOP/main/v1.0.0/saas-user-service
	appConfig := proto.Clone(engineHandler.AppConfig()).(*configs.App)
	appConfig.ServerName = registryName
	endpoint := "discovery:///" + apputil.ID(appConfig)

	return r, endpoint, nil
}
