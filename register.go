/**
 * 注册器接口
 *
 * @title Register
 * @projectName discovery
 * @author niuk
 * @date 2022/8/25 16:26
 */

package discovery

import "github.com/sauryniu/discovery/etcd"

type Register interface {
	Register(node ServiceNode, option ...func(r Register)) error
	Unregister()
}

func NewRegister(registerAddrs []string, option ...func(register Register)) Register {
	return etcd.NewRegisterImpl(registerAddrs, option...)
}
