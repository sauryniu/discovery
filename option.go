/**
 * TODO:文件描述
 *
 * @title config
 * @projectName discovery
 * @author niuk
 * @date 2024/4/7 17:05
 */

package discovery

import (
	"github.com/sirupsen/logrus"
	"time"
)

func RegisterWithTTL(ttl time.Duration) func(r Register) {
	return func(r Register) {
		switch r.(type) {
		case *etcdRegisterImpl:
			r.(*etcdRegisterImpl).setTTL(ttl)
		}
	}
}

func RegisterWithDialTimeout(timeout time.Duration) func(r Register) {
	return func(r Register) {
		switch r.(type) {
		case *etcdRegisterImpl:
			r.(*etcdRegisterImpl).setDialTimeout(timeout)
		}
	}
}

func RegisterWithLogger(logger *logrus.Logger) func(r Register) {
	return func(r Register) {
		switch r.(type) {
		case *etcdRegisterImpl:
			r.(*etcdRegisterImpl).setLogger(logger)
		}
	}
}

func ResolverWithDialTimeout(timeout time.Duration) func(r Resolver) {
	return func(r Resolver) {
		switch r.(type) {
		case *etcdResolverBuilder:
			r.(*etcdResolverBuilder).setDialTimeout(timeout)
		}
	}
}

func ResolverWithLogger(logger *logrus.Logger) func(r Resolver) {
	return func(r Resolver) {
		switch r.(type) {
		case *etcdResolverBuilder:
			r.(*etcdResolverBuilder).setLogger(logger)
		}
	}
}
