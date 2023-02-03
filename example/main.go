package main

import (
	"context"
	"fmt"
	"time"

	etcdcenter "github.com/grin-ch/grin-etcd-center"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

var (
	server_name = "grin-server"
	server_host = "192.168.1.103"
	server_port = 8082
)

func main() {
	etcdCenter, err := etcdcenter.NewEtcdCenter([]string{"http://192.168.1.103:2379"}, 10)
	if err != nil {
		panic(err)
	}
	go tryClient()
	r := etcdCenter.Registrar(server_name, server_host, server_port)
	time.Sleep(3 * time.Second)
	err = r.Registry()
	if err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Second)
	err = r.Deregistry()
	if err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Second)
}

func tryClient() {
	etcdCenter, err := etcdcenter.NewEtcdCenter([]string{"http://192.168.1.103:2379"}, 10)
	if err != nil {
		panic(err)
	}
	connect := etcdConncter(etcdCenter.Builder())
	conn, err := connect(server_name)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	for {
		err := conn.Invoke(context.Background(), "test", nil, nil)
		if err != nil {
			fmt.Println("err:", err.Error())
		}
		time.Sleep(1 * time.Second)
	}
}

type connecter func(string) (*grpc.ClientConn, error)

func etcdConncter(builder resolver.Builder) connecter {
	return func(s string) (*grpc.ClientConn, error) {
		addr := fmt.Sprintf("etcd:///%s", s)
		return grpc.Dial(addr,
			grpc.WithResolvers(builder),
			grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy":"%s"}`, roundrobin.Name)),
			grpc.WithKeepaliveParams(
				keepalive.ClientParameters{
					Time:                10 * time.Second,
					Timeout:             100 * time.Millisecond,
					PermitWithoutStream: true},
			),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
}
