package etcdcenter

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Registrar interface {
	// 注册服务到注册中心
	Registry() error
	// 从注册中心注销服务
	Deregistry() error
}

func newEtcdRegistrar(cli *clientv3.Client, serverName, host string, port, ttl int) Registrar {
	return &EtcdRegistrar{
		cli: cli,
		key: fmt.Sprintf("%s-%s-%d", serverName, host, port),
		val: fmt.Sprintf("%s:%d", host, port),
		ttl: int64(ttl),
	}
}

type EtcdRegistrar struct {
	cli     *clientv3.Client
	ctx     context.Context
	key     string
	val     string
	ttl     int64
	leaseid clientv3.LeaseID
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
	err = s.keepAlive()
	if err != nil {
		return err
	}
	return nil
}

// 续约
func (s *EtcdRegistrar) keepAlive() error {
	rspChan, err := s.cli.KeepAlive(s.ctx, s.leaseid)
	if err != nil {
		return err
	}

	go func() {
		for range rspChan {
			for {
				_, err := s.cli.KeepAliveOnce(context.TODO(), s.leaseid)
				if err != nil {
					log.Errorf("etcd keep alive once err:%s", err)
					break
				}
			}
		}
	}()
	return nil
}

func (s *EtcdRegistrar) Deregistry() error {
	if _, err := s.cli.Revoke(s.ctx, s.leaseid); err != nil {
		return err
	}
	return s.cli.Close()
}
