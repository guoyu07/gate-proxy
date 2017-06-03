package gateway

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type (
	// Engine .
	Engine struct {
		pool       sync.Pool
		clientPool sync.Pool

		routeTable *RouteTable
		clusters   *ClusterGroup

		plugins HandlesChain
	}
	// HandlesChain .
	HandlesChain []Plugin
	//
)

func New() *Engine {
	engine := &Engine{
		routeTable: NewRouteTable(),
		clusters:   &ClusterGroup{},
		plugins:    make(HandlesChain, 0),
	}
	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}
	engine.clientPool.New = func() interface{} {
		return &http.Client{}
	}
	engine.RegisterPlugin(Proxy{})
	engine.RegisterPlugin(NewRecovery())
	return engine
}

func (engine *Engine) allocateContext() *Context {
	return &Context{engine: engine}
}

// RegisterPlugin . 注册插件
func (engine *Engine) RegisterPlugin(plugin Plugin) error {
	index := engine.plugins.indexOf(plugin.Name())
	if index != -1 {
		return PluginAlreadyExist
	}
	engine.plugins = append(engine.plugins, plugin)
	return nil
}

// Plugin . 获取插件
func (engine *Engine) Plugin(pluginName string) (bool, Plugin) {
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

// Plugins . 插件列表
func (engine *Engine) Plugins() []PluginInfo {
	plugins := make([]PluginInfo, 0)
	for i, l := 0, len(engine.plugins); i < l; i++ {
		plugins = append(plugins, PluginInfo{
			engine.plugins[i].Name(),
			engine.plugins[i].Private(),
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

// AddCluster .
func (engine *Engine) AddCluster(cluster *Cluster) error {
	return engine.clusters.Add(cluster)
}

// RemoveCluster .
func (engine *Engine) RemoveCluster(clusterName string) error {
	return engine.clusters.Remove(clusterName)
}

// Update .
func (engine *Engine) Update(cluster *Cluster) {
	engine.clusters.Update(cluster)
}

// Cluster .
func (engine *Engine) Cluster(clusterName string) (has bool, cluster *Cluster) {
	return engine.clusters.Get(clusterName)
}

// Clusters .
func (engine *Engine) Clusters() []*Cluster {
	return engine.clusters.Clusters()
}

func (engine *Engine) combinePlugins(routeInfo RouteInfo) RouteInfo {
	_, recovery := engine.Plugin("recovery")
	routeInfo.handles = append(routeInfo.handles, recovery)
	// 处理插件
	for i, l := 0, len(routeInfo.Handlers); i < l; i++ {
		if has, plugin := engine.Plugin(routeInfo.Handlers[i]); has {
			routeInfo.handles = append(routeInfo.handles, plugin)
		}
	}
	_, plugin := engine.Plugin("proxy")
	// 主调度
	routeInfo.handles = append(routeInfo.handles, plugin)
	return routeInfo
}

// Route .
func (engine *Engine) Route(routeInfo RouteInfo) error {
	routeInfo = engine.combinePlugins(routeInfo)
	return engine.routeTable.Add(routeInfo)
}

// UnRoute .
func (engine *Engine) UnRoute(method, url string) {
	engine.routeTable.Remove(method, url)
}

// UpdateRoute .
func (engine *Engine) UpdateRoute(method, url string, routeInfo RouteInfo) {
	routeInfo = engine.combinePlugins(routeInfo)
	engine.routeTable.Update(method, url, routeInfo)
}

// Routes .
func (engine *Engine) Routes() []RouteInfo {
	return engine.routeTable.tables
}

// Run .
func (engine *Engine) Run(addr string) (err error) {
	defer func() {
		log.Printf("[Gateway]%s", err.Error())
	}()
	fmt.Println("Gateway Listening and serving HTTP on ", addr)
	err = http.ListenAndServe(addr, engine)
	return
}

// RunTLS .
func (engine *Engine) RunTLS(addr string, certFile string, keyFile string) (err error) {
	defer func() {
		log.Printf("[Gateway]%s", err.Error())
	}()
	fmt.Println("Gateway Listening and serving HTTPS on ", addr)
	err = http.ListenAndServeTLS(addr, certFile, keyFile, engine)
	return
}

// SeverHTTP .
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
	// parse request
	has, routeInfo := engine.routeTable.Get(httpMethod, path)
	if has {
		context.routeInfo = routeInfo
		context.handlers = routeInfo.handles
		context.Next()
		return
	}
	context.Render(http.StatusOK, APINotFound)
}

type combineResponse struct {
	Attr     string
	Error    error
	Response []byte
}
