package gateway

import (
    "sync"
    "net/http"
    "log"
)

type (
    Engine struct {
        pool       sync.Pool
        clientPool sync.Pool

        routeTable *RouteTable
        clusters   *ClusterGroup

        plugins    HandlesChain
    }
    HandlesChain []Plugin
)

func New() *Engine {
    engine := &Engine{
        routeTable:         NewRouteTable(),
        clusters:            &ClusterGroup{},
    }
    plugins := make(HandlesChain, 0)
    engine.plugins = plugins
    engine.pool.New = func() interface{} {
        return engine.allocateContext()
    }
    engine.clientPool.New = func() interface{} {
        return &http.Client{}
    }
    return engine
}

func (engine *Engine) allocateContext() *Context {
    return &Context{engine: engine}
}

// 插件

func (engine *Engine)RegisterPlugin(plugin Plugin) error {
    index := engine.plugins.indexOf(plugin.Name())
    if index != -1 {
        return PluginAlreadyExist
    }
    engine.plugins = append(engine.plugins, plugin)
    return nil
}

func (engine *Engine)Plugin(pluginName string) (bool, Plugin) {
    for i, l := 0, len(engine.plugins); i < l; i++ {
        if pluginName == engine.plugins[i].Name() {
            return true, engine.plugins[i]
        }
    }
    return false, nil
}

func (handlesChain HandlesChain) indexOf(pluginName string) (index int) {
    index = -1
    for i, l := 0, len(handlesChain); i < l; i++ {
        if pluginName == handlesChain[i].Name() {
            return i
        }
    }
    return
}



// 插件列表
func (engine *Engine) Plugins() []PluginInfo {
    plugins := make([]PluginInfo, 0)
    for i, l := 0, len(engine.plugins); i < l; i++ {
        plugins = append(plugins, PluginInfo{
            engine.plugins[i].Name(),
            engine.plugins[i].Version(),
        })
    }
    return plugins
}


// Client Get http client .
func (engine *Engine) Client() *http.Client {
    return engine.clientPool.Get().(*http.Client)
}

// Release .
func (engine *Engine) Release(client *http.Client) {
    client.Timeout = 10
    engine.clientPool.Put(client)
}

// add cluster

func (engine *Engine)AddCluster(cluster *Cluster) error {
    return engine.clusters.Add(cluster)
}

// remove cluster
func (engine *Engine)RemoveCluster(clusterName string) error {
    return engine.clusters.Remove(clusterName)
}

// update cluster
func (engine *Engine)Update(cluster *Cluster) {
    engine.clusters.Update(cluster)
}

func (engine *Engine)Cluster(clusterName string) (has bool, cluster *Cluster) {
    return engine.clusters.Get(clusterName)
}

func (engine *Engine) Clusters() []*Cluster {
    return engine.clusters.Clusters()
}

func (engine *Engine)combinePlugins(routeInfo RouteInfo) RouteInfo {
    // 处理前置
    for i, l := 0, len(routeInfo.BeginHandles); i < l; i ++ {
        if has, plugin := engine.Plugin(routeInfo.BeginHandles[i]); has {
            routeInfo.handles = append(routeInfo.handles, plugin)
        }
    }
    _, plugin := engine.Plugin("proxy")
    // 主调度
    routeInfo.handles = append(routeInfo.handles, plugin)
    // 处理后置
    for i, l := 0, len(routeInfo.AfterHandles); i < l; i ++ {
        if has, plugin := engine.Plugin(routeInfo.AfterHandles[i]); has {
            routeInfo.handles = append(routeInfo.handles, plugin)
        }
    }
    return routeInfo
}

// add route
func (engine *Engine)Route(routeInfo RouteInfo) error {
    routeInfo = engine.combinePlugins(routeInfo)
    return engine.routeTable.Add(routeInfo)
}

func (engine *Engine)UnRoute(method, url string) {
    engine.routeTable.Remove(method, url)
}

func (engine *Engine)UpdateRoute(routeInfo RouteInfo) {
    routeInfo = engine.combinePlugins(routeInfo)
    engine.routeTable.Update(routeInfo)
}

func (engine *Engine)Routes() []RouteInfo {
    return engine.routeTable.tables
}

func (engine *Engine) Run(addr string) (err error) {
    defer func() {
        log.Printf("[Gateway]%s", err.Error())
    }()
    err = http.ListenAndServe(addr, engine)
    return
}

func (engine *Engine) RunTLS(addr string, certFile string, keyFile string) (err error) {
    defer func() {
        log.Printf("[Gateway]%s", err.Error())
    }()

    err = http.ListenAndServeTLS(addr, certFile, keyFile, engine)
    return
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    c := engine.pool.Get().(*Context)
    c.writermem.reset(w)
    c.Request = req
    c.reset()
    c.responses = make([]combineResponse, 0)
    engine.handleHTTPRequest(c)
    engine.pool.Put(c)
}

func (engine *Engine) handleHTTPRequest(context *Context) {
    httpMethod := context.Request.Method
    path := context.Request.URL.Path
    has, routeInfo := engine.routeTable.Get(httpMethod, path)
    if has {
        context.routeInfo = routeInfo
        context.handlers = routeInfo.handles
        context.Next()
        return
    }
    context.JSON(APINotFound)
}

type combineResponse struct {
    Attr     string
    Error    error
    Response *http.Response
}
