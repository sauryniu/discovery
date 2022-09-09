/**
 * TODO:文件描述
 *
 * @title Register
 * @projectName etcd
 * @author niuk
 * @date 2022/8/25 16:26
 */

package discovery

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Register interface {
	Register(node ServiceNode, ttl time.Duration) error
	Unregister()
}

func NewRegister(registerAddrs []string, dialTimeOUT time.Duration, logger *logrus.Logger) Register {
	return newEtcdRegisterImpl(registerAddrs, dialTimeOUT, logger)
}
