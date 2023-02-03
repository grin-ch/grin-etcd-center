package etcdcenter

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type IEtcdCenter interface {
	Registrar(serverName, host string, port int, opts ...Option) Registrar
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

const (
	DefaultTTl int64 = 10
)

type opt struct {
	ttl    int64
	logger func(error)
}

func newOpt(opts ...Option) *opt {
	o := new(opt)
	o.ttl = DefaultTTl
	o.logger = func(error) {}
	for _, fn := range opts {
		fn(o)
	}
	return o
}

type Option func(*opt)

func WithTTL(ttl int64) Option {
	return func(o *opt) {
		if ttl > 0 {
			o.ttl = ttl
		}
	}
}

func WithLogger(fn func(error)) Option {
	return func(o *opt) {
		if fn != nil {
			o.logger = fn
		}
	}
}

type EtcdCenter struct {
	cli *clientv3.Client
}

func (c *EtcdCenter) Registrar(serverName, host string, port int, opts ...Option) Registrar {
	return newEtcdRegistrar(c.cli, serverName, host, port)
}

func (c *EtcdCenter) Builder() resolver.Builder {
	return newEtcdBuilder(c.cli)
}
