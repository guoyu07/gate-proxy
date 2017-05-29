package gateway

import (
    "sync"
    "net/http"
)

type Proxy struct {
}

func (p Proxy) Name() string {
    return "proxy"
}

func (p Proxy) Version() string {
    return "0.1"
}

func (p Proxy) Handle(ctx *Context) {
    ctx.Writer.Header().Set("Server", "Gate 0.1")
    nodes := len(ctx.RouteInfo().NodeGroup)
    switch nodes {
    case 0:
        ctx.Render(http.StatusOK, BackendServiceError)
        return
    case 1:
        // 执行单个节点
        ctx.RouteInfo().NodeGroup[0].Do(ctx, nil)
    default:
        // 合并执行多个节点
        wg := &sync.WaitGroup{}
        wg.Add(nodes)
        for i := 0; i < nodes; i++ {
            go ctx.RouteInfo().NodeGroup[i].Do(ctx, wg)
        }
        wg.Wait()
    }
    ctx.Render(http.StatusOK, nil)
}
