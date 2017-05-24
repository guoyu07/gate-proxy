package gateway

import (
    "sync"
    "sort"
)

type (
    Cluster struct {
        // 集群名称
        Name     string `json:"name,omitempty"`
        // 后端服务
        backends BackendGroup `json:"backendGroup,omitempty"`

        rwMutex  sync.RWMutex
    }
    ClusterGroup struct {
        rwMutex  sync.RWMutex
        clusters []*Cluster
    }
)

var heartStop = make(chan string)

func (cluster *Cluster)Backends() BackendGroup {
    return cluster.backends
}

// 增加一个后端服务
func (cluster *Cluster) Add(backend *Backend) error {
    index := cluster.indexOf(backend.Addr)
    if index != - 1 {
        return BackendAlreadyExist
    }
    cluster.addBackend(backend)
    return nil
}

func (cluster *Cluster) addBackend(backend *Backend) {
    cluster.rwMutex.Lock()
    defer cluster.rwMutex.Unlock()
    backend.getDefaultSetting()
    backend.heartbeat()
    cluster.backends = append(cluster.backends, backend)
}

// 移除后端服务
func (cluster *Cluster) Remove(addr string) error {
    index := cluster.indexOf(addr)
    if index != - 1 {
        return BackendNotFound
    }
    heartStop <- cluster.backends[index].Addr
    cluster.rwMutex.Lock()
    defer cluster.rwMutex.Unlock()
    cluster.backends = append(cluster.backends[:index], cluster.backends[index + 1:]...)
    return nil
}

// 更新后端服务
func (cluster *Cluster)Update(backend *Backend) {
    index := cluster.indexOf(backend.Addr)
    if index == -1 {
        cluster.addBackend(backend)
        return
    }
    heartStop <- cluster.backends[index].Addr
    cluster.rwMutex.Lock()
    defer cluster.rwMutex.Unlock()
    backend.getDefaultSetting()
    backend.heartbeat()
    cluster.backends[index] = backend
}

func (cluster *Cluster)Balance() (backend *Backend, err error) {
    if cluster.backends.Len() < 1 {
        return nil, BackendServiceNotAvailable
    }
    sort.Sort(cluster.backends)
    backend = cluster.backends[0]
    if backend.Status != BackendUp {
        return nil, BackendServiceNotAvailable
    }
    return backend, nil
}

func (cluster *Cluster)indexOf(addr string) (index int) {
    index = -1
    for i, l := 0, len(cluster.backends); i < l; i ++ {
        if cluster.backends[i].Addr == addr {
            return i
        }
    }
    return index
}

// 添加一个集群
func (clusterGroup *ClusterGroup)Add(cluster *Cluster) error {
    index := clusterGroup.indexOf(cluster.Name)
    if index != -1 {
        return ClusterAlreadyExist
    }
    clusterGroup.rwMutex.Lock()
    defer clusterGroup.rwMutex.Unlock()
    clusterGroup.clusters = append(clusterGroup.clusters, cluster)
    return nil
}

// 移除一个集群
func (clusterGroup *ClusterGroup)Remove(clusterName string) error {
    index := clusterGroup.indexOf(clusterName)
    if index == -1 {
        return ClusterNotFound
    }
    clusterGroup.rwMutex.Lock()
    defer clusterGroup.rwMutex.Unlock()
    clusterGroup.clusters = append(clusterGroup.clusters[:index], clusterGroup.clusters[index + 1:]...)
    return nil
}

// 获取一个集群
func (clusterGroup *ClusterGroup)Get(clusterName string) (has bool, cluster *Cluster) {
    index := clusterGroup.indexOf(clusterName)
    if index == -1 {
        return false, cluster
    }
    return true, clusterGroup.clusters[index]
}

// 更新集群
func (clusterGroup *ClusterGroup)Update(cluster *Cluster) {
    index := clusterGroup.indexOf(cluster.Name)
    clusterGroup.rwMutex.Lock()
    defer clusterGroup.rwMutex.Unlock()
    if index == -1 {
        clusterGroup.clusters = append(clusterGroup.clusters, cluster)
        return
    }
    clusterGroup.clusters[index] = cluster
    return
}

func (clusterGroup *ClusterGroup)Clusters() []*Cluster {
    return clusterGroup.clusters
}

// 获取集群索引
func (clusterGroup *ClusterGroup)indexOf(clusterName string) (index int) {
    index = -1
    for i, l := 0, len(clusterGroup.clusters); i < l; i++ {
        if clusterGroup.clusters[i].Name == clusterName {
            return i
        }
    }
    return index
}