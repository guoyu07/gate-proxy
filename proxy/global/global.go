package global

import (
    "github.com/tidwall/buntdb"
    "goodsogood/gateway"
    "goodsogood/gateway/proxy/types"
    "encoding/json"
)

const (
    CLUSTER_INDEX_KEY = "cluster"
    BACKEND_INDEX_KEY = "backend"
    API_INDEX_KEY = "api"
)

type GlobalStore struct {
    db    *buntdb.DB
    proxy *gateway.Engine
}

var Store  *GlobalStore

func init() {
    Store = &GlobalStore{}
}

func (s *GlobalStore)SetDB(db *buntdb.DB) {
    s.db = db
    db.CreateIndex(CLUSTER_INDEX_KEY, "cluster:*", buntdb.IndexString)
    db.CreateIndex(BACKEND_INDEX_KEY, "backend:*", buntdb.IndexString)
    db.CreateIndex(API_INDEX_KEY, "api:*", buntdb.IndexString)
}

func (s *GlobalStore)CloseDB() error {
    return s.db.Close()
}

func (s *GlobalStore)SetProxy(proxy *gateway.Engine) {
    s.proxy = proxy
}

func (s *GlobalStore)DB() *buntdb.DB {
    return s.db
}

func (s *GlobalStore)Proxy() *gateway.Engine {
    return s.proxy
}

func (s *GlobalStore)LoadCache() {
    s.db.View(func(tx *buntdb.Tx) error {
        err := tx.Ascend("cluster", func(key, value string) bool {
            s.proxy.AddCluster(&gateway.Cluster{Name:value})
            return true
        })
        err = tx.Ascend("backend", func(key, value string) bool {
            var backendInfo types.BackendInfo
            json.Unmarshal([]byte(value), &backendInfo)
            _, cluster := s.proxy.Cluster(backendInfo.ClusterName)
            cluster.Add(&gateway.Backend{
                Addr  :backendInfo.Addr,
                Schema :backendInfo.Schema,
                HeartPath        :backendInfo.HeartPath,
                HeartResponseBody:backendInfo.HeartResponseBody,
                HeartDuration :backendInfo.HeartDuration,
                Timeout          :backendInfo.Timeout,
                MaxQPS           :backendInfo.MaxQPS,
            })
            return true
        })
        err = tx.Ascend("api", func(key, value string) bool {
            var routeInfo gateway.RouteInfo
            json.Unmarshal([]byte(value), &routeInfo)
            s.proxy.Route(routeInfo)
            return true
        })
        return err
    })
}