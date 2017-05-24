package gateway

import (
    "testing"
    "fmt"
)

var route = NewRouteTable()
//
func TestRouteTable_Add(t *testing.T) {
    route.Add(RouteInfo{
        Name: "登录接口",
        Method:"POST",
        URL:"/login",
        Domain:"",
        NodeGroup: []Node{
            Node{
                Attr: "info",
                Cluster:"UserBaseCluster",
                Rewrite: "/user/login",
            },
        },
    })
}

func TestRouteTable_Update(t *testing.T) {
    route.Update(RouteInfo{
        Name: "登录接口",
        Method:"POST",
        URL:"/login",
        Domain:"1",
        NodeGroup: []Node{
            Node{
                Attr: "info",
                Cluster:"UserBaseCluster",
                Rewrite: "/user/login",
            },
        },
    })
}

func TestRouteTable_Get(t *testing.T) {
    route.Get("POST", "/login")
}

func TestRouteTable_Remove(t *testing.T) {
    route.Remove("POST", "login")
}

func TestRouteTable_Add2(t *testing.T) {
    for i := 1; i < 2; i++ {
        uri := fmt.Sprintf("/login/%d", i)
        rule := RouteInfo{
            Name: "登录接口",
            Method:"POST",
            URL:uri,
            Domain:"",
            NodeGroup: []Node{
                Node{
                    Attr: "info",
                    Cluster:"UserBaseCluster",
                    Rewrite: "/user/login",
                },
            },
        }
        go func() {
            route.Add(rule)
        }()
    }
}
func BenchmarkRouteTable_Add(b *testing.B) {
    for i := 1; i < b.N; i++ {
        uri := fmt.Sprintf("/login/%d", i)
        rule := RouteInfo{
            Name: "登录接口",
            Method:"POST",
            URL:uri,
            Domain:"",
            NodeGroup: []Node{
                Node{
                    Attr: "info",
                    Cluster:"UserBaseCluster",
                    Rewrite: "/user/login",
                },
            },
        }
        route.Add(rule)
    }
}
