package middlewareutil

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	authpkg "github.com/ikaiguang/go-srv-kit/kratos/auth"
	setuputil "github.com/my-saas-platform/saas-api-proto/util/setup"
)

// TransportServiceKind 通行类型
type TransportServiceKind int32

const (
	TransportServiceKindALL  = 0
	TransportServiceKindHTTP = 1
	TransportServiceKindGRPC = 2
)

func (s TransportServiceKind) MatchServiceKind(ctx context.Context) bool {
	switch s {
	default:
		return true
	case TransportServiceKindALL:
		return true
	case TransportServiceKindHTTP:
		tr, ok := transport.FromServerContext(ctx)
		return ok && tr.Kind() == transport.KindHTTP
	case TransportServiceKindGRPC:
		tr, ok := transport.FromServerContext(ctx)
		return ok && tr.Kind() == transport.KindGRPC
	}
	//return false
}

// NewWhiteListMatcher 路由白名单
func NewWhiteListMatcher(whiteList map[string]TransportServiceKind) selector.MatchFunc {
	return func(ctx context.Context, operation string) bool {
		//if tr, ok := contextutil.MatchHTTPServerContext(ctx); ok {
		//	if _, ok := whiteList[tr.Request().URL.Path]; ok {
		//		return false
		//	}
		//}

		tsk, ok := whiteList[operation]
		if !ok {
			return true
		}
		if tsk.MatchServiceKind(ctx) {
			return false
		}
		return true
	}
}

// NewJWTMiddleware jwt中间
func NewJWTMiddleware(engineHandler setuputil.Engine, whiteList map[string]TransportServiceKind) (m middleware.Middleware, err error) {
	// redis
	redisCC, err := engineHandler.GetRedisClient()
	if err != nil {
		return m, err
	}
	authTokenRepo, err := engineHandler.GetAuthTokenRepo(redisCC)
	if err != nil {
		return m, err
	}

	m = selector.Server(
		authpkg.Server(
			authTokenRepo.JWTSigningKeyFunc,
			authpkg.WithSigningMethod(authTokenRepo.JWTSigningMethod()),
			authpkg.WithClaims(authTokenRepo.JWTSigningClaims),
			authpkg.WithAccessTokenValidator(authTokenRepo.VerifyAccessToken),
		),
	).
		Match(NewWhiteListMatcher(whiteList)).
		Build()

	return m, err
}
