package setuputil

import (
	stdlog "log"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	pkgerrors "github.com/pkg/errors"
)

// watchConfigApp 监听配置 app
func (s *engines) watchConfigApp() (err error) {
	stdlog.Println("|*** 加载：监听配置：App")
	var observer = func(k string, v config.Value) {
		_ = s.logger.Log(log.LevelInfo,
			"watch config.App",
			"监听配置：数据有改动：key = "+k,
		)

		// app
		appConfig := s.AppConfig()
		if err := v.Scan(appConfig); err != nil {
			_ = s.logger.Log(log.LevelError,
				"watch config.App",
				"config.Value.Scan(appConfig) err : "+err.Error(),
			)
		}
	}
	if err = s.Watch("app", observer); err != nil {
		return pkgerrors.WithStack(err)
	}
	return
}

// watchConfigData 监听配置 data
func (s *engines) watchConfigData() (err error) {
	if s.Config.InfrastructureConfig() == nil {
		err = pkgerrors.New("[请配置服务再启动] config key : data")
		return err
	}

	stdlog.Println("|*** 加载：监听配置：Infrastructure")
	var observer = func(k string, v config.Value) {
		_ = s.logger.Log(log.LevelInfo,
			"watch config.Infrastructure",
			"监听配置：数据有改动：key = "+k,
		)

		// app
		infrastructureConfig := s.InfrastructureConfig()
		if err := v.Scan(infrastructureConfig); err != nil {
			_ = s.logger.Log(log.LevelError,
				"watch config.Infrastructure",
				"config.Value.Scan(infrastructureConfig) err : "+err.Error(),
			)
		}

		// 重新加载 mysql
		//if err := s.reloadMysqlGormDB(); err != nil {
		//	_ = s.logger.Log(log.LevelError,
		//		"watchConfigData",
		//		"reloadMysqlGormDB failed : "+err.Error(),
		//	)
		//}

		// 重新加载 postgres
		//if err := s.reloadPostgresGormDB(); err != nil {
		//	_ = s.logger.Log(log.LevelError,
		//		"watchConfigData",
		//		"reloadPostgresGormDB failed : "+err.Error(),
		//	)
		//}

		// 重新加载 redis
		//if err := s.reloadRedisClient(); err != nil {
		//	_ = s.logger.Log(log.LevelError,
		//		"watchConfigData",
		//		"reloadRedisClient failed : "+err.Error(),
		//	)
		//}
	}
	if err = s.Watch("infrastructure", observer); err != nil {
		return pkgerrors.WithStack(err)
	}
	return
}
