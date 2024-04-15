/**
 * grpc etcd解析器
 *
 * @title etcd_resolver
 * @projectName discovery
 * @author niuk
 * @date 2022/8/25 19:05
 */

package etcd

import (
	"context"
	"encoding/json"
	"github.com/sauryniu/discovery"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"sync"
	"time"
)

const (
	scheme = "etcd"
)

type ResolverBuilder struct {
	cli         *clientv3.Client
	etcdAddrs   []string
	dialTimeOut time.Duration

	mu        sync.RWMutex
	log       *logrus.Logger
	nodes     []discovery.ServiceNode
	closeChan chan struct{}

	serviceNodes map[string][]discovery.ServiceNode
	mru          sync.RWMutex
	mr           map[string]*discovery.ManuResolver
}

func (e *ResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	m := &discovery.ManuResolver{
		Target: target,
		Cc:     cc,
		R:      e,
	}
	e.setManuResolver(target.URL.Host, m)
	m.ResolveNow(resolver.ResolveNowOptions{})
	return m, nil
}

func (e *ResolverBuilder) Scheme() string {
	return scheme
}

func (e *ResolverBuilder) Start(node []discovery.ServiceNode) error {
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

func (e *ResolverBuilder) SetDialTimeout(timeout time.Duration) {
	e.dialTimeOut = timeout
}

func (e *ResolverBuilder) SetLogger(logger *logrus.Logger) {
	e.log = logger
}

func (e *ResolverBuilder) GetServiceNodes(name string) []discovery.ServiceNode {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.serviceNodes[name]
}

func (e *ResolverBuilder) updateServerNode(name string, nodes []discovery.ServiceNode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.serviceNodes[name] = nodes
}

func (e *ResolverBuilder) removeServerNode(name, addr string) {
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

func (e *ResolverBuilder) addAddr(name string, node discovery.ServiceNode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.log.Info("add addr:", node.Addr)
	e.serviceNodes[name] = append(e.serviceNodes[name], node)
}

func (e *ResolverBuilder) syncAll() error {
	for _, node := range e.nodes {
		ctx, cancel := context.WithTimeout(context.Background(), e.dialTimeOut)
		rsp, err := e.cli.Get(ctx, node.BuildPrefix(), clientv3.WithPrefix())
		if err != nil {
			cancel()
			return err
		}
		serviceNodes := make([]discovery.ServiceNode, 0, len(rsp.Kvs))
		for _, element := range rsp.Kvs {
			serviceNode := discovery.ServiceNode{}
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

func (e *ResolverBuilder) setManuResolver(host string, m *discovery.ManuResolver) {
	e.mru.Lock()
	defer e.mru.Unlock()
	e.mr[host] = m
}

func (e *ResolverBuilder) getManuResolver(host string) (*discovery.ManuResolver, bool) {
	e.mru.RLock()
	defer e.mru.RUnlock()
	if m, ok := e.mr[host]; ok {
		return m, ok
	}
	return nil, false
}

func (e *ResolverBuilder) watch() {
	for _, node := range e.nodes {
		go func(n discovery.ServiceNode) {
			watchChan := e.cli.Watch(context.Background(), n.BuildPrefix(), clientv3.WithPrefix())
			ticker := time.NewTicker(time.Minute)
			for {
				select {
				case watchRsp := <-watchChan:
					for _, event := range watchRsp.Events {
						switch event.Type {
						case clientv3.EventTypePut:
							serviceNode := discovery.ServiceNode{}
							_ = json.Unmarshal(event.Kv.Value, &serviceNode)
							e.addAddr(n.Name, serviceNode)
							if r, ok := e.getManuResolver(n.Name); ok {
								r.ResolveNow(resolver.ResolveNowOptions{})
							}
						case clientv3.EventTypeDelete:
							addr, err := discovery.ParseKeyToAddr(string(event.Kv.Key))
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

func NewResolver(registerAddrs []string, option ...func(resolver discovery.Resolver)) *ResolverBuilder {
	e := &ResolverBuilder{
		etcdAddrs:    registerAddrs,
		dialTimeOut:  defaultDialTimeout,
		log:          logrus.New(),
		closeChan:    make(chan struct{}),
		serviceNodes: make(map[string][]discovery.ServiceNode),
		mr:           make(map[string]*discovery.ManuResolver),
	}
	for i := range option {
		option[i](e)
	}

	return e
}
