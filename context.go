package gateway

import (
	"encoding/json"
	"fmt"
	"goodsogood/gateway/render"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/context"
)

const abortIndex int8 = math.MaxInt8 / 2

type (
	// ExecInfo . 执行信息
	ExecInfo struct {
		BackendADDR string  `json:"addr"`
		BackendURI  string  `json:"uri"`
		Success     bool    `json:"success"`
		ExecTime    float64 `json:"execTime"`
	}
	// Context .
	Context struct {
		writermem responseWriter
		Request   *http.Request
		Writer    ResponseWriter
		routeInfo RouteInfo
		index     int8
		engine    *Engine
		responses []combineResponse

		ExecInfoGroup []ExecInfo

		handlers HandlesChain
		Keys     map[string]interface{}
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

// Next . 继续执行
func (c *Context) Next() {
	c.index++
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index].Handle(c)
	}
}

// IsAborted . 是否终止
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// Abort . 终止
func (c *Context) Abort() {
	c.index = abortIndex
}

// Set . 设置值
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

// Get . 获取值
func (c *Context) Get(key string) (value interface{}, exists bool) {
	if c.Keys != nil {
		value, exists = c.Keys[key]
	}
	return
}

// Query .
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

// PostForm
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

func (c *Context) RouteInfo() RouteInfo {
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

// JSONP 渲染
func (c *Context) JSONP(code int, obj interface{}, callback string) {
	c.Status(code)
	if err := render.WriteJSONP(c.Writer, obj, callback); err != nil {
		panic(err)
	}
}

// JSON 渲染
func (c *Context) JSON(code int, obj interface{}) {
	c.Status(code)
	if err := render.WriteJSON(c.Writer, obj); err != nil {
		panic(err)
	}
}

type H map[string]interface{}

// Render .
func (c *Context) Render(code int, obj interface{}) {
	c.Status(code)
	if obj == nil {
		switch len(c.responses) {
		case 1:
			if c.responses[0].Error != nil {
				obj = c.responses[0].Error
			} else {
				res := H{}
				if err := json.Unmarshal(c.responses[0].Response, &res); err != nil {
					obj = string(c.responses[0].Response)
				} else {
					obj = res
				}
			}
		default:
			combine := H{}
			// 合并返回
			for i, l := 0, len(c.responses); i < l; i++ {
				if c.responses[i].Error != nil {
					combine[c.responses[i].Attr] = c.responses[i].Error
				} else {
					res := H{}
					if err := json.Unmarshal(c.responses[i].Response, &res); err != nil {
						combine[c.responses[i].Attr] = string(c.responses[i].Response)
					} else {
						combine[c.responses[i].Attr] = res
					}
				}
			}
			obj = combine
		}
	}
	// debug
	if c.Query("debug") == "true" {
		obj = H{
			"exec":     c.ExecInfoGroup,
			"response": obj,
		}
	}
	if callback := c.Query("callback"); len(callback) > 0 {
		c.JSONP(code, obj, callback)
	} else {
		c.JSON(code, obj)
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
