package auth

import (
	"fmt"
	"goodsogood/gateway"
	"goodsogood/mall/lib/errors"
	gsthrift "goodsogood/thrift"
	tokenService "goodsogood/thrift/protocol/token/service"
	"goodsogood/thrift/protocol/types"
	"net/http"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
)

type (
	AuthPlugin struct {
		maxIdle     uint32
		maxConn     uint32
		servers     []string
		connTimeout time.Duration
		readTimeout time.Duration
		pool        *gsthrift.ChannelClientPool
	}
	Options struct {
		MaxIdle     int
		MaxConn     int
		Servers     []string
		ConnTimeout int64
		ReadTimeout int64
	}
)

var (
	// DefaultMaxIdle . 默认空闲链接
	DefaultMaxIdle = 50
	// DefaultMaxConn . 默认最大链接
	DefaultMaxConn = 80
	// DefaultConnTimeout . 链接超时
	DefaultConnTimeout int64 = 3
	// DefaultReadTimeout . 读取超时
	DefaultReadTimeout int64 = 5
)

// NewAuth .
func NewAuth(options Options) (*AuthPlugin, error) {
	if len(options.Servers) < 1 {
		return nil, errors.New(-1, "Auth Server empty")
	}
	authPlugin := &AuthPlugin{
		servers: options.Servers,
	}
	if options.MaxIdle < 1 {
		options.MaxIdle = DefaultMaxIdle
	}
	authPlugin.maxIdle = uint32(options.MaxIdle)
	if options.MaxConn < 1 {
		options.MaxConn = DefaultMaxConn
	}
	authPlugin.maxConn = uint32(options.MaxConn)
	if options.ConnTimeout < 0 {
		options.ConnTimeout = DefaultConnTimeout
	}
	authPlugin.connTimeout = time.Duration(options.ConnTimeout) * time.Second
	if options.ReadTimeout < 0 {
		options.ReadTimeout = DefaultReadTimeout
	}
	authPlugin.readTimeout = time.Duration(options.ReadTimeout) * time.Second
	authPlugin.pool = gsthrift.NewChannelClientPool(
		authPlugin.maxIdle,
		authPlugin.maxConn,
		authPlugin.servers,
		authPlugin.connTimeout,
		authPlugin.readTimeout,
		func(openedSocket thrift.TTransport) gsthrift.Client {
			protocol := thrift.NewTCompactProtocol(openedSocket)
			return tokenService.NewTokenServiceClientProtocol(openedSocket, protocol, protocol)
		},
		func(client gsthrift.Client) bool {
			pipe, ok := client.(*tokenService.TokenServiceClient)
			if !ok {
				return false
			}
			state, err := pipe.Ping()
			if err != nil {
				return false
			}
			return state.GetStatus()
		},
	)
	return authPlugin, nil
}

func (auth *AuthPlugin) Name() string {
	return "auth"
}

func (auth *AuthPlugin) Private() bool {
	return false
}

func (auth *AuthPlugin) Version() string {
	return "0.1"
}

// 鉴权处理
func (authPlugin *AuthPlugin) Handle(ctx *gateway.Context) {
	var (
		token, userId, platformDeviceType, platformDeviceInfo string
		has                                                   bool
	)
	if has, token = authPlugin.getVal(ctx, "token"); !has {
		ctx.Render(http.StatusOK, gateway.TokenEmpty)
		ctx.Abort()
		return
	}
	if has, userId = authPlugin.getVal(ctx, "userId"); !has {
		ctx.Render(http.StatusOK, gateway.UserIDEmpty)
		ctx.Abort()
		return
	}
	if has, platformDeviceType = authPlugin.getVal(ctx, "platformDeviceType"); !has {
		ctx.Render(http.StatusOK, gateway.DeviceTypeEmpty)
		ctx.Abort()
		return
	}
	if has, platformDeviceInfo = authPlugin.getVal(ctx, "platformDeviceInfo"); !has {
		ctx.Render(http.StatusOK, gateway.DeviceInfoEmpty)
		ctx.Abort()
		return
	}
	pooledClient, err := authPlugin.pool.Get()
	if err != nil {
		fmt.Println(err)
		ctx.Render(http.StatusOK, gateway.TokenServiceConnectFailed)
		ctx.Abort()
		return
	}
	defer pooledClient.Close()
	client := pooledClient.RawClient().(*tokenService.TokenServiceClient)
	res, err := client.CheckAccessToken(&types.RID{
		RId: userId + time.Now().Format("20060102150405"),
	}, userId, platformDeviceType, platformDeviceInfo, token)
	if err != nil {
		ctx.Render(http.StatusOK, errors.New(-1, err.Error()))
		ctx.Abort()
		return
	}
	if res.GetCode() != 0 {
		ctx.Render(http.StatusOK, errors.New(int(res.GetCode()), res.GetMessage()))
		ctx.Abort()
		return
	}
	ctx.Next()
}

func (auth *AuthPlugin) getVal(ctx *gateway.Context, key string) (has bool, val string) {
	val = ctx.Query(key)
	if val != "" {
		return true, val
	}
	val = ctx.Request.Header.Get(key)
	if val != "" {
		return true, val
	}
	return
}
