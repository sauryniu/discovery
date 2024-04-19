/**
 * TODO:文件描述
 *
 * @title builder
 * @projectName grpcDemo
 * @author niuk
 * @date 2024/4/17 16:17
 */

package discovery

import "google.golang.org/grpc/resolver"

const (
	etcdScheme = "etcd"
)

type builder struct{}

func (builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	mr := manuResolver{
		cc:     cc,
		target: target,
		r:      eResolver,
	}
	// 记录解析器
	mr.r.setManuResolver(target.URL.Host, mr)
	// 记录需要解析的节点
	mr.r.setTargetNode(target.URL.Host)
	return mr, nil
}

func (builder) Scheme() string {
	return etcdScheme
}

func init() {
	resolver.Register(builder{})
}
