/**
 * TODO:文件描述
 *
 * @title resolver
 * @projectName grpcDemo
 * @author niuk
 * @date 2024/4/17 16:15
 */

package discovery

import "google.golang.org/grpc/resolver"

type IResolver interface {
	getServiceNodes(host string) []*Node
	setTargetNode(host string)
	setManuResolver(host string, m resolver.Resolver)
}

type manuResolver struct {
	cc     resolver.ClientConn
	target resolver.Target
	r      IResolver
}

func (m manuResolver) ResolveNow(options resolver.ResolveNowOptions) {
	nodes := m.r.getServiceNodes(m.target.URL.Host)
	addresses := make([]resolver.Address, 0)
	for i := range nodes {
		addresses = append(addresses, resolver.Address{
			Addr: nodes[i].Addr,
		})
	}
	if err := m.cc.UpdateState(resolver.State{
		Addresses: addresses,
	}); err != nil {
		logger.Errorf("resolver update cc state error:%s", err.Error())
	}
}

func (manuResolver) Close() {

}
