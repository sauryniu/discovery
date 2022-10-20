/**
 * grpc解析器
 *
 * @title manu_resolver
 * @projectName discovery
 * @author niuk
 * @date 2022/9/7 11:02
 */

package discovery

import (
	"google.golang.org/grpc/resolver"
)

type manuResolver struct {
	cc     resolver.ClientConn
	target resolver.Target
	r      Resolver
}

func (m *manuResolver) ResolveNow(options resolver.ResolveNowOptions) {
	var addresses []resolver.Address
	serviceNodes := m.r.getServiceNodes(m.target.URL.Host)
	for _, node := range serviceNodes {
		addresses = append(addresses, resolver.Address{
			Addr:       node.Addr,
			ServerName: m.target.URL.Host,
		})
	}
	_ = m.cc.UpdateState(resolver.State{
		Addresses: addresses,
	})
}

func (m *manuResolver) Close() {

}
