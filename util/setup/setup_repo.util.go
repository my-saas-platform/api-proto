package setuputil

import (
	strerrors "errors"
	"io"
	"strings"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	consulapi "github.com/hashicorp/consul/api"
	apppkg "github.com/ikaiguang/go-srv-kit/kratos/app"
	authpkg "github.com/ikaiguang/go-srv-kit/kratos/auth"
	registrypkg "github.com/ikaiguang/go-srv-kit/kratos/registry"
	configs "github.com/my-saas-platform/saas-api-proto/api/config"
	pkgerrors "github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"gorm.io/gorm"
)

var (
	_ Config = &configuration{}
	_ Engine = &engines{}

	// ErrUnimplemented 未实现
	ErrUnimplemented = strerrors.New("unimplemented")
	// ErrUninitialized 未初始化
	ErrUninitialized = strerrors.New("uninitialized")
)

// IsUnimplementedError 未实现
func IsUnimplementedError(err error) bool {
	return strerrors.Is(pkgerrors.Cause(err), ErrUnimplemented)
}

// IsUninitializedError 未初始化
func IsUninitializedError(err error) bool {
	return strerrors.Is(pkgerrors.Cause(err), ErrUninitialized)
}

// LoggerPrefixField with logger fields.
type LoggerPrefixField struct {
	AppName    string `json:"name"`
	AppVersion string `json:"version"`
	AppEnv     string `json:"env"`
	Hostname   string `json:"hostname"`
	ServerIP   string `json:"serverIP"`
}

// String 日志前缀
func (s *LoggerPrefixField) String() string {
	strSlice := []string{
		"name=" + `"` + s.AppName + `"`,
		"hostname=" + `"` + s.Hostname + `"`,
		"env=" + `"` + s.AppEnv + `"`,
		"version=" + `"` + s.AppVersion + `"`,
	}
	return strings.Join(strSlice, " ")
}

// Prefix 日志前缀
func (s *LoggerPrefixField) Prefix() []interface{} {
	ss := []string{
		"hostname=" + `"` + s.Hostname + `"`,
		"env=" + `"` + s.AppEnv + `"`,
		"version=" + `"` + s.AppVersion + `"`,
	}
	return []interface{}{
		"service", s.AppName,
		"app", strings.Join(ss, " "),
	}
}

// Config 配置
type Config interface {
	Close() error
	Watch(key string, o config.Observer) error
	Scan(key string, value interface{}) error

	ParseEnv(appEnv string) apppkg.RuntimeEnvEnum_RuntimeEnv
	// RuntimeEnv app环境
	RuntimeEnv() apppkg.RuntimeEnvEnum_RuntimeEnv
	// IsDebugMode 是否启用 调试模式
	IsDebugMode() bool
	// EnableLoggingConsole 是否启用 日志输出到控制台
	EnableLoggingConsole() bool
	// EnableLoggingFile 是否启用 日志输出到文件
	EnableLoggingFile() bool

	// LoggerConfigForConsole 日志配置 控制台
	LoggerConfigForConsole() *configs.Infrastructure_Log_Console
	// LoggerConfigForFile 日志配置 文件
	LoggerConfigForFile() *configs.Infrastructure_Log_File

	// AppConfig APP配置
	AppConfig() *configs.App
	HTTPConfig() *configs.Server_HTTP
	GRPCConfig() *configs.Server_GRPC
	SettingConfig() *configs.Setting
	InfrastructureConfig() *configs.Infrastructure
	ClientApiConfig() *configs.ClientApi
	TokenEncryptConfig() *configs.Setting_EncryptSecret_TokenEncrypt

	// MySQLConfig mysql配置
	MySQLConfig() *configs.Infrastructure_MySQL
	PostgresConfig() *configs.Infrastructure_PSQL
	RedisConfig() *configs.Infrastructure_Redis
	ConsulConfig() *configs.Infrastructure_Consul
	JaegerConfig() *configs.Infrastructure_Jaeger
	SnowflakeWorkerConfig() *configs.Infrastructure_Snowflake
}

// Engine 引擎模块、组件、单元
type Engine interface {
	// Config 配置
	Config

	Close() error

	// Logger 日志处理实例 runtime.caller.skip + 1
	// 用于 log.Helper 输出；例子：log.Helper.Info
	Logger() (log.Logger, []io.Closer, error)
	// LoggerHelper 日志处理实例 runtime.caller.skip + 2
	// 用于包含 log.Helper 输出；例子：func Info(){log.Helper.Info()}
	LoggerHelper() (log.Logger, []io.Closer, error)
	// LoggerMiddleware 日志处理实例 runtime.caller.skip - 1
	// 用于包含 http.Middleware(logging.Server)
	LoggerMiddleware() (log.Logger, []io.Closer, error)

	// GetMySQLGormDB mysql gorm 数据库
	GetMySQLGormDB() (*gorm.DB, error)
	GetPostgresGormDB() (*gorm.DB, error)
	GetRedisClient() (redis.UniversalClient, error)

	// SetRegistryType 设置 服务注册类型
	SetRegistryType(rt registrypkg.RegistryType)
	GetRegistryType() registrypkg.RegistryType
	GetConsulClient() (*consulapi.Client, error)
	GetJaegerExporter() (*jaeger.Exporter, error)

	// InitTracerProvider ...
	InitTracerProvider() error

	// GetAuthTokenRepo 验证Token工具
	GetAuthTokenRepo(redisCC redis.UniversalClient) (authpkg.AuthRepo, error)
}
