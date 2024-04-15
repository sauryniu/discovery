/**
 * TODO:文件描述
 *
 * @title register_test.go
 * @projectName discovery
 * @author niuk
 * @date 2024/4/15 23:10
 */

package discovery

import (
	"testing"
)

func Test(t *testing.T) {
	host := "10.1.30.79:12379"
	name := "discovery.test"
	version := "V1.0"
	ip := "127.0.0.1"

	t.Run("test", func(t *testing.T) {
		reg := NewRegister([]string{host})
		err := reg.Register(ServiceNode{
			Name:    name,
			Addr:    ip,
			Version: version,
			Weight:  0,
		})
		if err != nil {
			panic(err)
		}
		defer reg.Unregister()

		res := NewResolver([]string{host})
		err = res.Start([]ServiceNode{{Name: name, Version: version}})
		if err != nil {
			t.Error(err)
		}

		nodes := res.GetServiceNodes(name)
		if nodes == nil {
			t.Error("nodes nil")
		}
		if nodes[0].Addr != ip {
			t.Error("add resolve error")
		}
	})
}