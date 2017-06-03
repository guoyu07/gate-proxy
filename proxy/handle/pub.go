package handle

import (
	"goodsogood/gateway/proxy/global"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", nil)
}

// Plugins .
func Plugins(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": global.Store.Proxy().Plugins(),
	})
}

func Status(ctx *gin.Context) {

}
