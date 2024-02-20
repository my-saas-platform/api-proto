package setuputil

import (
	"context"
	"io"
	stdlog "log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	ippkg "github.com/ikaiguang/go-srv-kit/kit/ip"
	writerpkg "github.com/ikaiguang/go-srv-kit/kit/writer"
	authpkg "github.com/ikaiguang/go-srv-kit/kratos/auth"
	contextpkg "github.com/ikaiguang/go-srv-kit/kratos/context"
	headerpkg "github.com/ikaiguang/go-srv-kit/kratos/header"
	logpkg "github.com/ikaiguang/go-srv-kit/kratos/log"
	pkgerrors "github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

// Logger 日志处理示例
func (s *engines) Logger() (log.Logger, []io.Closer, error) {
	var err error
	s.loggerMutex.Do(func() {
		s.logger, s.loggerCloseFnSlice, err = s.loadingLogger()
	})
	if err != nil {
		s.loggerMutex = sync.Once{}
	}
	return s.logger, s.loggerCloseFnSlice, err
}

// LoggerHelper 日志处理示例
func (s *engines) LoggerHelper() (log.Logger, []io.Closer, error) {
	var err error
	s.loggerHelperMutex.Do(func() {
		s.loggerHelper, s.loggerHelperCloseFnSlice, err = s.loadingLoggerHelper()
	})
	if err != nil {
		s.loggerHelperMutex = sync.Once{}
	}
	return s.loggerHelper, s.loggerHelperCloseFnSlice, err
}

// LoggerMiddleware 中间件的日志处理示例
func (s *engines) LoggerMiddleware() (log.Logger, []io.Closer, error) {
	var err error
	s.loggerMiddlewareMutex.Do(func() {
		s.loggerMiddleware, s.loggerMiddlewareCloseFnSlice, err = s.loadingLoggerMiddleware()
	})
	if err != nil {
		s.loggerMiddlewareMutex = sync.Once{}
	}
	return s.loggerMiddleware, s.loggerMiddlewareCloseFnSlice, err
}

// loadingLogHelper 加载日志工具
func (s *engines) loadingLogHelper() (closeFnSlice []io.Closer, err error) {
	loggerInstance, closeFnSlice, err := s.LoggerHelper()
	if err != nil {
		return closeFnSlice, pkgerrors.WithStack(err)
	}
	if loggerInstance == nil {
		stdlog.Println("|*** 未加载日志工具")
		return closeFnSlice, err
	}

	// 日志
	if s.Config.EnableLoggingConsole() && s.LoggerConfigForConsole() != nil {
		stdlog.Println("|*** 加载：日志工具：日志输出到控制台")
	}
	if s.Config.EnableLoggingFile() && s.LoggerConfigForFile() != nil {
		stdlog.Println("|*** 加载：日志工具：日志输出到文件")
	}

	logpkg.Setup(loggerInstance)
	return closeFnSlice, err
}

// loadingLogger 初始化日志输出实例
func (s *engines) loadingLogger() (logger log.Logger, closeFnSlice []io.Closer, err error) {
	skip := logpkg.CallerSkipForLogger
	//return s.loadingLoggerWithCallerSkip(skip)
	logger, closeFnSlice, err = s.loadingLoggerWithCallerSkip(skip)
	if err != nil {
		return logger, closeFnSlice, err
	}
	logger = s.withLoggerPrefix(logger)
	return logger, closeFnSlice, err
}

// loadingLoggerHelper 初始化日志工具输出实例
func (s *engines) loadingLoggerHelper() (logger log.Logger, closeFnSlice []io.Closer, err error) {
	skip := logpkg.CallerSkipForHelper
	//return s.loadingLoggerWithCallerSkip(skip)
	logger, closeFnSlice, err = s.loadingLoggerWithCallerSkip(skip)
	if err != nil {
		return logger, closeFnSlice, err
	}
	logger = s.withLoggerPrefix(logger)
	return logger, closeFnSlice, err
}

// loadingLoggerMiddleware 初始化中间价的日志输出实例
func (s *engines) loadingLoggerMiddleware() (logger log.Logger, closeFnSlice []io.Closer, err error) {
	skip := logpkg.CallerSkipForMiddleware
	//return s.loadingLoggerWithCallerSkip(skip)
	logger, closeFnSlice, err = s.loadingLoggerWithCallerSkip(skip)
	if err != nil {
		return logger, closeFnSlice, err
	}
	logger = s.withLoggerPrefix(logger)
	return logger, closeFnSlice, err
}

// loadingLoggerWithCallerSkip 初始化日志输出实例
func (s *engines) loadingLoggerWithCallerSkip(skip int) (logger log.Logger, closeFnSlice []io.Closer, err error) {
	// loggers
	var loggers []log.Logger

	// DummyLogger
	stdLogger, err := logpkg.NewDummyLogger()
	if err != nil {
		return logger, closeFnSlice, err
	}

	// 日志 输出到控制台
	loggerConfigForConsole := s.LoggerConfigForConsole()
	if s.Config.EnableLoggingConsole() && loggerConfigForConsole != nil {
		stdLoggerConfig := &logpkg.ConfigStd{
			Level:      logpkg.ParseLevel(loggerConfigForConsole.Level),
			CallerSkip: skip,
		}
		stdLoggerImpl, err := logpkg.NewStdLogger(stdLoggerConfig)
		if err != nil {
			return logger, closeFnSlice, err
		}
		closeFnSlice = append(closeFnSlice, stdLoggerImpl)
		stdLogger = stdLoggerImpl
	}
	// 覆盖 stdLogger
	loggers = append(loggers, stdLogger)

	// 日志 输出到文件
	loggerConfigForFile := s.LoggerConfigForFile()
	if s.Config.EnableLoggingFile() && loggerConfigForFile != nil {
		// file logger
		fileLoggerConfig := &logpkg.ConfigFile{
			Level:      logpkg.ParseLevel(loggerConfigForFile.Level),
			CallerSkip: skip,

			Dir:      loggerConfigForFile.Dir,
			Filename: loggerConfigForFile.Filename,

			RotateTime: loggerConfigForFile.RotateTime.AsDuration(),
			RotateSize: loggerConfigForFile.RotateSize,

			StorageCounter: uint(loggerConfigForFile.StorageCounter),
			StorageAge:     loggerConfigForFile.StorageAge.AsDuration(),
		}
		writer, err := s.getLoggerFileWriter()
		if err != nil {
			return logger, closeFnSlice, err
		}
		fileLogger, err := logpkg.NewFileLogger(
			fileLoggerConfig,
			logpkg.WithWriter(writer),
		)
		closeFnSlice = append(closeFnSlice, fileLogger)
		if err != nil {
			return logger, closeFnSlice, err
		}
		loggers = append(loggers, fileLogger)
	}

	// 日志工具
	if len(loggers) == 0 {
		return logger, closeFnSlice, err
	}
	return logpkg.NewMultiLogger(loggers...), closeFnSlice, err
}

// SetLoggerPrefixField .
func (s *engines) SetLoggerPrefixField() *LoggerPrefixField {
	s.loggerPrefixFieldMutex.Do(func() {
		s.loggerPrefixField = s.setLoggerPrefixField()
	})
	return s.loggerPrefixField
}

// getLoggerFileWriter 文件日志写手柄
func (s *engines) getLoggerFileWriter() (io.Writer, error) {
	if s.loggerFileWriter != nil {
		return s.loggerFileWriter, nil
	}
	var err error
	s.loggerFileWriterMutex.Do(func() {
		s.loggerFileWriter, err = s.loadingLoggerFileWriter()
	})
	if err != nil {
		s.loggerFileWriterMutex = sync.Once{}
	}
	return s.loggerFileWriter, err
}

// loadingLoggerFileWriter 启动日志文件写手柄
func (s *engines) loadingLoggerFileWriter() (io.Writer, error) {
	fileLoggerConfig := s.Config.LoggerConfigForFile()
	if !s.Config.EnableLoggingFile() || fileLoggerConfig == nil {
		stdlog.Println("|*** 加载：日志工具：虚拟的文件写手柄")
		return writerpkg.NewDummyWriter()
	}
	rotateConfig := &writerpkg.ConfigRotate{
		Dir:            fileLoggerConfig.Dir,
		Filename:       fileLoggerConfig.Filename,
		RotateTime:     fileLoggerConfig.RotateTime.AsDuration(),
		RotateSize:     fileLoggerConfig.RotateSize,
		StorageCounter: uint(fileLoggerConfig.StorageCounter),
		StorageAge:     fileLoggerConfig.StorageAge.AsDuration(),
	}
	if appConfig := s.Config.AppConfig(); appConfig != nil {
		replaceHandler := strings.NewReplacer(
			" ", "-",
			"/", "--",
		)
		if appConfig.ServerEnv != "" {
			rotateConfig.Filename += "_" + replaceHandler.Replace(appConfig.ServerEnv)
		}
		if appConfig.ServerVersion != "" {
			rotateConfig.Filename += "_" + replaceHandler.Replace(appConfig.ServerVersion)
		}
	}
	return writerpkg.NewRotateFile(rotateConfig)
}

// setLoggerPrefixField 组装日志前缀
func (s *engines) setLoggerPrefixField() *LoggerPrefixField {
	appConfig := s.AppConfig()
	if appConfig == nil {
		return &LoggerPrefixField{
			ServerIP: ippkg.LocalIP(),
		}
	}

	fields := &LoggerPrefixField{
		AppName:    appConfig.ServerName,
		AppVersion: appConfig.ServerVersion,
		AppEnv:     appConfig.ServerEnv,
		ServerIP:   ippkg.LocalIP(),
	}
	fields.Hostname, _ = os.Hostname()
	return fields
}

// withLoggerPrefix ...
func (s *engines) withLoggerPrefix(logger log.Logger) log.Logger {
	//var kvs = []interface{}{
	//	"app",
	//	s.SetLoggerPrefixField().String(),
	//}
	var kvs = s.SetLoggerPrefixField().Prefix()
	kvs = append(kvs, "tracer", s.withLoggerTracer())
	return log.With(logger, kvs...)
}

// withLoggerTracer returns a traceid valuer.
func (s *engines) withLoggerTracer() log.Valuer {
	return func(ctx context.Context) interface{} {
		var (
			res string
		)
		span := trace.SpanContextFromContext(ctx)
		switch {
		case span.HasTraceID():
			res += `traceId="` + span.TraceID().String() + `"`
		default:
			if httpTr, ok := contextpkg.MatchHTTPServerContext(ctx); ok {
				if traceId := httpTr.RequestHeader().Get(headerpkg.RequestID); traceId != "" {
					res += `traceId="` + traceId + `"`
				}
			} else if grpcTr, ok := contextpkg.MatchGRPCServerContext(ctx); ok {
				if traceId := grpcTr.RequestHeader().Get(headerpkg.RequestID); traceId != "" {
					res += `traceId="` + traceId + `"`
				}
			}
		}
		if span.HasSpanID() {
			res += ` spanId="` + span.SpanID().String() + `"`
		}
		if claims, ok := authpkg.GetAuthClaimsFromContext(ctx); ok && claims.Payload != nil {
			if claims.Payload.UserID > 0 {
				res += ` userId="` + strconv.FormatUint(claims.Payload.UserID, 10) + `"`
			} else if claims.Payload.UserUuid != "" {
				res += ` userUuid="` + claims.Payload.UserUuid + `"`
			}
		}
		return res
	}
}
