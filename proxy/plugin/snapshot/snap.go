package snapshot

import (
	"fmt"
	"goodsogood/gateway"
)

type Snapshot struct {
}

func (snap Snapshot) Name() string {
	return "snapshot"
}

func (snap Snapshot) Private() bool {
	return false
}

func (snap Snapshot) Version() string {
	return "0.1"
}

func (snap Snapshot) Handle(ctx *gateway.Context) {
	fmt.Println("Dispatch snap start")
	ctx.Next()
	fmt.Println("Dispatch snap finished. exec ", ctx.ExecInfoGroup)
}
