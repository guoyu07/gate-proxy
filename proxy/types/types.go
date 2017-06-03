package types

type ClusterInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BackendInfo struct {
	// 访问地址
	Addr          string `json:"addr"`
	ClusterName   string `json:"clusterName"`
	Schema        string `json:"schema"`
	HeartDisabled bool   `json:"heartDisabled"`
	// 心跳地址
	HeartPath string `json:"heartPath"`
	// 心跳返回校验
	HeartResponseBody string `json:"heartResponseBody"`
	// 心跳时间
	HeartDuration int64 `json:"heartDuration"`
	// 请求超时
	Timeout int64 `json:"Timeout"`
	// 最大qps
	MaxQPS uint64 `json:"maxQPS"`
}
