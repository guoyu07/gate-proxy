package gateway

import (
    "net/http"
    "golang.org/x/net/context"
    "strings"
    "net/url"
    "fmt"
    "io"
    "encoding/json"
    "time"
)

type (
    ExecInfo struct {
        BackendADDR string `json:"addr"`
        BackendURI  string `json:"uri"`
        Success     bool `json:"success"`
        ExecTime    float64 `json:"execTime"`
    }
    Context struct {
        writermem     responseWriter
        Request       *http.Request
        Writer        ResponseWriter
        routeInfo     RouteInfo
        index         int8
        engine        *Engine
        responses     []combineResponse

        ExecInfoGroup []ExecInfo

        handlers      HandlesChain
        Keys          map[string]interface{}
    }
)

var _ context.Context = &Context{}

func (c *Context) reset() {
    c.Writer = &c.writermem
    c.index = -1
    c.routeInfo = RouteInfo{}
    c.responses = nil
    c.ExecInfoGroup = nil
    c.handlers = nil
    c.Keys = nil
}

func (c *Context) Next() {
    c.index++
    s := int8(len(c.handlers))
    for ; c.index < s; c.index++ {
        c.handlers[c.index].Handle(c)
    }
}


func (c *Context) Set(key string, value interface{}) {
    if c.Keys == nil {
        c.Keys = make(map[string]interface{})
    }
    c.Keys[key] = value
}

func (c *Context) Get(key string) (value interface{}, exists bool) {
    if c.Keys != nil {
        value, exists = c.Keys[key]
    }
    return
}

func (c *Context) Query(key string) string {
    value, _ := c.GetQuery(key)
    return value
}

func (c *Context) GetQuery(key string) (string, bool) {
    if values, ok := c.GetQueryArray(key); ok {
        return values[0], ok
    }
    return "", false
}

func (c *Context) QueryArray(key string) []string {
    values, _ := c.GetQueryArray(key)
    return values
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
    req := c.Request
    if values, ok := req.URL.Query()[key]; ok && len(values) > 0 {
        return values, true
    }
    return []string{}, false
}

func (c *Context) PostForm(key string) string {
    value, _ := c.GetPostForm(key)
    return value
}

func (c *Context) DefaultPostForm(key, defaultValue string) string {
    if value, ok := c.GetPostForm(key); ok {
        return value
    }
    return defaultValue
}

func (c *Context) GetPostForm(key string) (string, bool) {
    if values, ok := c.GetPostFormArray(key); ok {
        return values[0], ok
    }
    return "", false
}

func (c *Context) PostFormArray(key string) []string {
    values, _ := c.GetPostFormArray(key)
    return values
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
    req := c.Request
    req.ParseForm()
    req.ParseMultipartForm(32 << 20) // 32 MB
    if values := req.PostForm[key]; len(values) > 0 {
        return values, true
    }
    if req.MultipartForm != nil && req.MultipartForm.File != nil {
        if values := req.MultipartForm.Value[key]; len(values) > 0 {
            return values, true
        }
    }
    return []string{}, false
}

func (c *Context) ClientIP() string {
    clientIP := strings.TrimSpace(c.requestHeader("X-Real-Ip"))
    if len(clientIP) > 0 {
        return clientIP
    }
    clientIP = c.requestHeader("X-Forwarded-For")
    if index := strings.IndexByte(clientIP, ','); index >= 0 {
        clientIP = clientIP[0:index]
    }
    clientIP = strings.TrimSpace(clientIP)
    return clientIP
}

func (c *Context)RouteInfo() RouteInfo {
    return c.routeInfo
}

func (c *Context) ContentType() string {
    return filterFlags(c.requestHeader("Content-Type"))
}

func filterFlags(content string) string {
    for i, char := range content {
        if char == ' ' || char == ';' {
            return content[:i]
        }
    }
    return content
}
func (c *Context) requestHeader(key string) string {
    if values, _ := c.Request.Header[key]; len(values) > 0 {
        return values[0]
    }
    return ""
}

func (c *Context) Status(code int) {
    c.writermem.WriteHeader(code)
}

func (c *Context) Header(key, value string) {
    if len(value) == 0 {
        c.Writer.Header().Del(key)
    } else {
        c.Writer.Header().Set(key, value)
    }
}

func (c *Context) SetCookie(
name string,
value string,
maxAge int,
path string,
domain string,
secure bool,
httpOnly bool,
) {
    if path == "" {
        path = "/"
    }
    http.SetCookie(c.Writer, &http.Cookie{
        Name:     name,
        Value:    url.QueryEscape(value),
        MaxAge:   maxAge,
        Path:     path,
        Domain:   domain,
        Secure:   secure,
        HttpOnly: httpOnly,
    })
}

func (c *Context) Cookie(name string) (string, error) {
    cookie, err := c.Request.Cookie(name)
    if err != nil {
        return "", err
    }
    val, _ := url.QueryUnescape(cookie.Value)
    return val, nil
}

var plainContentType = []string{"text/plain; charset=utf-8"}
var jsonContentType = []string{"application/json; charset=utf-8"}

func writeContentType(w http.ResponseWriter, value []string) {
    header := w.Header()
    if val := header["Content-Type"]; len(val) == 0 {
        header["Content-Type"] = value
    }
}

func (c *Context)Render() {
    switch len(c.responses) {
    case 0:
        c.JSON(BackendServiceNotAvailable)
    case 1:
        if c.responses[0].Error != nil {
            c.JSON(c.responses[0].Error)
        } else {
            defer c.responses[0].Response.Body.Close()
            c.JSONWithReader(c.responses[0].Response.Body)
        }
    default:
        c.MulitiJSONWithReader(c.responses)
    }
}

func (c *Context) isJsonp() (callback string, has bool) {
    callback = c.Query("callback")
    if len(callback) > 0 {
        return callback, true
    }
    return callback, false
}

func (c *Context)isDebug() bool {
    debug := c.Query("debug")
    if debug == "true" {
        return true
    }
    return false
}

// JSONWithReader .
func (c *Context) JSONWithReader(reader io.Reader) {
    c.Status(http.StatusOK)
    callback, has := c.isJsonp()
    if has {
        writeContentType(c.Writer, plainContentType)
        c.Writer.WriteString(callback)
        c.Writer.WriteString(`(`)
    } else {
        writeContentType(c.Writer, jsonContentType)
    }
    if c.isDebug() {
        c.Writer.WriteString("{")
        c.Writer.WriteString("\"exec\":")
        json.NewEncoder(c.Writer).Encode(c.ExecInfoGroup)
        c.Writer.WriteString(",\"response\":")
        io.Copy(c.Writer, reader)
        c.Writer.WriteString("}")
    } else {
        io.Copy(c.Writer, reader)
    }
    if has {
        c.Writer.WriteString(")")
    }
}

// MulitiJSONWithReader . 支持JSONP
func (c *Context) MulitiJSONWithReader(combines []combineResponse) {
    c.Status(http.StatusOK)
    callback, has := c.isJsonp()
    if has {
        writeContentType(c.Writer, plainContentType)
        c.Writer.WriteString(callback)
        c.Writer.WriteString("(")
    } else {
        writeContentType(c.Writer, jsonContentType)
    }
    // combine
    c.Writer.WriteString("{")
    for i, l := 0, len(combines); i < l; i++ {
        c.Writer.WriteString("\"")
        c.Writer.WriteString(combines[i].Attr)
        c.Writer.WriteString("\":")
        if combines[i].Error != nil {
            if err := json.NewEncoder(c.Writer).Encode(combines[i].Error); err != nil {
                panic(err)
            }
        } else {
            defer combines[i].Response.Body.Close()
            io.Copy(c.Writer, combines[i].Response.Body)
        }
        if i < l - 1 {
            c.Writer.WriteString(",")
        }
    }
    if c.isDebug() {
        c.Writer.WriteString("\"exec\":")
        json.NewEncoder(c.Writer).Encode(c.ExecInfoGroup)
    }
    c.Writer.WriteString("}")
    if has {
        c.Writer.WriteString(")")
    }
}

// JSON . 返回json
func (c *Context) JSON(obj interface{}) {
    c.Status(http.StatusOK)
    callback, has := c.isJsonp()
    if has {
        writeContentType(c.Writer, plainContentType)
        c.Writer.WriteString(callback)
        c.Writer.WriteString("(")
    } else {
        writeContentType(c.Writer, jsonContentType)
    }
    if c.isDebug() {
        c.Writer.WriteString("{")
        c.Writer.WriteString("\"exec\":")
        json.NewEncoder(c.Writer).Encode(c.ExecInfoGroup)
        c.Writer.WriteString(",\"response\":")
        json.NewEncoder(c.Writer).Encode(obj)
        c.Writer.WriteString("}")
    } else {
        json.NewEncoder(c.Writer).Encode(obj)
    }
    if has {
        c.Writer.WriteString(")")
    }
}

func (c *Context) String(code int, format string, values ...interface{}) {
    c.Status(code)
    writeContentType(c.Writer, plainContentType)
    if len(values) > 0 {
        fmt.Fprintf(c.Writer, format, values...)
    } else {
        io.WriteString(c.Writer, format)
    }
}

// Redirect returns a HTTP redirect to the specific location.
func (c *Context) Redirect(code int, location string) {
    if (code < 300 || code > 308) && code != 201 {
        panic(fmt.Sprintf("Cannot redirect with status code %d", code))
    }
    http.Redirect(c.Writer, c.Request, location, code)
}

func (c *Context) Stream(step func(w io.Writer) bool) {
    w := c.Writer
    clientGone := w.CloseNotify()
    for {
        select {
        case <-clientGone:
            return
        default:
            keepOpen := step(w)
            w.Flush()
            if !keepOpen {
                return
            }
        }
    }
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
    return
}

func (c *Context) Done() <-chan struct{} {
    return nil
}

func (c *Context) Err() error {
    return nil
}

func (c *Context) Value(key interface{}) interface{} {
    if key == 0 {
        return c.Request
    }
    if keyAsString, ok := key.(string); ok {
        val, _ := c.Get(keyAsString)
        return val
    }
    return nil
}