package gateway

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Status . 服务状态
type BackendStatus int

func (status BackendStatus) String() string {
	switch status {
	case BackendDown:
		return "下线"
	case BackendUp:
		return "上线"
	}
	return "未知"
}

const (
	// Down 下线
	BackendDown BackendStatus = iota + 1
	// Up 上线
	BackendUp
)

const (
	// DefaultHeartDurationInSeconds .
	DefaultHeartDurationInSeconds = 5
	// DefaultTimeoutInSeconds .
	DefaultTimeoutInSeconds = 3

	DefaultMaxFail = 3
)

type (
	Backend struct {
		// 访问协议 http|https
		Schema string `json:"schema"`
		// 访问地址
		Addr string `json:"addr"`
		// 是否开启心跳
		HeartDisabled bool `json:"heartDisabled"`
		// 心跳地址
		HeartPath string `json:"heartPath"`
		// 心跳返回校验
		HeartResponseBody string `json:"heartResponseBody"`
		// 心跳时间
		HeartDuration int64 `json:"heartDuration"`
		// 请求超时
		Timeout int64 `json:"Timeout"`
		// 最后检测时间
		LastHeartTime int64 `json:"lastHeartTime"`
		// 服务状态
		Status BackendStatus `json:"status"`
		// 最大qps
		MaxQPS uint64 `json:"maxQPS"`

		QPS         uint64 `json:"QPS"`
		MaxTime     uint64 `json:"maxTime"`
		AverageTime uint64 `json:"averageTime"`
		// Wait
		Waiting uint64 `json:"waiting"`

		// private
		httpClient     *http.Client
		heartFailCount uint64
	}
	BackendGroup []*Backend
)

// 设置默认值
func (backend *Backend) getDefaultSetting() *Backend {
	if backend.Timeout < 1 {
		backend.Timeout = DefaultTimeoutInSeconds
	}
	if backend.HeartDuration < 1 {
		backend.HeartDuration = DefaultHeartDurationInSeconds
	}
	if backend.HeartDisabled {
		backend.Status = BackendUp
	} else {
		backend.Status = BackendDown
	}
	return backend
}

func (backend *Backend) heartbeat() {
	ticker := time.NewTicker(time.Second * time.Duration(backend.HeartDuration))
	uri := fmt.Sprintf("%s://%s%s", backend.Schema, backend.Addr, backend.HeartPath)
	go func(uri string) {
		for {
			select {
			// 终止心跳监测
			case addr := <-heartStop:
				if addr == backend.Addr {
					return
				}
			case <-ticker.C:
				backend.LastHeartTime = time.Now().Unix()
				res, err := http.Get(uri)
				if err != nil {
					backend.heartFailCount++
					if backend.heartFailCount == DefaultMaxFail {
						// 移出上线队列
						if backend.Status == BackendUp {
							backend.Status = BackendDown
						}
					}
					continue
				}
				if backend.HeartResponseBody == "" {
					// 校验为空 校验状态
					if res.StatusCode != http.StatusOK {
						backend.heartFailCount++
						if backend.heartFailCount == DefaultMaxFail {
							// 移出上线队列
							if backend.Status == BackendUp {
								backend.Status = BackendDown
							}
						}
					} else {
						backend.heartFailCount = 0
						if backend.Status == BackendDown {
							backend.Status = BackendUp
						}
					}
				} else {
					defer res.Body.Close()
					resBody, _ := ioutil.ReadAll(res.Body)
					// 校验为空 校验状态
					if string(resBody) != backend.HeartResponseBody {
						backend.heartFailCount++
						if backend.heartFailCount == DefaultMaxFail {
							// 移出上线队列
							if backend.Status == BackendUp {
								backend.Status = BackendDown
							}
						}
					} else {
						backend.heartFailCount = 0
						if backend.Status == BackendDown {
							backend.Status = BackendUp
						}
					}
				}
			}
		}
	}(uri)
}

func (backendGroup BackendGroup) Len() int {
	return len(backendGroup)
}

func (backendGroup BackendGroup) Less(i, j int) bool {
	return backendGroup[i].Waiting/backendGroup[i].MaxQPS-(0-uint64(backendGroup[i].Status)) < backendGroup[j].Waiting/backendGroup[j].MaxQPS-(0-uint64(backendGroup[j].Status))
}

func (backendGroup BackendGroup) Swap(i, j int) {
	backendGroup[i], backendGroup[j] = backendGroup[j], backendGroup[i]
}
