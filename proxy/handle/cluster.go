package handle

import (
    "github.com/gin-gonic/gin"
    "goodsogood/gateway"
    "github.com/tidwall/buntdb"
    "fmt"
    "goodsogood/gateway/proxy/types"
    "net/http"
    "goodsogood/gateway/proxy/global"
    "encoding/json"
)

func Clusters(ctx *gin.Context) {
    var clusters = make([]gateway.Cluster, 0)
    global.Store.DB().View(func(tx *buntdb.Tx) error {
        err := tx.Ascend("cluster", func(key, value string) bool {
            clusters = append(clusters, gateway.Cluster{Name:value})
            return true
        })
        return err
    })
    ctx.JSON(http.StatusOK, clusters)
}

func AddCluster(ctx *gin.Context) {
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
    if err := global.Store.Proxy().AddCluster(cluster); err != nil {
        ctx.JSON(http.StatusOK, err)
        return
    }
    global.Store.DB().Update(func(tx *buntdb.Tx) error {
        key := fmt.Sprintf("cluster:%s", form.Name)
        _, _, err := tx.Set(key, form.Name, nil)
        return err
    })
    ctx.JSON(http.StatusOK, gateway.SUCCESS)
}

func AddBackend(ctx *gin.Context) {
    var form types.BackendInfo
    err := ctx.BindJSON(&form)
    if err != nil {
        ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
        return
    }
    has, cluster := global.Store.Proxy().Cluster(form.ClusterName)
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
    global.Store.DB().Update(func(tx *buntdb.Tx) error {
        backendByte, _ := json.Marshal(form)
        key := fmt.Sprintf("backend:%s", form.Addr)
        _, _, err := tx.Set(key, string(backendByte), nil)
        return err
    })
    ctx.JSON(http.StatusOK, gateway.SUCCESS)
}