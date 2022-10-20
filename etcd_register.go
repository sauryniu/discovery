/**
 * grpc etcd注册器
 *
 * @title etcd_register
 * @projectName discovery
 * @author niuk
 * @date 2022/8/25 16:35
 */

package discovery

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type etcdRegisterImpl struct {
	etcdAddrs   []string
	dialTimeOut time.Duration
	leasesId    clientv3.LeaseID

	log           *logrus.Logger
	ttl           int64
	cli           *clientv3.Client
	node          ServiceNode
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	closeChan     chan struct{}
}

func (e *etcdRegisterImpl) Register(node ServiceNode, ttl time.Duration) error {
	var err error
	if e.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   e.etcdAddrs,
		DialTimeout: e.dialTimeOut,
	}); err != nil {
		return err
	}
	e.node = node
	e.ttl = int64(ttl / time.Second)
	e.closeChan = make(chan struct{})

	if err = e.register(); err != nil {
		return err
	}

	go e.keepAlive()
	return nil
}

func (e *etcdRegisterImpl) Unregister() {
	e.closeChan <- struct{}{}
}

func (e *etcdRegisterImpl) register() error {
	ctx, cancel := context.WithTimeout(context.Background(), e.dialTimeOut)
	defer cancel()
	rsp, err := e.cli.Grant(ctx, e.ttl)
	if err != nil {
		return err
	}
	e.leasesId = rsp.ID

	if e.keepAliveChan, err = e.cli.KeepAlive(context.Background(), e.leasesId); err != nil {
		return err
	}

	data, err := json.Marshal(&e.node)
	if err != nil {
		return err
	}
	pCtx, pCancel := context.WithTimeout(context.Background(), e.dialTimeOut)
	defer pCancel()
	_, err = e.cli.Put(pCtx, e.node.BuildPath(), string(data), clientv3.WithLease(e.leasesId))
	e.log.Infof("register to etcd, path:%s, value：%s", e.node.BuildPath(), string(data))
	return err
}

func (e *etcdRegisterImpl) keepAlive() {
	ticker := time.NewTicker(time.Second * time.Duration(e.ttl))

	for {
		select {
		case <-ticker.C:
			if e.keepAliveChan == nil {
				if err := e.register(); err != nil {
					e.log.Error("register error:", err)
				}
			}
		case rsp := <-e.keepAliveChan:
			if rsp == nil {
				if err := e.register(); err != nil {
					e.log.Error("register error:", err)
				}
			}
		case <-e.closeChan:
			if err := e.unregister(); err != nil {
				e.log.Error("unregister error:", err)
			}
			e.log.Info("keepalive exit")
			return
		}
	}
}

func (e *etcdRegisterImpl) unregister() error {
	_, err := e.cli.Delete(context.Background(), e.node.BuildPath())
	defer func(cli *clientv3.Client) {
		err := cli.Close()
		if err != nil {
			e.log.Error("etcd close error:", err)
		}
	}(e.cli)
	return err
}

func newEtcdRegisterImpl(registerAddrs []string, dialTimeOUT time.Duration, logger *logrus.Logger) *etcdRegisterImpl {
	return &etcdRegisterImpl{
		etcdAddrs:   registerAddrs,
		dialTimeOut: dialTimeOUT,
		log:         logger,
	}
}
