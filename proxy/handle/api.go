package handle

import (
	"encoding/json"
	"fmt"
	"goodsogood/gateway"
	"goodsogood/gateway/proxy/global"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/buntdb"
)

func Apis(ctx *gin.Context) {
	apis := make([]gateway.RouteInfo, 0)
	global.Store.DB().View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("api", func(key, value string) bool {
			var routeInfo gateway.RouteInfo
			json.Unmarshal([]byte(value), &routeInfo)
			apis = append(apis, routeInfo)
			return true
		})
		return err
	})
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": apis,
	})
}

// AddAPI .
func AddAPI(ctx *gin.Context) {
	var form gateway.RouteInfo
	err := ctx.BindJSON(&form)
	if err != nil {
		ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
		return
	}
	if err := global.Store.Proxy().Route(form); err != nil {
		ctx.JSON(http.StatusOK, err)
		return
	}
	if len(form.NodeGroup) > 5 {
		ctx.JSON(http.StatusOK, gateway.ToManyNodes)
		return
	}
	global.Store.DB().Update(func(tx *buntdb.Tx) error {
		apiByte, _ := json.Marshal(form)
		key := fmt.Sprintf("api:%s:%s", form.Method, form.URL)
		_, _, err := tx.Set(key, string(apiByte), nil)
		return err
	})
	ctx.JSON(http.StatusOK, gateway.SUCCESS)
}

type UpdateApiForm struct {
	Method string            `json:"method"`
	URL    string            `json:"url"`
	Info   gateway.RouteInfo `json:"info"`
}

// UpdateAPI .
func UpdateAPI(ctx *gin.Context) {
	var form UpdateApiForm
	err := ctx.BindJSON(&form)
	if err != nil {
		ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
		return
	}
	if len(form.Info.NodeGroup) > 5 {
		ctx.JSON(http.StatusOK, gateway.ToManyNodes)
		return
	}
	global.Store.Proxy().UpdateRoute(form.Method, form.URL, form.Info)
	global.Store.DB().Update(func(tx *buntdb.Tx) error {
		apiByte, _ := json.Marshal(form.Info)
		key := fmt.Sprintf("api:%s:%s", form.Method, form.URL)
		_, _, err := tx.Set(key, string(apiByte), nil)
		return err
	})
	ctx.JSON(http.StatusOK, gateway.SUCCESS)
}

type DeleteApiForm struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

func DeleteAPI(ctx *gin.Context) {
	var form DeleteApiForm
	err := ctx.BindJSON(&form)
	if err != nil {
		ctx.JSON(http.StatusOK, gateway.ParamParseFailed)
		return
	}
	if form.Method == "" || form.URL == "" {
		ctx.JSON(http.StatusOK, gateway.URLNotValid)
		return
	}
	global.Store.Proxy().UnRoute(form.Method, form.URL)
	global.Store.DB().Update(func(tx *buntdb.Tx) error {
		key := fmt.Sprintf("api:%s:%s", form.Method, form.URL)
		_, err := tx.Delete(key)
		return err
	})
	ctx.JSON(http.StatusOK, gateway.SUCCESS)
}
