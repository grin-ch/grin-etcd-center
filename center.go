package etcdcenter

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type IEtcdCenter interface {
	Registrar(serverName, host string, port, ttl int) Registrar
	Builder() resolver.Builder
}

func NewEtcdCenter(endpoints []string, timeout int) (IEtcdCenter, error) {
	cli, err := clientv3.New(
		clientv3.Config{
			Endpoints:            endpoints,
			DialTimeout:          time.Second * time.Duration(timeout),
			DialKeepAliveTimeout: time.Second * time.Duration(timeout),
		},
	)
	if err != nil {
		return nil, err
	}
	return &EtcdCenter{cli: cli}, nil
}

type EtcdCenter struct {
	cli *clientv3.Client
}

func (c *EtcdCenter) Registrar(serverName, host string, port, ttl int) Registrar {
	return newEtcdRegistrar(c.cli, serverName, host, port, ttl)
}

func (c *EtcdCenter) Builder() resolver.Builder {
	return newEtcdBuilder(c.cli)
}
