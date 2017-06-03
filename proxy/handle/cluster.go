package handle

import (
	"encoding/json"
	"fmt"
	"goodsogood/gateway"
	"goodsogood/gateway/proxy/global"
	"goodsogood/gateway/proxy/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/buntdb"
)

// Clusters .
func Clusters(ctx *gin.Context) {
	var clusters = make([]gateway.H, 0)
	global.Store.DB().View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("cluster", func(key, value string) bool {
			info := gateway.H{}
			clusterInfo := gateway.Cluster{}
			if err := json.Unmarshal([]byte(value), &clusterInfo); err == nil {
				info["clusterName"] = clusterInfo.Name
				info["description"] = clusterInfo.Description
			} else {
				info["clusterName"] = value
				info["description"] = ""
			}
			clusterName, _ := info["clusterName"]
			has, cluster := global.Store.Proxy().Cluster(clusterName.(string))
			if has {
				info["exist"] = true
				info["backendNum"] = cluster.Backends().Len()
			}
			clusters = append(clusters, info)
			return true
		})
		return err
	})
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": clusters,
	})
}

// AddCluster .
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
	cluster := &gateway.Cluster{Name: form.Name, Description: form.Description}
	if err := global.Store.Proxy().AddCluster(cluster); err != nil {
		ctx.JSON(http.StatusOK, err)
		return
	}
	global.Store.DB().Update(func(tx *buntdb.Tx) error {
		clusterByte, _ := json.Marshal(cluster)
		key := fmt.Sprintf("cluster:%s", form.Name)
		_, _, err := tx.Set(key, string(clusterByte), nil)
		return err
	})
	ctx.JSON(http.StatusOK, gateway.SUCCESS)
}

// UpdateCluster .
func UpdateCluster(ctx *gin.Context) {
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
	cluster := &gateway.Cluster{Name: form.Name, Description: form.Description}
	global.Store.Proxy().Update(cluster)
	global.Store.DB().Update(func(tx *buntdb.Tx) error {
		clusterByte, _ := json.Marshal(cluster)
		key := fmt.Sprintf("cluster:%s", form.Name)
		_, _, err := tx.Set(key, string(clusterByte), nil)
		return err
	})
	ctx.JSON(http.StatusOK, gateway.SUCCESS)
}

// DelCluster .
func DelCluster(ctx *gin.Context) {
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
	has, cluster := global.Store.Proxy().Cluster(form.Name)
	if !has {
		ctx.JSON(http.StatusOK, gateway.ClusterNotFound)
		return
	}
	if cluster.Backends().Len() > 0 {
		ctx.JSON(http.StatusOK, gateway.BackendsNumNotZero)
		return
	}
	err = global.Store.Proxy().RemoveCluster(form.Name)
	if err != nil {
		ctx.JSON(http.StatusOK, err)
		return
	}
	global.Store.DB().Update(func(tx *buntdb.Tx) error {
		key := fmt.Sprintf("cluster:%s", form.Name)
		_, err := tx.Delete(key)
		return err
	})
	ctx.JSON(http.StatusOK, gateway.SUCCESS)
}

// Backends .
func Backends(ctx *gin.Context) {
	clusterName := ctx.Param("clusterName")
	if clusterName == "" {
		ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
		return
	}
	has, cluster := global.Store.Proxy().Cluster(clusterName)
	if !has {
		ctx.JSON(http.StatusOK, gateway.ClusterNotFound)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": cluster.Backends(),
	})
}

// AddBackend . 增加后端服务
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
	if !form.HeartDisabled && len(form.HeartPath) < 1 {
		ctx.JSON(http.StatusOK, gateway.HeartPathNotEmpty)
		return
	}
	// maxQPS
	if form.MaxQPS < 1 {
		ctx.JSON(http.StatusOK, gateway.MaxQPSNotZero)
		return
	}
	backend := &gateway.Backend{
		Addr:              form.Addr,
		Schema:            form.Schema,
		HeartPath:         form.HeartPath,
		HeartDisabled:     form.HeartDisabled,
		HeartResponseBody: form.HeartResponseBody,
		HeartDuration:     form.HeartDuration,
		Timeout:           form.Timeout,
		MaxQPS:            form.MaxQPS,
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

// UpdateBackendForm .
type UpdateBackendForm struct {
	Addr    string            `json:"addr"`
	Backend types.BackendInfo `json:"backendInfo"`
}

// UpdateBackend . 更新后端服务
func UpdateBackend(ctx *gin.Context) {
	var form UpdateBackendForm
	err := ctx.BindJSON(&form)
	if err != nil {
		ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
		return
	}
	has, cluster := global.Store.Proxy().Cluster(form.Backend.ClusterName)
	if !has {
		ctx.JSON(http.StatusOK, gateway.ClusterNotFound)
		return
	}
	// schema
	if form.Backend.Schema != "http" && form.Backend.Schema != "https" {
		ctx.JSON(http.StatusOK, gateway.SchemaUnknowable)
		return
	}
	// addr
	if len(form.Addr) < 1 {
		ctx.JSON(http.StatusOK, gateway.AddrUnknowable)
		return
	}
	// heartPath
	if !form.Backend.HeartDisabled && len(form.Backend.HeartPath) < 1 {
		ctx.JSON(http.StatusOK, gateway.HeartPathNotEmpty)
		return
	}
	// maxQPS
	if form.Backend.MaxQPS < 1 {
		ctx.JSON(http.StatusOK, gateway.MaxQPSNotZero)
		return
	}
	backend := &gateway.Backend{
		Addr:              form.Backend.Addr,
		Schema:            form.Backend.Schema,
		HeartPath:         form.Backend.HeartPath,
		HeartResponseBody: form.Backend.HeartResponseBody,
		HeartDuration:     form.Backend.HeartDuration,
		Timeout:           form.Backend.Timeout,
		MaxQPS:            form.Backend.MaxQPS,
	}
	var key string
	if backend.Addr != form.Addr {
		// 变换了addr
		cluster.Remove(form.Addr)
		if err := cluster.Add(backend); err != nil {
			ctx.JSON(http.StatusOK, err)
			return
		}
		key = fmt.Sprintf("backend:%s", form.Addr)
	} else {
		cluster.Update(backend)
		key = fmt.Sprintf("backend:%s", backend.Addr)
	}
	global.Store.DB().Update(func(tx *buntdb.Tx) error {
		backendByte, _ := json.Marshal(backend)
		_, _, err := tx.Set(key, string(backendByte), nil)
		return err
	})
	ctx.JSON(http.StatusOK, gateway.SUCCESS)
}

// DelBackendForm .
type DelBackendForm struct {
	ClusterName string `json:"clusterName"`
	Addr        string `json:"addr"`
}

// DelBackend . 删除后端服务
func DelBackend(ctx *gin.Context) {
	var form DelBackendForm
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
	err = cluster.Remove(form.Addr)
	if err != nil {
		ctx.JSON(http.StatusOK, err)
		return
	}
	global.Store.DB().Update(func(tx *buntdb.Tx) error {
		key := fmt.Sprintf("backend:%s", form.Addr)
		_, err := tx.Delete(key)
		return err
	})
	ctx.JSON(http.StatusOK, gateway.SUCCESS)
}
