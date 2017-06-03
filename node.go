package gateway

import (
	"bytes"
	"fmt"
	"goodsogood/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ParamFrom .
type ParamFrom int

const (
	// ParamFromHeader .
	ParamFromHeader = iota + 1
	// ParamFromQuery .
	ParamFromQuery
	// ParamFromBody .
	ParamFromBody
)

func (paramFrom ParamFrom) String() string {
	switch paramFrom {
	case ParamFromHeader:
		return "Header"
	case ParamFromQuery:
		return "Query"
	case ParamFromBody:
		return "Body"
	}
	return "Unknown"
}

type (
	// Param .
	Param struct {
		Attr       string         `json:"attr"` // 属性名
		From       ParamFrom      `json:"from"` // 来源
		To         ParamFrom      `json:"to"`
		ToName     string         `json:"toName"`
		Required   bool           `json:"required"`   // 是否必要
		Validation string         `json:"validation"` // 校验规则
		rule       *regexp.Regexp // 正则规则
	}
	// Node .
	Node struct {
		Attr       string  `json:"attr"`
		Cluster    string  `json:"cluster"`
		Rewrite    string  `json:"rewrite"`
		ParamGroup []Param `json:"paramGroup"`
	}
	// ParseParam .
	ParseParam struct {
		header map[string]string
		query  url.Values
		body   url.Values
	}
)

// 解析参数
func (node Node) parse(ctx *Context) (parseParam ParseParam, err error) {
	parseParam = ParseParam{
		header: make(map[string]string),
		query:  make(url.Values),
		body:   make(url.Values),
	}
	for i, l := 0, len(node.ParamGroup); i < l; i++ {
		param := node.ParamGroup[i]
		var (
			val string
		)
		switch param.From {
		case ParamFromHeader:
			val = ctx.Request.Header.Get(param.Attr)
		case ParamFromQuery:
			val = ctx.Request.URL.Query().Get(param.Attr)
		case ParamFromBody:
			val = ctx.PostForm(param.Attr)
		}
		if len(val) < 1 && param.Required {
			return parseParam, errors.New(-9004, fmt.Sprintf("Attr %s Is Required", param.Attr))
		}
		if param.rule != nil && !param.rule.MatchString(val) {
			return parseParam, errors.New(-9004, fmt.Sprintf("Attr %s Is't Validation", param.Attr))
		}
		parseParam = parseParam.set(param, val)
	}
	return parseParam, nil
}

func (parseParam ParseParam) set(param Param, val string) ParseParam {
	switch param.To {
	case ParamFromHeader:
		parseParam.header[param.ToName] = val
	case ParamFromQuery:
		parseParam.query.Add(param.ToName, val)
	case ParamFromBody:
		parseParam.body.Add(param.ToName, val)
	}
	return parseParam
}

var step = -1

// Do . 执行
func (node Node) Do(ctx *Context, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	has, cluster := ctx.engine.clusters.Get(node.Cluster)
	response := combineResponse{
		Attr: node.Attr,
	}
	if !has {
		response.Error = ClusterNotFound
		ctx.responses = append(ctx.responses, response)
		return
	}
	backend, err := cluster.Balance()
	if err != nil {
		response.Error = err
		ctx.responses = append(ctx.responses, response)
		return
	}
	parseParam, err := node.parse(ctx)
	if err != nil {
		response.Error = err
		ctx.responses = append(ctx.responses, response)
		return
	}
	uri := uriEncode(backend.Schema,
		"://",
		backend.Addr,
		node.Rewrite,
		"?",
		parseParam.query.Encode(),
	)
	execInfo := ExecInfo{
		BackendADDR: backend.Addr,
		BackendURI:  uri,
		Success:     true,
	}
	req, err := http.NewRequest(ctx.routeInfo.Method, uri, strings.NewReader(parseParam.body.Encode()))
	// set header
	for k, v := range parseParam.header {
		req.Header.Set(k, v)
	}
	setDefaultHeader(ctx.routeInfo.Method, req)
	req.Header.Set("Gate-Cluster", cluster.Name)
	req.Header.Set("X-Forwarded-For", ctx.ClientIP())
	client := ctx.engine.Client()
	defer ctx.engine.Release(client)
	client.Timeout = time.Second * time.Duration(backend.Timeout)
	atomic.AddUint64(&backend.Waiting, 1)
	atomic.AddUint64(&backend.QPS, 1)
	now := time.Now()
	res, err := client.Do(req)
	execInfo.ExecTime = float64(time.Since(now).Nanoseconds() / 1000000)
	atomic.AddUint64(&backend.Waiting, ^uint64(-step-1))
	if err != nil {
		execInfo.Success = false
		response.Error = BackendServiceError
		goto WALK
	}
	defer res.Body.Close()
	response.Response, err = ioutil.ReadAll(res.Body)
	if err != nil {
		execInfo.Success = false
		response.Error = BackendServiceError
	}
WALK:
	ctx.ExecInfoGroup = append(ctx.ExecInfoGroup, execInfo)
	ctx.responses = append(ctx.responses, response)
	return
}

func uriEncode(vals ...string) string {
	var buffer bytes.Buffer
	for i, l := 0, len(vals); i < l; i++ {
		buffer.WriteString(vals[i])
	}
	return buffer.String()
}

func setDefaultHeader(method string, req *http.Request) {
	switch method {
	case "POST":
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
}
