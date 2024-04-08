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

type ManuResolver struct {
	Cc     resolver.ClientConn
	Target resolver.Target
	R      Resolver
}

func (m *ManuResolver) ResolveNow(options resolver.ResolveNowOptions) {
	var addresses []resolver.Address
	serviceNodes := m.R.GetServiceNodes(m.Target.URL.Host)
	for _, node := range serviceNodes {
		addresses = append(addresses, resolver.Address{
			Addr:       node.Addr,
			ServerName: m.Target.URL.Host,
		})
	}
	_ = m.Cc.UpdateState(resolver.State{
		Addresses: addresses,
	})
}

func (m *ManuResolver) Close() {

}
