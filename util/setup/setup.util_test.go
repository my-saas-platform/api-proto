package setuputil

import (
	"context"
	"testing"

	debugpkg "github.com/ikaiguang/go-srv-kit/debug"
	errorpkg "github.com/ikaiguang/go-srv-kit/kratos/error"
	logpkg "github.com/ikaiguang/go-srv-kit/kratos/log"
	"github.com/stretchr/testify/require"
)

// go test -v ./business-kit/setup/ -count=1 -test.run=TestSetup
// 或使用 setup_main.pkg_test.go TestMain 配置的 testdata/configs
// go test -v ./pkg/setup/ -count=1 -test.run=TestSetup
func TestSetup(t *testing.T) {
	//engineHandler, err := New(WithConfigPath(confPath))
	p := "./../../testdata/configs"
	engineHandler, err := New(
		WithConfigPath(p),
	)
	if err != nil {
		t.Errorf("%+v\n", err)
		t.FailNow()
	}
	defer func() { _ = engineHandler.Close() }()

	ctx := context.Background()

	// env
	logpkg.Infof("testing : app env is %v", engineHandler.RuntimeEnv())
	debugpkg.Println("testing : print message by debugpkg.Println")
	logpkg.Info("testing : print message by logpkg.Info")
	logpkg.Errorw("testing: print message by logpkg.Errorw")
	e := errorpkg.BadRequest(errorpkg.ERROR_BAD_REQUEST.String(), "testdata")
	err = errorpkg.WithStack(e)
	logpkg.Errorf("%+v", err)

	// db
	db, err := engineHandler.GetMySQLGormDB()
	require.Nil(t, err)
	require.NotNil(t, db)
	type DBRes struct {
		DBName string `gorm:"column:db_name"`
	}
	var dbRes DBRes
	err = db.WithContext(ctx).Raw("SELECT DATABASE() AS db_name").Scan(&dbRes).Error
	require.Nil(t, err)
	t.Logf("db res : %+v\n", dbRes)

	// redis
	redisCC, err := engineHandler.GetRedisClient()
	require.Nil(t, err)
	require.NotNil(t, redisCC)
	redisKey := "test-foo"
	redisValue := "test-bar"
	err = redisCC.Set(ctx, redisKey, redisValue, 0).Err()
	require.Nil(t, err)
	redisGotValue, err := redisCC.Get(ctx, redisKey).Result()
	require.Nil(t, err)
	require.Equal(t, redisValue, redisGotValue)
	t.Logf("redis res : %+v\n", redisGotValue)
}
