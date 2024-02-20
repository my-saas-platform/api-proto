package setuputil

import (
	"testing"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/stretchr/testify/require"
)

// go test -v ./pkg/setup/ -count=1 -test.run=TestNewConfiguration
func TestNewConfiguration(t *testing.T) {
	var opts []config.Option
	opts = append(opts, config.WithSource(
		file.NewSource(confPath),
	))
	handler, err := NewConfiguration(opts...)
	if err != nil {
		t.Errorf("%+v\n", err)
		t.FailNow()
	}

	t.Log("*** | env：", handler.RuntimeEnv())
	t.Logf("*** | AppConfig：%+v\n", handler.AppConfig())
}

// go test -v ./pkg/setup/ -count=1 -test.run=TestNew_newConfigWithConsul -conf-consul=./../../app/admin-service/configs/consul
func TestNew_newConfigWithConsul(t *testing.T) {
	var opts = &options{}

	// 在 初始化Consul配置中心 结束 前没有错误即为测试成功
	_, _, _ = newConfigWithConsul(opts)
	//if err != nil {
	//t.Logf("%+v\n", err)
	//t.FailNow()
	//}

	//t.Log("*** | env：", handler.RuntimeEnv())
	//t.Logf("*** | AppConfig：%+v\n", handler.AppConfig())
}

// go test -v ./pkg/setup/ -count=1 -test.run=TestNewUpPackages
func TestNewUpPackages(t *testing.T) {
	// config
	var opts []config.Option
	opts = append(opts, config.WithSource(
		file.NewSource(confPath),
	))
	configHandler, err := NewConfiguration(opts...)
	if err != nil {
		t.Errorf("%+v\n", err)
		t.FailNow()
	}
	t.Log("*** | env：", configHandler.RuntimeEnv())

	// up
	upHandler := initEngine(configHandler)

	// db
	db, err := upHandler.GetMySQLGormDB()
	require.Nil(t, err)
	require.NotNil(t, db)

	// redis
	redisCC, err := upHandler.GetRedisClient()
	require.Nil(t, err)
	require.NotNil(t, redisCC)
}
