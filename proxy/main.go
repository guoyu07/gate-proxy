package main

import (
	"flag"
	"goodsogood/gateway"
	"goodsogood/gateway/proxy/global"
	"goodsogood/gateway/proxy/handle"
	"goodsogood/gateway/proxy/plugin/auth"
	"goodsogood/gateway/proxy/plugin/snapshot"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/buntdb"
)

var (
	authHost   = flag.String("auth", "", "Auth Host Port Example:127.0.0.1:9090")
	assetsPath = flag.String("static", "./dashboard/assets", "Dashboard assets path Example: ./dashboard/assets")
	indexPath  = flag.String("index", "./dashboard/index.html", "Dashboard index.html path Example: ./dashboard/index.html")
	dbPath     = flag.String("db", "./gateway.db", "Gateway local DB path Example: ./gateway.db")
)

func init() {
	flag.Parse()
	if *authHost == "" {
		log.Fatal("Auth Host Port Empty.")
	}
}

func main() {
	db, err := buntdb.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	gin.SetMode(gin.ReleaseMode)
	global.Store.SetDB(db)
	engine := gateway.New()
	// 初始化授权插件
	authPlugin, err := auth.NewAuth(auth.Options{
		Servers: []string{*authHost},
	})
	if err != nil {
		log.Fatal(err)
	}
	engine.RegisterPlugin(authPlugin)
	// 注册快照插件
	engine.RegisterPlugin(snapshot.Snapshot{})
	global.Store.SetProxy(engine)
	// 加载配置
	global.Store.LoadCache()
	router := gin.New()
	router.Static("/assets", *assetsPath)
	router.LoadHTMLFiles(*indexPath)
	router.GET("/", handle.Index)
	router.NoRoute(handle.Index)
	api := router.Group("/v1")
	// 获取网关状态
	api.GET("/status", handle.Status)
	// 获取所有的集群
	api.GET("/clusters", handle.Clusters)
	// 获取所有的接口
	api.GET("/apis", handle.Apis)
	// 插件列表
	api.GET("/plugins", handle.Plugins)
	// 增加集群
	api.POST("/cluster", handle.AddCluster)
	// 删除集群
	api.POST("/cluster/delete", handle.DelCluster)
	// 更新集群
	api.POST("/cluster/update", handle.UpdateCluster)
	// 获取指定集群的后端服务
	api.GET("/backends/:clusterName", handle.Backends)
	// 增加后端服务
	api.POST("/backend", handle.AddBackend)
	// 移除后端服务
	api.POST("/backend/delete", handle.DelBackend)
	// 更新后端服务
	api.POST("/backend/update", handle.UpdateBackend)
	// 增加路由规则
	api.POST("/api", handle.AddAPI)
	// 更新路由规则
	api.POST("/api/update", handle.UpdateAPI)
	// 删除路由规则
	api.POST("/api/delete", handle.DeleteAPI)
	go router.Run(":8081")
	go engine.Run(":80")
	engine.RunTLS("", "./cert/_.goodsogood.com.pem", "./cert/_.goodsogood.com.key")
}
