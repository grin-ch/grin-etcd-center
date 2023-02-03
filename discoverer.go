package etcdcenter

import (
	"context"
	"strings"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

func newEtcdBuilder(cli *clientv3.Client) resolver.Builder {
	return &EtcdBuilder{cli: cli}
}

type EtcdBuilder struct {
	cli *clientv3.Client
}

func (s *EtcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &etcdResolver{
		cli: s.cli,
		cc:  cc,
	}

	prefix := strings.Replace(target.URL.Path, "/", "", 1)
	r.search(prefix)
	go r.watch(prefix)
	return r, nil
}

func (s *EtcdBuilder) Scheme() string {
	return "etcd"
}

type etcdResolver struct {
	cli   *clientv3.Client
	cc    resolver.ClientConn
	addrs []resolver.Address
}

func (s *etcdResolver) search(prefix string) error {
	rsp, err := s.cli.Get(context.TODO(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, v := range rsp.Kvs {
		s.addrs = append(s.addrs, resolver.Address{Addr: string(v.Value)})
	}
	s.cc.UpdateState(resolver.State{Addresses: s.addrs})
	return nil
}

func (s *etcdResolver) watch(prefix string) {
	wchan := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for ch := range wchan {
		for _, v := range ch.Events {
			addr := string(v.Kv.Key)
			switch v.Type {
			case mvccpb.PUT:
				if !s.existAddr(addr) {
					s.addrs = append(s.addrs, resolver.Address{Addr: addr})
					s.cc.UpdateState(resolver.State{Addresses: s.addrs})
				}
			case mvccpb.DELETE:
				if s.removeAddr(addr) {
					s.cc.UpdateState(resolver.State{Addresses: s.addrs})
				}
			}
		}
	}
}

func (s *etcdResolver) existAddr(addr string) bool {
	for _, v := range s.addrs {
		if v.Addr == addr {
			return true
		}
	}
	return false
}

func (s *etcdResolver) removeAddr(addr string) bool {
	isRemove := false
	addrs := make([]resolver.Address, 0, len(s.addrs))
	for _, v := range s.addrs {
		if v.Addr == addr {
			isRemove = true
			continue
		}
		addrs = append(addrs, v)
	}
	s.addrs = addrs
	return isRemove
}

func (s *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {}
func (s *etcdResolver) Close()                                {}
