package setuputil

import (
	stdlog "log"

	consulpkg "github.com/ikaiguang/go-srv-kit/data/consul"
	jaegerpkg "github.com/ikaiguang/go-srv-kit/data/jaeger"
	mysqlpkg "github.com/ikaiguang/go-srv-kit/data/mysql"
	psqlpkg "github.com/ikaiguang/go-srv-kit/data/postgres"
	redispkg "github.com/ikaiguang/go-srv-kit/data/redis"
	debugpkg "github.com/ikaiguang/go-srv-kit/debug"
	configs "github.com/my-saas-platform/api-proto/api/config"
	pkgerrors "github.com/pkg/errors"
)

// loadingDebugUtil 加载调试工具
func (s *engines) loadingDebugUtil() error {
	if !s.Config.IsDebugMode() {
		return nil
	}
	stdlog.Printf("|*** 加载：调试工具debugutil")
	syncFn, err := debugpkg.Setup()
	if err != nil {
		return pkgerrors.WithStack(err)
	}
	s.debugHelperCloseFnSlice = append(s.debugHelperCloseFnSlice, syncFn)
	return err
}

// ToMysqlConfig ...
func ToMysqlConfig(cfg *configs.Infrastructure_MySQL) *mysqlpkg.Config {
	return &mysqlpkg.Config{
		Dsn:             cfg.Dsn,
		SlowThreshold:   cfg.SlowThreshold,
		LoggerEnable:    cfg.LoggerEnable,
		LoggerColorful:  cfg.LoggerColorful,
		LoggerLevel:     cfg.LoggerLevel,
		ConnMaxActive:   cfg.ConnMaxActive,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdle:     cfg.ConnMaxIdle,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	}
}

// ToPSQLConfig ...
func ToPSQLConfig(cfg *configs.Infrastructure_PSQL) *psqlpkg.Config {
	return &psqlpkg.Config{
		Dsn:             cfg.Dsn,
		SlowThreshold:   cfg.SlowThreshold,
		LoggerEnable:    cfg.LoggerEnable,
		LoggerColorful:  cfg.LoggerColorful,
		LoggerLevel:     cfg.LoggerLevel,
		ConnMaxActive:   cfg.ConnMaxActive,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdle:     cfg.ConnMaxIdle,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	}
}

// ToRedisConfig ...
func ToRedisConfig(cfg *configs.Infrastructure_Redis) *redispkg.Config {
	return &redispkg.Config{
		Addresses:       cfg.Addresses,
		Username:        cfg.Username,
		Password:        cfg.Password,
		Db:              cfg.Db,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ConnMaxActive:   cfg.ConnMaxActive,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdle:     cfg.ConnMaxIdle,
		ConnMinIdle:     cfg.ConnMinIdle,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	}
}

// ToConsulConfig ...
func ToConsulConfig(cfg *configs.Infrastructure_Consul) *consulpkg.Config {
	return &consulpkg.Config{
		Scheme:             cfg.Scheme,
		Address:            cfg.Address,
		PathPrefix:         cfg.PathPrefix,
		Datacenter:         cfg.Datacenter,
		WaitTime:           cfg.WaitTime,
		Token:              cfg.Token,
		Namespace:          cfg.Namespace,
		Partition:          cfg.Partition,
		WithHttpBasicAuth:  cfg.WithHttpBasicAuth,
		AuthUsername:       cfg.AuthUsername,
		AuthPassword:       cfg.AuthPassword,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		TlsAddress:         cfg.TlsAddress,
		TlsCaPem:           cfg.TlsCaPem,
		TlsCertPem:         cfg.TlsCertPem,
		TlsKeyPem:          cfg.TlsKeyPem,
	}
}

// ToJaegerTracerConfig ...
func ToJaegerTracerConfig(cfg *configs.Infrastructure_Jaeger) *jaegerpkg.Config {
	return &jaegerpkg.Config{
		Endpoint:          cfg.Endpoint,
		WithHttpBasicAuth: cfg.WithHttpBasicAuth,
		Username:          cfg.Username,
		Password:          cfg.Password,
	}
}
