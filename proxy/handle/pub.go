package handle

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "goodsogood/gateway/proxy/global"
)

func Plugins(ctx *gin.Context) {
    ctx.JSON(http.StatusOK, global.Store.Proxy().Plugins())
}
