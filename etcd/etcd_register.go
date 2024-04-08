/**
 * grpc etcd注册器
 *
 * @title etcd_register
 * @projectName discovery
 * @author niuk
 * @date 2022/8/25 16:35
 */

package etcd

import (
	"context"
	"encoding/json"
	"github.com/sauryniu/discovery"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

const defaultTTL = 30
const defaultDialTimeout = 30

type RegisterImpl struct {
	etcdAddrs   []string
	dialTimeOut time.Duration
	leasesId    clientv3.LeaseID

	log           *logrus.Logger
	ttl           int64
	cli           *clientv3.Client
	node          discovery.ServiceNode
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	closeChan     chan struct{}
}

func (e *RegisterImpl) Register(node discovery.ServiceNode, option ...func(r discovery.Register)) error {
	var err error
	if e.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   e.etcdAddrs,
		DialTimeout: e.dialTimeOut,
	}); err != nil {
		return err
	}
	e.node = node
	e.closeChan = make(chan struct{})

	for i := range option {
		option[i](e)
	}

	if err = e.register(); err != nil {
		return err
	}

	go e.keepAlive()
	return nil
}

func (e *RegisterImpl) Unregister() {
	e.closeChan <- struct{}{}
}

func (e *RegisterImpl) register() error {
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

func (e *RegisterImpl) keepAlive() {
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

func (e *RegisterImpl) unregister() error {
	_, err := e.cli.Delete(context.Background(), e.node.BuildPath())
	defer func(cli *clientv3.Client) {
		err := cli.Close()
		if err != nil {
			e.log.Error("etcd close error:", err)
		}
	}(e.cli)
	return err
}

func (e *RegisterImpl) SetTTL(ttl time.Duration) {
	e.ttl = int64(ttl / time.Second)
}

func (e *RegisterImpl) SetDialTimeout(timeout time.Duration) {
	e.dialTimeOut = timeout
}

func (e *RegisterImpl) SetLogger(logger *logrus.Logger) {
	e.log = logger
}

func NewRegisterImpl(registerAddrs []string, option ...func(register discovery.Register)) *RegisterImpl {
	e := &RegisterImpl{
		etcdAddrs:   registerAddrs,
		dialTimeOut: defaultDialTimeout * time.Second,
		ttl:         int64(defaultTTL),
		log:         logrus.New(),
	}
	for i := range option {
		option[i](e)
	}
	return e
}
