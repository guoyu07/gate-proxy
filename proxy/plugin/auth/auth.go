package auth

import (
    "goodsogood/gateway"
    "fmt"
)

type Auth struct {

}

func (auth Auth)Name() string {
    return "auth"
}

func (auth Auth)Version() string {
    return "0.1"
}

func (auth Auth) Handle(ctx *gateway.Context) {
    fmt.Println("Dispatch Auth")
}