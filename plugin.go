package gateway

type Plugin interface {
    Name() string
    Version() string
    Handle(ctx *Context)
}

type PluginInfo struct {
    Name    string
    Version string
}
