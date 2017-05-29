package handle

import (
	"goodsogood/gateway/proxy/global"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Plugins(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, global.Store.Proxy().Plugins())
}

func Status(ctx *gin.Context) {

}
