package setuputil

import (
	stdlog "log"
	"sync"

	authpkg "github.com/ikaiguang/go-srv-kit/kratos/auth"
	errorpkg "github.com/ikaiguang/go-srv-kit/kratos/error"
	pkgerrors "github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// GetAuthTokenRepo 验证Token工具
func (s *engines) GetAuthTokenRepo(redisCC redis.UniversalClient) (authpkg.AuthRepo, error) {
	if s.authTokenRepo != nil {
		return s.authTokenRepo, nil
	}
	var err error
	s.authTokenRepoMutex.Do(func() {
		s.authTokenRepo, err = s.loadingAuthTokenRepo(redisCC)
	})
	if err != nil {
		s.authTokenRepoMutex = sync.Once{}
	}
	return s.authTokenRepo, err
}

// loadingAuthTokenRepo 验证Token工具
func (s *engines) loadingAuthTokenRepo(redisCC redis.UniversalClient) (authpkg.AuthRepo, error) {
	tokenConfig := s.TokenEncryptConfig()
	if tokenConfig == nil {
		err := pkgerrors.New("[请配置服务再启动] config key : setting.encrypt_secret.token_encrypt")
		return nil, err
	}
	if tokenConfig.GetSignKey() == "" {
		err := pkgerrors.New("[请配置服务再启动] config key : setting.encrypt_secret.token_encrypt.sign_key")
		return nil, err
	}
	if tokenConfig.RefreshKey == "" {
		err := pkgerrors.New("[请配置服务再启动] config key : setting.encrypt_secret.token_encrypt.refresh_key")
		return nil, err
	}
	logger, _, err := s.LoggerMiddleware()
	if err != nil {
		return nil, err
	}
	tokenManger := authpkg.NewTokenManger(logger, redisCC, authpkg.CheckAuthCacheKeyPrefix(nil))
	config := &authpkg.Config{
		SignCrypto:    authpkg.NewSignEncryptor(tokenConfig.SignKey),
		RefreshCrypto: authpkg.NewCBCCipher(tokenConfig.RefreshKey),
	}
	stdlog.Println("|*** 加载：验证Token工具：...")
	authRepo, err := authpkg.NewAuthRepo(*config, logger, tokenManger)
	if err != nil {
		e := errorpkg.FromError(err)
		return nil, errorpkg.WithStack(e)
	}
	return authRepo, nil
}
