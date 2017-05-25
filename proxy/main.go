package main

import (
    "goodsogood/gateway"
    "goodsogood/gateway/proxy/plugin/auth"
    "goodsogood/gateway/proxy/plugin/snapshot"
    "github.com/gin-gonic/gin"
    "github.com/tidwall/buntdb"
    "log"
    "goodsogood/gateway/proxy/global"
    "goodsogood/gateway/proxy/handle"
)

func main() {
    db, err := buntdb.Open("gateway.db")
    if err != nil {
        log.Fatal(err)
    }
    global.Store.SetDB(db)
    engine := gateway.New()
    // 注册插件
    engine.RegisterPlugin(auth.Auth{})
    engine.RegisterPlugin(snapshot.Snapshot{})
    global.Store.SetProxy(engine)
    global.Store.LoadCache()
    router := gin.New()
    // 获取所有的集群
    router.GET("/clusters", handle.Clusters)
    // 获取所有的接口
    router.GET("/apis", handle.Apis)
    // 插件列表
    router.GET("/plugins", handle.Plugins)
    // 增加集群
    router.POST("/cluster", handle.AddCluster)
    // 增加后端服务
    router.POST("/backend", handle.AddBackend)
    // 增加路由规则
    router.POST("/api", handle.AddApi)
    // 更新路由规则
    router.POST("/api/update", handle.UpdateApi)
    go router.Run(":8081")
    engine.Run(":8080")
}

