package etcdcenter

import (
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Registrar interface {
	// 注册服务到注册中心
	Registry() error
	// 从注册中心注销服务
	Deregistry() error
}

func newEtcdRegistrar(cli *clientv3.Client, serverName, host string, port int, opts ...Option) Registrar {
	o := newOpt(opts...)
	return &EtcdRegistrar{
		cli:    cli,
		key:    fmt.Sprintf("%s-%s-%d", serverName, host, port),
		val:    fmt.Sprintf("%s:%d", host, port),
		ttl:    o.ttl,
		logger: o.logger,
	}
}

type EtcdRegistrar struct {
	cli     *clientv3.Client
	ctx     context.Context
	key     string
	val     string
	ttl     int64
	leaseid clientv3.LeaseID
	logger  func(error)
}

func (s *EtcdRegistrar) Registry() error {
	s.ctx = context.Background()
	// 设置租约
	resp, err := s.cli.Grant(s.ctx, s.ttl)
	if err != nil {
		return err
	}
	s.leaseid = resp.ID
	_, err = s.cli.Put(context.Background(), s.key, s.val, clientv3.WithLease(s.leaseid))
	if err != nil {
		return err
	}

	rspChan, err := s.cli.KeepAlive(s.ctx, s.leaseid)
	if err != nil {
		return err
	}
	go s.keepAlive(rspChan)
	return nil
}

// 续约
func (s *EtcdRegistrar) keepAlive(rspChan <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case <-s.ctx.Done():
			s.logger(fmt.Errorf("keep alive done for lease id: %d", s.leaseid))
			return
		case _, ok := <-rspChan:
			if !ok {
				s.logger(fmt.Errorf("keep alive exit, lease id: %d", s.leaseid))
				return
			}
		}
	}
}

func (s *EtcdRegistrar) Deregistry() error {
	_, err := s.cli.Delete(s.ctx, s.key)
	if s.cli.Lease != nil {
		_ = s.cli.Close()
	}
	return err
}
