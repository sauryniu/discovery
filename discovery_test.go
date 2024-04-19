/**
 * TODO:文件描述
 *
 * @title discovery_test.go
 * @projectName discovery
 * @author vivi
 * @date 2024/4/19 22:39
 */

package discovery

import (
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"net/url"
	"testing"
)

type testCc struct {
	addr string
}

func (t *testCc) UpdateState(state resolver.State) error {
	t.addr = state.Addresses[0].Addr
	return nil
}

func (*testCc) ReportError(err error) {
	//TODO implement me
	panic("implement me")
}

func (*testCc) NewAddress(addresses []resolver.Address) {
	//TODO implement me
	panic("implement me")
}

func (*testCc) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	//TODO implement me
	panic("implement me")
}

func Test(t *testing.T) {
	SetDiscoveryAddress([]string{"192.168.31.124:2379"})
	node := &Node{
		Name: "api.test.com",
		Addr: "127.0.0.1",
	}
	eRegister.addServiceNode(node)

	cc := &testCc{}
	r, _ := builder{}.Build(resolver.Target{URL: url.URL{Host: node.Name}}, cc, resolver.BuildOptions{})

	t.Run("test1", func(t *testing.T) {
		r.ResolveNow(resolver.ResolveNowOptions{})
		if cc.addr != node.Addr {
			t.Errorf("register %s != resolver %s", node.Addr, cc.addr)
		}
	})
}
