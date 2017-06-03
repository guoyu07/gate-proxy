package gateway

// Plugin .
type Plugin interface {
	Name() string
	Private() bool
	Version() string
	Handle(ctx *Context)
}

// PluginInfo
type PluginInfo struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
	Version string `json:"version"`
}
