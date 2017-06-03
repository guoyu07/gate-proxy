package gateway

import (
	"sort"
	"sync"
)

type (
	Cluster struct {
		// 集群名称
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		// 后端服务
		backends BackendGroup

		rwMutex sync.RWMutex
	}
	ClusterGroup struct {
		rwMutex  sync.RWMutex
		clusters []*Cluster
	}
)

var heartStop = make(chan string)

// Backends . 获取所有的后端服务
func (cluster *Cluster) Backends() BackendGroup {
	return cluster.backends
}

// Add . 增加一个后端服务
func (cluster *Cluster) Add(backend *Backend) error {
	index := cluster.indexOf(backend.Addr)
	if index != -1 {
		return BackendAlreadyExist
	}
	cluster.addBackend(backend)
	return nil
}

func (cluster *Cluster) addBackend(backend *Backend) {
	cluster.rwMutex.Lock()
	defer cluster.rwMutex.Unlock()
	backend.getDefaultSetting()
	if !backend.HeartDisabled {
		backend.heartbeat()
	}
	cluster.backends = append(cluster.backends, backend)
}

// Remove . 移除后端服务
func (cluster *Cluster) Remove(addr string) error {
	index := cluster.indexOf(addr)
	if index == -1 {
		return BackendNotFound
	}
	if !cluster.backends[index].HeartDisabled {
		heartStop <- cluster.backends[index].Addr
	}
	cluster.rwMutex.Lock()
	defer cluster.rwMutex.Unlock()
	cluster.backends = append(cluster.backends[:index], cluster.backends[index+1:]...)
	return nil
}

// Update . 更新后端服务
func (cluster *Cluster) Update(backend *Backend) {
	index := cluster.indexOf(backend.Addr)
	if index == -1 {
		cluster.addBackend(backend)
		return
	}
	heartStop <- cluster.backends[index].Addr
	cluster.rwMutex.Lock()
	defer cluster.rwMutex.Unlock()
	backend.getDefaultSetting()
	if !backend.HeartDisabled {
		backend.heartbeat()
	}
	cluster.backends[index] = backend
}

// Balance . 负载均衡
func (cluster *Cluster) Balance() (backend *Backend, err error) {
	if cluster.backends.Len() < 1 {
		return nil, BackendServiceNotAvailable
	}
	sort.Sort(cluster.backends)
	for i, l := 0, cluster.backends.Len(); i < l; i++ {
		if cluster.backends[i].Status == BackendUp {
			return cluster.backends[i], nil
		}
	}
	return nil, BackendServiceNotAvailable
}

func (cluster *Cluster) indexOf(addr string) (index int) {
	index = -1
	for i, l := 0, len(cluster.backends); i < l; i++ {
		if cluster.backends[i].Addr == addr {
			return i
		}
	}
	return index
}

// Add . 添加一个集群
func (clusterGroup *ClusterGroup) Add(cluster *Cluster) error {
	index := clusterGroup.indexOf(cluster.Name)
	if index != -1 {
		return ClusterAlreadyExist
	}
	clusterGroup.rwMutex.Lock()
	defer clusterGroup.rwMutex.Unlock()
	clusterGroup.clusters = append(clusterGroup.clusters, cluster)
	return nil
}

// Remove . 移除一个集群
func (clusterGroup *ClusterGroup) Remove(clusterName string) error {
	index := clusterGroup.indexOf(clusterName)
	if index == -1 {
		return ClusterNotFound
	}
	clusterGroup.rwMutex.Lock()
	defer clusterGroup.rwMutex.Unlock()
	clusterGroup.clusters = append(clusterGroup.clusters[:index], clusterGroup.clusters[index+1:]...)
	return nil
}

// Get . 获取一个集群
func (clusterGroup *ClusterGroup) Get(clusterName string) (has bool, cluster *Cluster) {
	index := clusterGroup.indexOf(clusterName)
	if index == -1 {
		return false, cluster
	}
	return true, clusterGroup.clusters[index]
}

// Update . 更新集群
func (clusterGroup *ClusterGroup) Update(cluster *Cluster) {
	index := clusterGroup.indexOf(cluster.Name)
	clusterGroup.rwMutex.Lock()
	defer clusterGroup.rwMutex.Unlock()
	if index == -1 {
		clusterGroup.clusters = append(clusterGroup.clusters, cluster)
		return
	}
	// copy info
	clusterGroup.clusters[index].Description = cluster.Description
	return
}

// Clusters . 集群列表
func (clusterGroup *ClusterGroup) Clusters() []*Cluster {
	return clusterGroup.clusters
}

// 获取集群索引
func (clusterGroup *ClusterGroup) indexOf(clusterName string) (index int) {
	index = -1
	for i, l := 0, len(clusterGroup.clusters); i < l; i++ {
		if clusterGroup.clusters[i].Name == clusterName {
			return i
		}
	}
	return index
}
