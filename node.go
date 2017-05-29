package gateway

import (
    "net/url"
    "fmt"
    "goodsogood/errors"
    "net/http"
    "bytes"
    "time"
    "sync/atomic"
    "sync"
    "io/ioutil"
)

// ParamFrom .
type ParamFrom int

const (
    // ParamFromHeader .
    ParamFromHeader = iota
    // ParamFromQuery .
    ParamFromQuery
    // ParamFromBody
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
    Param struct {
        Attr     string   `json:"attr"`     // 属性名
        From     ParamFrom `json:"from"`    // 来源
        To       ParamFrom `json:"to"`
        ToName   string   `json:"toName"`
        Required bool     `json:"required"` // 是否必要
    }
    Node struct {
        Attr       string `json:"attr"`
        Cluster    string `json:"cluster"`
        Rewrite    string `json:"rewrite"`
        ParamGroup []Param `json:"paramGroup"`
    }
    ParseParam struct {
        Header map[string]string
        Values url.Values
    }
)

// 解析参数
func (node Node)parse(ctx *Context) (parseParam ParseParam, err error) {
    parseParam = ParseParam{
        Header : make(map[string]string),
        Values: make(url.Values),
    }
    for i, l := 0, len(node.ParamGroup); i < l; i++ {
        param := node.ParamGroup[i]
        switch param.From {
        case ParamFromHeader:
            val := ctx.Request.Header.Get(param.Attr)
            if len(val) < 1 && param.Required {
                return parseParam, errors.New(-9004, fmt.Sprintf("Attr %s Is Required", param.Attr))
            }
            parseParam = parseParam.set(param, val)
        case ParamFromQuery:
            val, ok := ctx.Request.URL.Query()[param.Attr]
            if !ok && param.Required {
                return parseParam, errors.New(-9004, fmt.Sprintf("Attr %s Is Required", param.Attr))
            }
            parseParam = parseParam.set(param, val[0])
        }
    }
    return parseParam, nil
}

func (parseParam ParseParam)set(param Param, val string) ParseParam {
    switch param.To {
    case ParamFromHeader:
        parseParam.Header[param.ToName] = val
    case ParamFromQuery:
        parseParam.Values.Set(param.ToName, val)
    }
    return parseParam
}

var step = -1
// 执行
func (node Node)Do(ctx *Context, wg *sync.WaitGroup) {
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
        parseParam.Values.Encode(),
    )
    execInfo := ExecInfo{
        BackendADDR:backend.Addr,
        BackendURI:uri,
        Success:true,
    }
    req, err := http.NewRequest(ctx.routeInfo.Method, uri, ctx.Request.Body)

    // set header
    for k, v := range parseParam.Header {
        req.Header.Set(k, v)
    }
    req.Header.Set("Gate-Cluster", cluster.Name)
    req.Header.Set("X-Forwarded-For", ctx.ClientIP())
    client := ctx.engine.Client()
    client.Timeout = time.Second * time.Duration(backend.Timeout)
    atomic.AddUint64(&backend.Waiting, 1)
    now := time.Now()
    res, err := client.Do(req)
    ctx.engine.Release(client)
    execInfo.ExecTime = float64(time.Since(now).Nanoseconds() / 1000000)
    atomic.AddUint64(&backend.Waiting, ^uint64(-step - 1))
    response.Response, _ = ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        execInfo.Success = false
        response.Error = BackendServiceError
    }
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