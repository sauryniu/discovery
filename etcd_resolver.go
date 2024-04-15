/**
 * grpc etcd解析器
 *
 * @title etcd_resolver
 * @projectName discovery
 * @author niuk
 * @date 2022/8/25 19:05
 */

package discovery

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"sync"
	"time"
)

const (
	scheme = "etcd"
)

type etcdResolverBuilder struct {
	cli         *clientv3.Client
	etcdAddrs   []string
	dialTimeOut time.Duration

	mu        sync.RWMutex
	log       *logrus.Logger
	nodes     []ServiceNode
	closeChan chan struct{}

	serviceNodes map[string][]ServiceNode
	mru          sync.RWMutex
	mr           map[string]*ManuResolver
}

func (e *etcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	m := &ManuResolver{
		Target: target,
		Cc:     cc,
		R:      e,
	}
	e.setManuResolver(target.URL.Host, m)
	m.ResolveNow(resolver.ResolveNowOptions{})
	return m, nil
}

func (e *etcdResolverBuilder) Scheme() string {
	return scheme
}

func (e *etcdResolverBuilder) Start(node []ServiceNode) error {
	var err error
	e.nodes = node
	e.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   e.etcdAddrs,
		DialTimeout: e.dialTimeOut,
	})
	if err != nil {
		return err
	}
	err = e.syncAll()
	if err != nil {
		return err
	}

	resolver.Register(e)

	go e.watch()
	return nil
}

func (e *etcdResolverBuilder) SetDialTimeout(timeout time.Duration) {
	e.dialTimeOut = timeout
}

func (e *etcdResolverBuilder) SetLogger(logger *logrus.Logger) {
	e.log = logger
}

func (e *etcdResolverBuilder) GetServiceNodes(name string) []ServiceNode {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.serviceNodes[name]
}

func (e *etcdResolverBuilder) updateServerNode(name string, nodes []ServiceNode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.serviceNodes[name] = nodes
}

func (e *etcdResolverBuilder) removeServerNode(name, addr string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	nodes := e.serviceNodes[name]
	for i, node := range nodes {
		if node.Addr == addr {
			e.log.Info("remove addr:", addr)
			nodes[i] = nodes[len(nodes)-1]
		}
	}
	e.serviceNodes[name] = nodes[:len(nodes)-1]
}

func (e *etcdResolverBuilder) addAddr(name string, node ServiceNode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.log.Info("add addr:", node.Addr)
	e.serviceNodes[name] = append(e.serviceNodes[name], node)
}

func (e *etcdResolverBuilder) syncAll() error {
	for _, node := range e.nodes {
		ctx, cancel := context.WithTimeout(context.Background(), e.dialTimeOut)
		rsp, err := e.cli.Get(ctx, node.BuildPrefix(), clientv3.WithPrefix())
		if err != nil {
			cancel()
			return err
		}
		serviceNodes := make([]ServiceNode, 0, len(rsp.Kvs))
		for _, element := range rsp.Kvs {
			serviceNode := ServiceNode{}
			_ = json.Unmarshal(element.Value, &serviceNode)
			serviceNodes = append(serviceNodes, serviceNode)
		}
		if len(serviceNodes) > 0 {
			e.updateServerNode(node.Name, serviceNodes)
		}
		cancel()
	}

	return nil
}

func (e *etcdResolverBuilder) setManuResolver(host string, m *ManuResolver) {
	e.mru.Lock()
	defer e.mru.Unlock()
	e.mr[host] = m
}

func (e *etcdResolverBuilder) getManuResolver(host string) (*ManuResolver, bool) {
	e.mru.RLock()
	defer e.mru.RUnlock()
	if m, ok := e.mr[host]; ok {
		return m, ok
	}
	return nil, false
}

func (e *etcdResolverBuilder) watch() {
	for _, node := range e.nodes {
		go func(n ServiceNode) {
			watchChan := e.cli.Watch(context.Background(), n.BuildPrefix(), clientv3.WithPrefix())
			ticker := time.NewTicker(time.Minute)
			for {
				select {
				case watchRsp := <-watchChan:
					for _, event := range watchRsp.Events {
						switch event.Type {
						case clientv3.EventTypePut:
							serviceNode := ServiceNode{}
							_ = json.Unmarshal(event.Kv.Value, &serviceNode)
							e.addAddr(n.Name, serviceNode)
							if r, ok := e.getManuResolver(n.Name); ok {
								r.ResolveNow(resolver.ResolveNowOptions{})
							}
						case clientv3.EventTypeDelete:
							addr, err := ParseKeyToAddr(string(event.Kv.Key))
							if err != nil {
								e.log.Error("etcd delete event error:", err)
								continue
							}
							e.removeServerNode(n.Name, addr)
							if r, ok := e.getManuResolver(n.Name); ok {
								r.ResolveNow(resolver.ResolveNowOptions{})
							}
						}
					}
				case <-ticker.C:
					_ = e.syncAll()
					if r, ok := e.getManuResolver(n.Name); ok {
						r.ResolveNow(resolver.ResolveNowOptions{})
					}
				case <-e.closeChan:
					e.log.Info(n.Name, " resolver exit")
					return
				}
			}
		}(node)
	}
}

func newEtcdResolverImpl(registerAddrs []string, option ...func(resolver Resolver)) *etcdResolverBuilder {
	e := &etcdResolverBuilder{
		etcdAddrs:    registerAddrs,
		dialTimeOut:  defaultDialTimeout,
		log:          logrus.New(),
		closeChan:    make(chan struct{}),
		serviceNodes: make(map[string][]ServiceNode),
		mr:           make(map[string]*ManuResolver),
	}
	for i := range option {
		option[i](e)
	}

	return e
}
