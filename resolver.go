/**
 * TODO:文件描述
 *
 * @title resolver
 * @projectName etcd
 * @author niuk
 * @date 2022/8/25 19:04
 */

package discovery

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Resolver interface {
	Start(node []ServiceNode) error
	getServiceNodes(host string) []ServiceNode
}

func NewResolver(registerAddrs []string, dialTimeOUT time.Duration, logger *logrus.Logger) Resolver {
	return newEtcdResolver(registerAddrs, dialTimeOUT, logger)
}
