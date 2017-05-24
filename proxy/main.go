package main

import (
    "goodsogood/gateway"
    "goodsogood/gateway/plugin/auth"
    "goodsogood/gateway/plugin/snapshot"
    "goodsogood/gateway/plugin/proxy"
    "github.com/gin-gonic/gin"
    "net/http"
    "goodsogood/gateway/proxy/types"
    "github.com/tidwall/buntdb"
    "log"
    "encoding/json"
    "fmt"
)

var (
    // 初始化
    db *buntdb.DB
    err error
    engine = gateway.New()
)

func checkPanic(err error) {
    if err != nil {
        panic(err)
    }
}

func main() {
    db, err = buntdb.Open("gateway.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    db.CreateIndex("cluster", "cluster:*", buntdb.IndexString)
    db.CreateIndex("backend", "backend:*", buntdb.IndexString)
    db.CreateIndex("api", "api:*", buntdb.IndexString)
    // 注册插件
    engine.RegisterPlugin(proxy.Proxy{})
    engine.RegisterPlugin(auth.Auth{})
    engine.RegisterPlugin(snapshot.Snapshot{})
    loadFromDB()
    router := gin.New()
    // 获取所有的集群
    router.GET("/clusters", func(ctx *gin.Context) {
        ctx.JSON(http.StatusOK, engine.Clusters())
    })
    // 获取所有的接口
    router.GET("/apis", func(ctx *gin.Context) {
        ctx.JSON(http.StatusOK, engine.Routes())
    })
    // 插件列表
    router.GET("/plugins", func(ctx *gin.Context) {
        ctx.JSON(http.StatusOK, engine.Plugins())
    })
    // 增加集群
    router.POST("/cluster", func(ctx *gin.Context) {
        var form types.ClusterInfo
        err := ctx.BindJSON(&form)
        if err != nil {
            ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
            return
        }
        if len(form.Name) < 1 {
            ctx.JSON(http.StatusOK, gateway.ClusterNameEmpty)
            return
        }
        cluster := &gateway.Cluster{Name:form.Name}
        if err := engine.AddCluster(cluster); err != nil {
            ctx.JSON(http.StatusOK, err)
            return
        }
        db.Update(func(tx *buntdb.Tx) error {
            key := fmt.Sprintf("cluster:%s", form.Name)
            _, _, err := tx.Set(key, form.Name, nil)
            return err
        })
        ctx.JSON(http.StatusOK, gateway.SUCCESS)
    })
    // 增加后端服务
    router.POST("/backend", func(ctx *gin.Context) {
        var form types.BackendInfo
        err := ctx.BindJSON(&form)
        if err != nil {
            ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
            return
        }
        has, cluster := engine.Cluster(form.ClusterName)
        if !has {
            ctx.JSON(http.StatusOK, gateway.ClusterNotFound)
            return
        }
        // schema
        if form.Schema != "http" && form.Schema != "https" {
            ctx.JSON(http.StatusOK, gateway.SchemaUnknowable)
            return
        }
        // addr
        if len(form.Addr) < 1 {
            ctx.JSON(http.StatusOK, gateway.AddrUnknowable)
            return
        }
        // heartPath
        if len(form.HeartPath) < 1 {
            ctx.JSON(http.StatusOK, gateway.HeartPathNotEmpty)
            return
        }
        // maxQPS
        if form.MaxQPS < 1 {
            ctx.JSON(http.StatusOK, gateway.MaxQPSNotZero)
            return
        }
        backend := &gateway.Backend{
            Addr  :form.Addr,
            Schema :form.Schema,
            HeartPath        :form.HeartPath,
            HeartResponseBody:form.HeartResponseBody,
            HeartDuration :form.HeartDuration,
            Timeout          :form.Timeout,
            MaxQPS           :form.MaxQPS,
        }
        if err := cluster.Add(backend); err != nil {
            ctx.JSON(http.StatusOK, err)
            return
        }
        db.Update(func(tx *buntdb.Tx) error {
            backendByte, _ := json.Marshal(form)
            key := fmt.Sprintf("backend:%s", form.Addr)
            _, _, err := tx.Set(key, string(backendByte), nil)
            return err
        })
        ctx.JSON(http.StatusOK, gateway.SUCCESS)
    })
    // 增加路由规则
    router.POST("/api", func(ctx *gin.Context) {
        var form gateway.RouteInfo
        err := ctx.BindJSON(&form)
        if err != nil {
            ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
            return
        }
        if err := engine.Route(form); err != nil {
            ctx.JSON(http.StatusOK, err)
            return
        }
        db.Update(func(tx *buntdb.Tx) error {
            apiByte, _ := json.Marshal(form)
            key := fmt.Sprintf("api:%s:%s", form.Method, form.URL)
            _, _, err := tx.Set(key, string(apiByte), nil)
            return err
        })
        ctx.JSON(http.StatusOK, gateway.SUCCESS)
    })
    // 更新路由规则
    router.POST("/api/update", func(ctx *gin.Context) {
        var form gateway.RouteInfo
        err := ctx.BindJSON(&form)
        if err != nil {
            ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
            return
        }
        engine.UpdateRoute(form)
        db.Update(func(tx *buntdb.Tx) error {
            apiByte, _ := json.Marshal(form)
            key := fmt.Sprintf("api:%s:%s", form.Method, form.URL)
            _, _, err := tx.Set(key, string(apiByte), nil)
            return err
        })
        ctx.JSON(http.StatusOK, gateway.SUCCESS)
    })
    go router.Run(":8081")
    engine.Run(":8080")
}

func loadFromDB() {
    err = db.View(func(tx *buntdb.Tx) error {
        err := tx.Ascend("cluster", func(key, value string) bool {
            engine.AddCluster(&gateway.Cluster{Name:value})
            return true
        })
        err = tx.Ascend("backend", func(key, value string) bool {
            var backendInfo types.BackendInfo
            json.Unmarshal([]byte(value), &backendInfo)
            _, cluster := engine.Cluster(backendInfo.ClusterName)
            cluster.Add(&gateway.Backend{
                Addr  :backendInfo.Addr,
                Schema :backendInfo.Schema,
                HeartPath        :backendInfo.HeartPath,
                HeartResponseBody:backendInfo.HeartResponseBody,
                HeartDuration :backendInfo.HeartDuration,
                Timeout          :backendInfo.Timeout,
                MaxQPS           :backendInfo.MaxQPS,
            })
            return true
        })
        err = tx.Ascend("api", func(key, value string) bool {
            var routeInfo gateway.RouteInfo
            json.Unmarshal([]byte(value), &routeInfo)
            engine.Route(routeInfo)
            return true
        })
        return err
    })
    if err != nil {
        log.Fatal("load db failed.")
    }
}
