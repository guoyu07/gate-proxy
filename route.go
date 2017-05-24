package gateway

import (
    "sync"
)

type (
    RouteTable struct {
        m      *sync.RWMutex
        tables []RouteInfo
        table  []*RouteGroup
    }
    RouteInfo struct {
        Name         string `json:"name"`
        Method       string `json:"method"`
        URL          string `json:"url"`
        Domain       string `json:"domain"`
        // 路由前操作
        BeginHandles []string `json:"beginHandles"`
        // 路由后操作
        AfterHandles []string `json:"afterHandles"`
        NodeGroup    []Node `json:"nodeGroup"`

        handles      HandlesChain
    }
    RouteGroup struct {
        Method string
        mtx    *sync.RWMutex
        routes []RouteInfo
    }
)

var methods = []string{"GET", "POST", "DELETE", "PUT"}

func NewRouteTable() *RouteTable {
    routeTable := &RouteTable{
        m: &sync.RWMutex{},
        tables: make([]RouteInfo, 0),
    }
    num := len(methods)
    routeTable.table = make([]*RouteGroup, num)
    for i := 0; i < num; i++ {
        routeTable.table[i] = &RouteGroup{
            Method:methods[i],
            mtx: &sync.RWMutex{},
            routes: make([]RouteInfo, 0),
        }
    }
    return routeTable
}
// 获取规则
func (table *RouteTable)Get(method, url string) (has bool, routeInfo RouteInfo) {
    for i, l := 0, len(methods); i < l; i++ {
        if table.table[i].Method == method {
            return table.table[i].Get(url)
        }
    }
    return false, routeInfo
}

// 增加路由
func (table *RouteTable)Add(routeInfo RouteInfo) (err error) {
    for i, l := 0, len(methods); i < l; i ++ {
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

// 删除路由
func (table *RouteTable)Remove(method, url string) (err error) {
    for i, l := 0, len(methods); i < l; i ++ {
        if table.table[i].Method == method {
            if index := table.table[i].indexOf(url); index != -1 {
                table.table[i].mtx.Lock()
                table.table[i].routes = append(table.table[i].routes[:index], table.table[i].routes[index + 1:]...)
                table.table[i].mtx.Unlock()
                return nil
            }
            return APINotFound
        }
    }
    return UnknowableMethod
}

// 更新路由
func (table *RouteTable)Update(routeInfo RouteInfo) (err error) {
    for i, l := 0, len(methods); i < l; i ++ {
        if table.table[i].Method == routeInfo.Method {
            if index := table.table[i].indexOf(routeInfo.URL); index != -1 {
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

func (r *RouteGroup)Get(url string) (has bool, routeInfo RouteInfo) {
    index := r.indexOf(url)
    if index != -1 {
        return true, r.routes[index]
    }
    return false, routeInfo
}

func (r *RouteGroup)indexOf(url string) (index int) {
    index = -1
    r.mtx.RLock()
    defer r.mtx.RUnlock()
    for i, l := 0, len(r.routes); i < l; i++ {
        if r.routes[i].URL == url {
            return i
        }
    }
    return index
}

// 路由是否存在
//func (table *RouteTable)IsExist(method, url string) (index int) {
//    index = -1
//    for i, l := 0, len(methods); i < l; i++ {
//        if table.table[i].Method == method {
//            table.table[i].mtx.RLock()
//            defer table.table[i].mtx.RUnlock()
//            for j, jl := 0, len(table.table[i].routes); j < jl; j++ {
//
//            }
//        }
//    }
//    nums := len(table.tables)
//    for i := 0; i < nums; i++ {
//        if table.tables[i].Method == method && table.tables[i].URL == url {
//            return i
//        }
//    }
//    return
//}


//// 获取规则
//func (table *RouteTable)Get(method, url string) (has bool, routeInfo RouteInfo) {
//    index := table.IsExist(method, url)
//    if index != -1 {
//        return true, table.tables[index]
//    }
//    return false, routeInfo
//}

//// 增加路由
//func (table *RouteTable)Add(routeInfo RouteInfo) (err error) {
//    if table.IsExist(routeInfo.Method, routeInfo.URL) != -1 {
//        return APIAlreadyExist
//    }
//    table.m.Lock()
//    defer table.m.Unlock()
//    table.tables = append(table.tables, routeInfo)
//    return nil
//}

//// 删除路由
//func (table *RouteTable)Remove(method, url string) (err error) {
//    if index := table.IsExist(method, url); index != -1 {
//        table.m.Lock()
//        defer table.m.Unlock()
//        table.tables = append(table.tables[:index], table.tables[index + 1:]...)
//        return nil
//    }
//    return APINotFound
//}

// 路由是否存在
//func (table *RouteTable)IsExist(method, url string) (index int) {
//    index = -1
//    table.m.RLock()
//    defer table.m.RUnlock()
//    nums := len(table.tables)
//    for i := 0; i < nums; i++ {
//        if table.tables[i].Method == method && table.tables[i].URL == url {
//            return i
//        }
//    }
//    return
//}

//// 更新路由
//func (table *RouteTable)Update(routeInfo RouteInfo) {
//    index := table.IsExist(routeInfo.Method, routeInfo.URL)
//    table.m.Lock()
//    defer table.m.Unlock()
//    if index == -1 {
//        table.tables = append(table.tables, routeInfo)
//        return
//    }
//    table.tables[index] = routeInfo
//    return
//}
