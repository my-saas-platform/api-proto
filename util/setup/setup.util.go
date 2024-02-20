package setuputil

import (
	"flag"
	stdlog "log"

	registrypkg "github.com/ikaiguang/go-srv-kit/kratos/registry"
)

var (
	configFlag string
)

func init() {
	flag.StringVar(&configFlag, "conf", "", "config path, eg: -conf ./configs")
}

// options 配置可选项
type options struct {
	configPath       string
	consulConfigPath string
}

// Option is config option.
type Option func(*options)

// WithConfigPath 配置路径
func WithConfigPath(configPath string) Option {
	return func(o *options) {
		o.configPath = configPath
	}
}

// WithConsulConfigPath consul配置路径
func WithConsulConfigPath(consulConfigPath string) Option {
	return func(o *options) {
		o.consulConfigPath = consulConfigPath
	}
}

// New 启动与配置
func New(opts ...Option) (engineHandler Engine, err error) {
	if !flag.Parsed() {
		flag.Parse()
	}
	// 启动选项
	setupOpts := &options{
		configPath: configFlag,
	}
	for i := range opts {
		opts[i](setupOpts)
	}

	// 配置方式
	var (
		configHandler Config
	)
	switch {
	case setupOpts.consulConfigPath != "":
		// 初始化配置手柄
		configHandler, _, err = newConfigWithConsul(setupOpts)
		if err != nil {
			return engineHandler, err
		}
	default:
		// 初始化配置手柄
		configHandler, err = newConfigWithFiles(setupOpts)
		if err != nil {
			return engineHandler, err
		}
	}

	// 开始配置
	stdlog.Println("|==================== 配置程序 开始 ====================|")
	defer stdlog.Println("|==================== 配置程序 结束 ====================|")

	return newEngine(configHandler)
}

// initEngine ...
func initEngine(conf Config) *engines {
	return &engines{
		Config: conf,
	}
}

// newEngine 启动与配置
func newEngine(configHandler Config) (Engine, error) {
	// 初始化手柄
	var (
		err          error
		setupHandler = initEngine(configHandler)
	)

	// 设置调试工具
	if err = setupHandler.loadingDebugUtil(); err != nil {
		return nil, err
	}

	// 设置日志工具
	if _, err = setupHandler.loadingLogHelper(); err != nil {
		return nil, err
	}

	// mysql gorm 数据库
	if cfg := setupHandler.Config.MySQLConfig(); cfg != nil && cfg.Enable {
		if _, err = setupHandler.GetMySQLGormDB(); err != nil {
			return nil, err
		}
	}

	// postgres gorm 数据库
	if cfg := setupHandler.Config.PostgresConfig(); cfg != nil && cfg.Enable {
		if _, err = setupHandler.GetPostgresGormDB(); err != nil {
			return nil, err
		}
	}

	// redis 客户端
	if cfg := setupHandler.Config.RedisConfig(); cfg != nil && cfg.Enable {
		redisCC, err := setupHandler.GetRedisClient()
		if err != nil {
			return nil, err
		}
		// 验证Token工具
		_, _ = setupHandler.GetAuthTokenRepo(redisCC)
	}

	// 服务注册
	setupHandler.SetRegistryType(registrypkg.RegistryTypeLocal)

	// consul 客户端
	if cfg := setupHandler.Config.ConsulConfig(); cfg != nil && cfg.Enable {
		_, err = setupHandler.GetConsulClient()
		if err != nil {
			return nil, err
		}
	}

	// jaeger
	if cfg := setupHandler.Config.JaegerConfig(); cfg != nil && cfg.Enable {
		_, err = setupHandler.GetJaegerExporter()
		if err != nil {
			return nil, err
		}
	}

	// 雪花算法
	if cfg := setupHandler.Config.SettingConfig(); cfg != nil && cfg.EnableSnowflakeWorker {
		err = setupHandler.loadingSnowflakeWorker()
		if err != nil {
			return nil, err
		}
	}

	// 监听配置 app
	//if err = setupHandler.watchConfigApp(); err != nil {
	//	return nil, err
	//}

	// 监听配置 data
	//if err = setupHandler.watchConfigData(); err != nil {
	//	return nil, err
	//}

	return setupHandler, err
}
