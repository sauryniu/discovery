/**
 * TODO:文件描述
 *
 * @title discovery
 * @projectName grpcDemo
 * @author niuk
 * @date 2024/4/18 10:27
 */

package discovery

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

var eResolver *etcdResolver
var eRegister *etcdRegister

func init() {
	loggerInit()
	etcdRegisterInit()
	etcdResolverInit()
}

func SetDiscoveryAddress(address []string) {
	if len(address) == 0 {
		return
	}
	eResolver.etcdAddrs = address
	eRegister.etcdAddrs = address
}

func AddServiceNode(node *Node) {
	eRegister.addServiceNode(node)
}

func SetLogger(l *logrus.Logger) {
	logger = l
}
