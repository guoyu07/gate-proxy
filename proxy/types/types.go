package types

type ClusterInfo struct {
    Name string `json:"name" storm:"id"`
}

type BackendInfo struct {
    // 访问地址
    Addr              string `json:"addr" storm:"id"`
    ClusterName       string `json:"clusterName"`
    Schema            string `json:"schema"`
    // 心跳地址
    HeartPath         string `json:"heartPath"`
    // 心跳返回校验
    HeartResponseBody string `json:"heartResponseBody"`
    // 心跳时间
    HeartDuration     int64 `json:"heartDuration"`
    // 请求超时
    Timeout           int64 `json:"Timeout"`
    // 最大qps
    MaxQPS            uint64 `json:"maxQPS"`
}