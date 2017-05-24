package gateway

import (
    "testing"
    "net/http"
    "sync"
    "fmt"
)

func TestRace(t *testing.T) {
    engine := New()
    engine.RegisterPlugin(Proxy{})
    engine.Route(RouteInfo{
        Name: "登录接口",
        Method:"GET",
        URL:"/login",
        Domain:"1",
        NodeGroup: []Node{
            Node{
                Attr: "info",
                Cluster:"UserBaseCluster",
                Rewrite: "/user/login",
            },
        },
    })
    runRequest(t, engine, "GET", "/login")
}

func BenchmarkEngineOneRouter(b *testing.B) {
    engine := New()
    engine.RegisterPlugin(Proxy{})
    engine.Route(RouteInfo{
        Name: "登录接口",
        Method:"GET",
        URL:"/login",
        Domain:"1",
        NodeGroup: []Node{
            Node{
                Attr: "info",
                Cluster:"UserBaseCluster",
                Rewrite: "/user/login",
            },
        },
    })
    runRequestBenchmark(b, engine, "GET", "/ping")
}

func BenchmarkEngineMultiRouter(b *testing.B) {
    engine := New()
    engine.RegisterPlugin(Proxy{})
    for i:=0; i<1000; i++ {
        engine.Route(RouteInfo{
            Name: "登录接口",
            Method:"GET",
            URL:fmt.Sprintf("/login%d", i),
            Domain:"1",
            NodeGroup: []Node{
                Node{
                    Attr: "info",
                    Cluster:"UserBaseCluster",
                    Rewrite: "/user/login",
                },
            },
        })
    }
    runRequestBenchmark(b, engine, "GET", "/login")
}

type Proxy struct {

}

func (p Proxy)Name()string {
    return "proxy"
}

func (p Proxy)Version()string {
    return "0.1"
}

func (p Proxy)Handle(ctx *Context) {
    ctx.Writer.Header().Set("Server", "Gate 0.1")
    nodes := len(ctx.RouteInfo().NodeGroup)
    switch nodes {
    case 0:
        ctx.JSON(BackendServiceError)
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
    ctx.Render()
}

type mockWriter struct {
    headers http.Header
}

func newMockWriter() *mockWriter {
    return &mockWriter{
        http.Header{},
    }
}

func (m *mockWriter) Header() (h http.Header) {
    return m.headers
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
    return len(p), nil
}

func (m *mockWriter) WriteString(s string) (n int, err error) {
    return len(s), nil
}

func (m *mockWriter) WriteHeader(int) {}

func runRequestBenchmark(B *testing.B, r *Engine, method, path string) {
    // create fake request
    req, err := http.NewRequest(method, path, nil)
    if err != nil {
        panic(err)
    }
    w := newMockWriter()
    B.ReportAllocs()
    B.ResetTimer()
    for i := 0; i < B.N; i++ {
        r.ServeHTTP(w, req)
    }
}

func runRequest(t *testing.T, r *Engine, method, path string) {
    for i := 0; i < 2; i++ {
        req, err := http.NewRequest(method, path, nil)
        if err != nil {
            panic(err)
        }
        w := newMockWriter()
        go r.ServeHTTP(w, req)
    }
}