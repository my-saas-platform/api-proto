package apputil

import (
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/transport/http"
	apppkg "github.com/ikaiguang/go-srv-kit/kratos/app"
	errorpkg "github.com/ikaiguang/go-srv-kit/kratos/error"
	configs "github.com/my-saas-platform/saas-api-proto/api/config"
)

const (
	_appIDSep      = ":"
	_configPathSep = "/"
)

// ID 程序ID
// 例：go-srv-saas/DEVELOP/main/v1.0.0/user-service
func ID(appConfig *configs.App) string {
	return appIdentifier(appConfig, _appIDSep)
}

// ConfigPath 配置路径；用于配置中心，如：consul、etcd、...
// @result = app.ProjectName + "/" + app.ServerName + "/" + app.ServerEnv + "/" + app.ServerVersion
// 例：go-srv-saas/DEVELOP/main/v1.0.0/user-service
func ConfigPath(appConfig *configs.App) string {
	return appIdentifier(appConfig, _configPathSep)
}

// appIdentifier app 唯一标准
// @result = app.ProjectName + "/" + app.ServerName + "/" + app.ServerEnv + "/" + app.ServerVersion
// 例：go-srv-saas/DEVELOP/main/v1.0.0/user-service
func appIdentifier(appConfig *configs.App, sep string) string {
	var ss = make([]string, 0, 5)
	if appConfig.ProjectName != "" {
		ss = append(ss, appConfig.ProjectName)
	}
	if appConfig.ServerName != "" {
		ss = append(ss, appConfig.ServerName)
	}
	ss = append(ss, apppkg.ParseEnv(appConfig.ServerEnv).String())
	if appConfig.ServerVersion != "" {
		ss = append(ss, appConfig.ServerVersion)
	}
	return strings.Join(ss, sep)
}

// RuntimePath ...
func RuntimePath() (string, error) {
	p, err := os.Getwd()
	if err != nil {
		e := errorpkg.ErrorInternalServer("os get runtime path failed")
		return "", errorpkg.WithStack(e)
	}
	return p, nil
}

// ServerDecoderEncoder ...
func ServerDecoderEncoder() []http.ServerOption {
	apppkg.RegisterCodec()
	return []http.ServerOption{
		http.RequestDecoder(apppkg.RequestDecoder),
		http.ResponseEncoder(apppkg.SuccessResponseEncoder),
		http.ErrorEncoder(apppkg.ErrorResponseEncoder),
	}
}

// ClientDecoderEncoder ...
func ClientDecoderEncoder() []http.ClientOption {
	return []http.ClientOption{
		http.WithResponseDecoder(apppkg.SuccessResponseDecoder),
	}
}
