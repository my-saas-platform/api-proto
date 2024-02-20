package setuputil

import (
	stdlog "log"
	"sync"

	consulapi "github.com/hashicorp/consul/api"
	consulpkg "github.com/ikaiguang/go-srv-kit/data/consul"
	gormpkg "github.com/ikaiguang/go-srv-kit/data/gorm"
	jaegerpkg "github.com/ikaiguang/go-srv-kit/data/jaeger"
	mysqlpkg "github.com/ikaiguang/go-srv-kit/data/mysql"
	psqlpkg "github.com/ikaiguang/go-srv-kit/data/postgres"
	redispkg "github.com/ikaiguang/go-srv-kit/data/redis"
	middlewarepkg "github.com/ikaiguang/go-srv-kit/kratos/middleware"
	apputil "github.com/my-saas-platform/saas-api-proto/util/app"
	pkgerrors "github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetMySQLGormDB 数据库
func (s *engines) GetMySQLGormDB() (*gorm.DB, error) {
	if s.mysqlGormDB != nil {
		return s.mysqlGormDB, nil
	}
	var err error
	s.mysqlGormMutex.Do(func() {
		s.mysqlGormDB, err = s.loadingMysqlGormDB()
	})
	if err != nil {
		s.mysqlGormMutex = sync.Once{}
	}
	return s.mysqlGormDB, err
}

// reloadMysqlGormDB 重新加载 mysql gorm 数据库
func (s *engines) reloadMysqlGormDB() error {
	if s.Config.MySQLConfig() == nil {
		return nil
	}
	dbConn, err := s.loadingMysqlGormDB()
	if err != nil {
		return err
	}
	*s.mysqlGormDB = *dbConn
	return nil
}

// loadingMysqlGormDB mysql gorm 数据库
func (s *engines) loadingMysqlGormDB() (*gorm.DB, error) {
	if s.Config.MySQLConfig() == nil {
		stdlog.Println("|*** 加载：MySQL-GORM：未初始化")
		return nil, pkgerrors.WithStack(ErrUninitialized)
	}
	stdlog.Println("|*** 加载：MySQL-GORM：...")

	// logger writer
	var (
		writers []logger.Writer
		opts    []gormpkg.Option
	)
	if s.Config.EnableLoggingConsole() {
		writers = append(writers, gormpkg.NewStdWriter())
	}
	if s.Config.EnableLoggingFile() {
		writer, err := s.getLoggerFileWriter()
		if err != nil {
			return nil, err
		}
		writers = append(writers, gormpkg.NewJSONWriter(writer))
	}
	if len(writers) > 0 {
		opts = append(opts, gormpkg.WithWriters(writers...))
	}
	return mysqlpkg.NewMysqlDB(ToMysqlConfig(s.Config.MySQLConfig()), opts...)
}

// GetPostgresGormDB 数据库
func (s *engines) GetPostgresGormDB() (*gorm.DB, error) {
	if s.postgresGormDB != nil {
		return s.postgresGormDB, nil
	}
	var err error
	s.postgresGormMutex.Do(func() {
		s.postgresGormDB, err = s.loadingPostgresGormDB()
	})
	if err != nil {
		s.postgresGormMutex = sync.Once{}
	}
	return s.postgresGormDB, err
}

// reloadPostgresGormDB 重新加载 postgres gorm 数据库
func (s *engines) reloadPostgresGormDB() error {
	if s.Config.PostgresConfig() == nil {
		return nil
	}
	dbConn, err := s.loadingPostgresGormDB()
	if err != nil {
		return err
	}
	*s.postgresGormDB = *dbConn
	return nil
}

// loadingPostgresGormDB postgres gorm 数据库
func (s *engines) loadingPostgresGormDB() (*gorm.DB, error) {
	if s.Config.PostgresConfig() == nil {
		stdlog.Println("|*** 加载：Postgres-GORM：未初始化")
		return nil, pkgerrors.WithStack(ErrUninitialized)
	}
	stdlog.Println("|*** 加载：Postgres-GORM：...")

	// logger writer
	var (
		writers []logger.Writer
		opts    []gormpkg.Option
	)
	if s.Config.EnableLoggingConsole() {
		writers = append(writers, gormpkg.NewStdWriter())
	}
	if s.Config.EnableLoggingFile() {
		writer, err := s.getLoggerFileWriter()
		if err != nil {
			return nil, err
		}
		writers = append(writers, gormpkg.NewJSONWriter(writer))
	}
	if len(writers) > 0 {
		opts = append(opts, gormpkg.WithWriters(writers...))
	}
	return psqlpkg.NewDB(ToPSQLConfig(s.Config.PostgresConfig()), opts...)
}

// GetRedisClient redis 客户端
func (s *engines) GetRedisClient() (redis.UniversalClient, error) {
	if s.redisClient != nil {
		return s.redisClient, nil
	}
	var err error
	s.redisClientMutex.Do(func() {
		s.redisClient, err = s.loadingRedisClient()
	})
	if err != nil {
		s.redisClientMutex = sync.Once{}
	}
	return s.redisClient, err
}

// reloadRedisClient 重新加载 redis 客户端
func (s *engines) reloadRedisClient() error {
	if s.Config.RedisConfig() == nil {
		return nil
	}
	redisClient, err := s.loadingRedisClient()
	if err != nil {
		return err
	}
	s.redisClient = redisClient
	return nil
}

// loadingRedisClient redis 客户端
func (s *engines) loadingRedisClient() (redis.UniversalClient, error) {
	if s.Config.RedisConfig() == nil {
		stdlog.Println("|*** 加载：Redis客户端：未初始化")
		return nil, pkgerrors.WithStack(ErrUninitialized)
	}
	stdlog.Println("|*** 加载：Redis客户端：...")

	return redispkg.NewDB(ToRedisConfig(s.Config.RedisConfig()))
}

// GetConsulClient consul 客户端
func (s *engines) GetConsulClient() (*consulapi.Client, error) {
	var err error
	s.consulClientMutex.Do(func() {
		s.consulClient, err = s.loadingConsulClient()
	})
	if err != nil {
		s.consulClientMutex = sync.Once{}
	}
	return s.consulClient, err
}

// SetConsulClient consul 客户端
//func (s *engines) SetConsulClient(cc *api.Client) {
//	if cc == nil {
//		return
//	}
//	var hasSet bool
//	s.consulClientMutex.Do(func() {
//		hasSet = true
//		s.consulClient = cc
//	})
//	if !hasSet {
//		s.consulClient = cc
//	}
//}

// loadingConsulClient consul 客户端
func (s *engines) loadingConsulClient() (*consulapi.Client, error) {
	if s.Config.ConsulConfig() == nil {
		stdlog.Println("|*** 加载：Consul客户端：未初始化")
		return nil, pkgerrors.WithStack(ErrUninitialized)
	}
	stdlog.Println("|*** 加载：Consul客户端：...")

	return consulpkg.NewConsulClient(ToConsulConfig(s.Config.ConsulConfig()))
}

// GetJaegerExporter ...
func (s *engines) GetJaegerExporter() (*jaeger.Exporter, error) {
	if s.jaegerTraceExporter != nil {
		return s.jaegerTraceExporter, nil
	}
	var err error
	s.jaegerTraceExporterMutex.Do(func() {
		s.jaegerTraceExporter, err = s.loadingJaegerTraceExporter()
	})
	if err != nil {
		s.jaegerTraceExporterMutex = sync.Once{}
	}
	return s.jaegerTraceExporter, err
}

// loadingJaegerTraceExporter jaegerTrace
func (s *engines) loadingJaegerTraceExporter() (*jaeger.Exporter, error) {
	if s.Config.JaegerConfig() == nil {
		stdlog.Println("|*** 加载：JaegerTrace：未初始化")
		return nil, pkgerrors.WithStack(ErrUninitialized)
	}
	stdlog.Println("|*** 加载：JaegerTrace：...")

	return jaegerpkg.NewJaegerExporter(ToJaegerTracerConfig(s.Config.JaegerConfig()))
}

// InitTracerProvider trace provider
func (s *engines) InitTracerProvider() error {
	// 未开启
	//settingConfig := s.SettingConfig()
	//if settingConfig == nil || !settingConfig.EnableServiceTracer {
	//	return nil
	//}

	stdlog.Println("|*** 加载：服务追踪：Tracer")
	// Create the Jaeger exporter
	var opts []middlewarepkg.TracerOption
	if cfg := s.JaegerConfig(); cfg != nil && cfg.Enable {
		exp, err := s.GetJaegerExporter()
		if err != nil {
			return err
		}
		opts = append(opts, middlewarepkg.WithTracerJaegerExporter(exp))
	}

	appConfig := s.AppConfig()

	return middlewarepkg.SetTracer(apputil.ID(appConfig), opts...)
}

// loadingSnowflakeWorker 加载雪花算法
func (s *engines) loadingSnowflakeWorker() error {
	//workerConfig := s.SnowflakeWorkerConfig()
	//if workerConfig == nil {
	//	stdlog.Println("|*** 加载：雪花算法：未初始化")
	//	return pkgerrors.WithStack(ErrUninitialized)
	//}
	//stdlog.Printf("|*** 加载：雪花算法")
	//
	//// http 选项
	//logger, _, err := s.LoggerMiddleware()
	//if err != nil {
	//	return err
	//}
	//var httpOptions = []http.ClientOption{
	//	http.WithMiddleware(
	//		recovery.Recovery(),
	//		metadata.Client(),
	//		tracing.Client(),
	//		apppkg.ClientLog(logger),
	//	),
	//	http.WithResponseDecoder(apppkg.ResponseDecoder),
	//	http.WithEndpoint(workerConfig.Endpoint),
	//}
	//if workerConfig.WithDiscovery {
	//	consulClient, err := s.GetConsulClient()
	//	if err != nil {
	//		return err
	//	}
	//	r := consul.New(consulClient)
	//	httpOptions = append(httpOptions, http.WithDiscovery(r))
	//}
	//
	//// http 链接
	//httpConn, err := clientpkg.NewHTTPClient(context.Background(), httpOptions...)
	//if err != nil {
	//	return pkgerrors.WithStack(err)
	//}
	//httpClient := servicev1.NewSrvWorkerHTTPClient(httpConn)
	//
	//// 雪花算法ID
	//appConfig := s.AppConfig()
	//workerReq := &apiv1.GetNodeIdReq{
	//	InstanceId:   apppkg.ID(appConfig),
	//	InstanceName: appConfig.Name,
	//	Endpoints:    appConfig.Endpoints,
	//	Metadata:     appConfig.Metadata,
	//}
	//workerResp, err := httpClient.GetNodeId(context.Background(), workerReq)
	//if err != nil {
	//	return pkgerrors.WithStack(err)
	//}
	//
	//// 雪花算法
	//stdlog.Printf("|*** 加载：雪花算法：nodeId = %d", workerResp.NodeId)
	//snowflakeNode, err := snowflake.NewNode(workerResp.NodeId)
	//if err != nil {
	//	return pkgerrors.WithStack(err)
	//}
	//idpkg.SetNode(snowflakeNode)
	//
	//// 续期
	//extendReq := &apiv1.ExtendNodeIdReq{
	//	Id:         workerResp.Id,
	//	InstanceId: workerReq.InstanceId,
	//	NodeId:     workerResp.NodeId,
	//}
	//s.snowflakeStopChannel = make(chan int)
	//go func() {
	//	ticker := time.NewTicker(time.Second * 3)
	//	defer ticker.Stop()
	//	for {
	//		select {
	//		case <-ticker.C:
	//			//debugpkg.Println("雪花算法：续期：", extendReq.String())
	//			_, _ = httpClient.ExtendNodeId(context.Background(), extendReq)
	//		case <-s.snowflakeStopChannel:
	//			return
	//		}
	//	}
	//}()

	return nil
}
