package gateway

import (
	"regexp"
	"sync"
)

type (
	RouteTable struct {
		m      *sync.RWMutex
		tables []RouteInfo
		table  []*RouteGroup
	}
	RouteInfo struct {
		Name   string `json:"name"`
		Method string `json:"method"`
		URL    string `json:"url"`
		Domain string `json:"domain"`
		// 路由前操作
		Handlers  []string `json:"handlers"`
		NodeGroup []Node   `json:"nodeGroup"`

		handles HandlesChain
	}
	RouteGroup struct {
		Method string
		mtx    *sync.RWMutex
		routes []RouteInfo
	}
)

var methods = []string{"GET", "POST", "DELETE", "PUT", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE"}

// NewRouteTable . 路由表
func NewRouteTable() *RouteTable {
	routeTable := &RouteTable{
		m:      &sync.RWMutex{},
		tables: make([]RouteInfo, 0),
	}
	num := len(methods)
	routeTable.table = make([]*RouteGroup, num)
	for i := 0; i < num; i++ {
		routeTable.table[i] = &RouteGroup{
			Method: methods[i],
			mtx:    &sync.RWMutex{},
			routes: make([]RouteInfo, 0),
		}
	}
	return routeTable
}

// Get . 获取规则
func (table *RouteTable) Get(method, url string) (has bool, routeInfo RouteInfo) {
	for i, l := 0, len(methods); i < l; i++ {
		if table.table[i].Method == method {
			return table.table[i].Get(url)
		}
	}
	return false, routeInfo
}

// Add . 增加路由
func (table *RouteTable) Add(routeInfo RouteInfo) (err error) {
	routeInfo = routeInfo.initRegexp()
	for i, l := 0, len(methods); i < l; i++ {
		if table.table[i].Method == routeInfo.Method {
			if table.table[i].indexOf(routeInfo.URL) != -1 {
				return APIAlreadyExist
			}
			table.table[i].mtx.Lock()
			table.table[i].routes = append(table.table[i].routes, routeInfo)
			table.table[i].mtx.Unlock()
			return nil
		}
	}
	return UnknowableMethod
}

// Remove . 删除路由
func (table *RouteTable) Remove(method, url string) (err error) {
	for i, l := 0, len(methods); i < l; i++ {
		if table.table[i].Method == method {
			if index := table.table[i].indexOf(url); index != -1 {
				table.table[i].mtx.Lock()
				table.table[i].routes = append(table.table[i].routes[:index], table.table[i].routes[index+1:]...)
				table.table[i].mtx.Unlock()
				return nil
			}
			return APINotFound
		}
	}
	return UnknowableMethod
}

// Update . 更新路由
func (table *RouteTable) Update(method, url string, routeInfo RouteInfo) (err error) {
	routeInfo = routeInfo.initRegexp()
	for i, l := 0, len(methods); i < l; i++ {
		if table.table[i].Method == method {
			if index := table.table[i].indexOf(url); index != -1 {
				table.table[i].mtx.Lock()
				table.table[i].routes[index] = routeInfo
				table.table[i].mtx.Unlock()
				return nil
			}
			table.table[i].mtx.Lock()
			table.table[i].routes = append(table.table[i].routes, routeInfo)
			table.table[i].mtx.Unlock()
			return nil
		}
	}
	return UnknowableMethod
}

// Get .
func (r *RouteGroup) Get(url string) (has bool, routeInfo RouteInfo) {
	index := r.indexOf(url)
	if index != -1 {
		return true, r.routes[index]
	}
	return false, routeInfo
}

func (routeInfo RouteInfo) initRegexp() RouteInfo {
	for i, l := 0, len(routeInfo.NodeGroup); i < l; i++ {
		for j, k := 0, len(routeInfo.NodeGroup[i].ParamGroup); j < k; j++ {
			if routeInfo.NodeGroup[i].ParamGroup[j].Validation != "" {
				routeInfo.NodeGroup[i].ParamGroup[j].rule = regexp.MustCompile(routeInfo.NodeGroup[i].ParamGroup[j].Validation)
			}
		}
	}
	return routeInfo
}

func (r *RouteGroup) indexOf(url string) (index int) {
	index = -1
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	l := len(r.routes)
	for i := 0; i < l; i++ {
		if r.routes[i].URL == url {
			return i
		}
	}
	return index
}
