package setuputil

import (
	"io"
	stdlog "log"
	"sync"

	consul "github.com/go-kratos/kratos/contrib/config/consul/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	consulpkg "github.com/ikaiguang/go-srv-kit/data/consul"
	apppkg "github.com/ikaiguang/go-srv-kit/kratos/app"
	authpkg "github.com/ikaiguang/go-srv-kit/kratos/auth"
	registrypkg "github.com/ikaiguang/go-srv-kit/kratos/registry"
	configs "github.com/my-saas-platform/saas-api-proto/api/config"
	apputil "github.com/my-saas-platform/saas-api-proto/util/app"
	pkgerrors "github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/jaeger"

	"github.com/go-kratos/kratos/v2/log"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// engines 引擎模块
type engines struct {
	Config

	// registryType 服务注册类型
	registryType registrypkg.RegistryType

	// loggerPrefixFieldMutex 日志前缀
	loggerPrefixFieldMutex sync.Once
	loggerPrefixField      *LoggerPrefixField

	// loggerFileWriterMutex 日志文件写手柄
	loggerFileWriterMutex sync.Once
	loggerFileWriter      io.Writer

	// debugHelperCloseFnSlice debug工具
	debugHelperCloseFnSlice []io.Closer

	// loggerMutex 日志
	loggerMutex                  sync.Once
	logger                       log.Logger
	loggerCloseFnSlice           []io.Closer
	loggerHelperMutex            sync.Once
	loggerHelper                 log.Logger
	loggerHelperCloseFnSlice     []io.Closer
	loggerMiddlewareMutex        sync.Once
	loggerMiddleware             log.Logger
	loggerMiddlewareCloseFnSlice []io.Closer

	// mysqlGormMutex mysql gorm
	mysqlGormMutex sync.Once
	mysqlGormDB    *gorm.DB

	// postgresGormMutex mysql gorm
	postgresGormMutex sync.Once
	postgresGormDB    *gorm.DB

	// redisClientMutex redis客户端
	redisClientMutex sync.Once
	redisClient      redis.UniversalClient

	// consulClientMutex consul客户端
	consulClientMutex sync.Once
	consulClient      *consulapi.Client

	// jaegerTraceExporterMutex jaeger trace
	jaegerTraceExporterMutex sync.Once
	jaegerTraceExporter      *jaeger.Exporter

	// snowflakeStopChannel 雪花算法
	snowflakeStopChannel chan int

	// authTokenRepoMutex 验证Token工具
	authTokenRepoMutex sync.Once
	authTokenRepo      authpkg.AuthRepo
}

// configuration 实现ConfigInterface
type configuration struct {
	// handler 配置处理手柄
	handler config.Config
	// conf 配置引导文件
	conf *configs.Bootstrap

	// env app环境
	env apppkg.RuntimeEnvEnum_RuntimeEnv

	// enableDebug 是否启用 调试模式
	enableDebug bool
	// enableLoggingConsole 是否启用 日志输出到控制台
	enableLoggingConsole bool
	// enableLoggingFile 是否启用 日志输出到文件
	enableLoggingFile bool
}

// NewConfiguration 配置处理手柄
func NewConfiguration(opts ...config.Option) (Config, error) {
	handler := &configuration{}
	if err := handler.init(opts...); err != nil {
		return nil, err
	}
	return handler, nil
}

// init 初始化
func (s *configuration) init(opts ...config.Option) (err error) {
	// 处理手柄
	s.handler = config.New(opts...)

	// 加载配置
	if err = s.handler.Load(); err != nil {
		err = pkgerrors.WithStack(err)
		return
	}

	// 读取配置文件
	s.conf = &configs.Bootstrap{}
	if err = s.handler.Scan(s.conf); err != nil {
		err = pkgerrors.WithStack(err)
		return
	}

	// App配置
	if s.conf.App == nil {
		err = pkgerrors.New("[请配置服务再启动] config key : app")
		return err
	}

	// 服务配置
	if s.conf.Server == nil {
		err = pkgerrors.New("[请配置服务再启动] config key : server")
		return err
	}

	// app环境
	s.env = apppkg.RuntimeEnvEnum_PRODUCTION
	if cfg := s.AppConfig(); cfg != nil {
		// app环境
		s.env = s.ParseEnv(cfg.ServerEnv)
		apppkg.SetRuntimeEnv(s.env)
		s.enableDebug = apppkg.IsDebugMode()
	}

	// 日志
	if cfg := s.LoggerConfigForConsole(); cfg != nil {
		s.enableLoggingConsole = cfg.Enable
	}
	if cfg := s.LoggerConfigForFile(); cfg != nil {
		s.enableLoggingFile = cfg.Enable
	}
	return err
}

// newConfigWithFiles 初始化配置手柄
func newConfigWithFiles(setupOpts *options) (Config, error) {

	stdlog.Println("|==================== 加载配置文件 开始 ====================|")
	defer stdlog.Println()
	defer stdlog.Println("|==================== 加载配置文件 结束 ====================|")
	// 配置路径
	var confPath = "../../configs"
	if setupOpts.configPath != "" {
		confPath = setupOpts.configPath
	}

	p, err := apputil.RuntimePath()
	if err != nil {
		return nil, err
	}
	stdlog.Println("|*** INFO：当前程序运行路径: ", p)

	var opts []config.Option
	stdlog.Println("|*** 加载：配置文件路径: ", confPath)
	opts = append(opts, config.WithSource(file.NewSource(confPath)))
	return NewConfiguration(opts...)
}

// newConfigWithConsul 初始化配置手柄
func newConfigWithConsul(setupOpts *options) (configImpl Config, consulClient *consulapi.Client, err error) {
	stdlog.Println("|==================== 初始化Consul配置中心 开始 ====================|")
	defer stdlog.Println()
	defer stdlog.Println("|==================== 初始化Consul配置中心 结束 ====================|")

	// 配置路径
	filePath := "../../configs/consul"
	if setupOpts.configPath != "" {
		filePath = setupOpts.configPath
	}
	stdlog.Println("|*** 加载：Consul初始化配置文件路径: ", filePath)
	configHandler := config.New(config.WithSource(
		file.NewSource(filePath),
	))

	// 加载配置
	if err = configHandler.Load(); err != nil {
		err = pkgerrors.WithStack(err)
		return configImpl, consulClient, err
	}

	// 读取配置文件
	cfg := &configs.Bootstrap{}
	if err = configHandler.Scan(cfg); err != nil {
		err = pkgerrors.WithStack(err)
		return configImpl, consulClient, err
	}

	// App配置
	if cfg.App == nil {
		err = pkgerrors.New("[请配置服务再启动] consul key : app")
		return configImpl, consulClient, err
	}

	// 服务配置
	if cfg.Infrastructure.Consul == nil {
		err = pkgerrors.New("[请配置服务再启动] consul key : base.consul")
		return configImpl, consulClient, err
	}

	// consul客户端
	stdlog.Println("|*** 加载：Consul客户端：for 配置中心")
	consulClient, err = consulpkg.NewConsulClient(ToConsulConfig(cfg.Infrastructure.Consul))
	if err != nil {
		err = pkgerrors.WithStack(err)
		return configImpl, consulClient, err
	}

	// 配置source
	consulKeyPath := apputil.ConfigPath(cfg.App)
	stdlog.Println("|*** 加载：Consul配置文件路径：", consulKeyPath)
	cs, err := consul.New(consulClient, consul.WithPath(consulKeyPath))
	if err != nil {
		err = pkgerrors.WithStack(err)
		return configImpl, consulClient, err
	}

	var opts []config.Option
	stdlog.Println("|*** 加载：Consul配置中心的配置: ...")
	opts = append(opts, config.WithSource(cs))

	// config impl
	configImpl, err = NewConfiguration(opts...)
	if err != nil {
		return configImpl, consulClient, err
	}
	return configImpl, consulClient, err
}

// ParseEnv 解析环境
func (s *configuration) ParseEnv(appEnv string) apppkg.RuntimeEnvEnum_RuntimeEnv {
	return apppkg.ParseEnv(appEnv)
}

// Watch 监听
func (s *configuration) Watch(key string, o config.Observer) error {
	return s.handler.Watch(key, o)
}

// Scan 读取配置
func (s *configuration) Scan(key string, value interface{}) error {
	return s.handler.Value(key).Scan(value)
}

// Close 关闭
func (s *configuration) Close() error {
	return s.handler.Close()
}

// RuntimeEnv app环境
func (s *configuration) RuntimeEnv() apppkg.RuntimeEnvEnum_RuntimeEnv {
	return s.env
}

// IsDebugMode 是否启用 调试模式
func (s *configuration) IsDebugMode() bool {
	return s.enableDebug
}

// EnableLoggingConsole 是否启用 日志输出到控制台
func (s *configuration) EnableLoggingConsole() bool {
	return s.enableLoggingConsole
}

// EnableLoggingFile 是否启用 日志输出到文件
func (s *configuration) EnableLoggingFile() bool {
	return s.enableLoggingFile
}

// AppConfig APP配置
func (s *configuration) AppConfig() *configs.App {
	return s.conf.App
}

// ServerConfig 服务配置
func (s *configuration) ServerConfig() *configs.Server {
	return s.conf.Server
}

// HTTPConfig http配置
func (s *configuration) HTTPConfig() *configs.Server_HTTP {
	if s.conf.Server == nil {
		return nil
	}
	return s.conf.Server.Http
}

// GRPCConfig grpc配置
func (s *configuration) GRPCConfig() *configs.Server_GRPC {
	if s.conf.Server == nil {
		return nil
	}
	return s.conf.Server.Grpc
}

// SettingConfig APP配置
func (s *configuration) SettingConfig() *configs.Setting {
	return s.conf.Setting
}

// InfrastructureConfig ...
func (s *configuration) InfrastructureConfig() *configs.Infrastructure {
	return s.conf.Infrastructure
}

// ClientApiConfig ...
func (s *configuration) ClientApiConfig() *configs.ClientApi {
	return s.conf.ClientApi
}

// TokenEncryptConfig ...
func (s *configuration) TokenEncryptConfig() *configs.Setting_EncryptSecret_TokenEncrypt {
	if s.conf.Setting == nil || s.conf.Setting.EncryptSecret == nil {
		return nil
	}
	return s.conf.Setting.EncryptSecret.TokenEncrypt
}

// LoggerConfigForConsole 日志配置 控制台
func (s *configuration) LoggerConfigForConsole() *configs.Infrastructure_Log_Console {
	if s.conf.Infrastructure == nil || s.conf.Infrastructure.Log == nil {
		return nil
	}
	return s.conf.Infrastructure.Log.Console
}

// LoggerConfigForFile 日志配置 文件
func (s *configuration) LoggerConfigForFile() *configs.Infrastructure_Log_File {
	if s.conf.Infrastructure == nil || s.conf.Infrastructure.Log == nil {
		return nil
	}
	return s.conf.Infrastructure.Log.File
}

// MySQLConfig mysql配置
func (s *configuration) MySQLConfig() *configs.Infrastructure_MySQL {
	if s.conf.Infrastructure == nil {
		return nil
	}
	return s.conf.Infrastructure.Mysql
}

// PostgresConfig mysql配置
func (s *configuration) PostgresConfig() *configs.Infrastructure_PSQL {
	if s.conf.Infrastructure == nil {
		return nil
	}
	return s.conf.Infrastructure.Psql
}

// RedisConfig redis配置
func (s *configuration) RedisConfig() *configs.Infrastructure_Redis {
	if s.conf.Infrastructure == nil {
		return nil
	}
	return s.conf.Infrastructure.Redis
}

// ConsulConfig consul配置
func (s *configuration) ConsulConfig() *configs.Infrastructure_Consul {
	if s.conf.Infrastructure == nil {
		return nil
	}
	return s.conf.Infrastructure.Consul
}

// JaegerConfig jaeger 配置
func (s *configuration) JaegerConfig() *configs.Infrastructure_Jaeger {
	if s.conf.Infrastructure == nil {
		return nil
	}
	return s.conf.Infrastructure.Jaeger
}

// SnowflakeWorkerConfig snowflake worker 配置
func (s *configuration) SnowflakeWorkerConfig() *configs.Infrastructure_Snowflake {
	if s.conf.Infrastructure == nil {
		return nil
	}
	return s.conf.Infrastructure.Snowflake
}
