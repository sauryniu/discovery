/**
 * 解析器接口
 *
 * @title resolver
 * @projectName discovery
 * @author niuk
 * @date 2022/8/25 19:04
 */

package discovery

import "github.com/sauryniu/discovery/etcd"

type Resolver interface {
	Start(node []ServiceNode) error
	GetServiceNodes(host string) []ServiceNode
}

func NewResolver(registerAddrs []string, option ...func(resolver Resolver)) Resolver {
	return etcd.NewResolver(registerAddrs, option...)
}
