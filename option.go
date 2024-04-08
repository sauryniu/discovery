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
	"github.com/sauryniu/discovery/etcd"
	"github.com/sirupsen/logrus"
	"time"
)

func RegisterWithTTL(ttl time.Duration) func(r Register) {
	return func(r Register) {
		switch r.(type) {
		case *etcd.RegisterImpl:
			r.(*etcd.RegisterImpl).SetTTL(ttl)
		}
	}
}

func RegisterWithDialTimeout(timeout time.Duration) func(r Register) {
	return func(r Register) {
		switch r.(type) {
		case *etcd.RegisterImpl:
			r.(*etcd.RegisterImpl).SetDialTimeout(timeout)
		}
	}
}

func RegisterWithLogger(logger *logrus.Logger) func(r Register) {
	return func(r Register) {
		switch r.(type) {
		case *etcd.RegisterImpl:
			r.(*etcd.RegisterImpl).SetLogger(logger)
		}
	}
}

func ResolverWithDialTimeout(timeout time.Duration) func(r Resolver) {
	return func(r Resolver) {
		switch r.(type) {
		case *etcd.ResolverBuilder:
			r.(*etcd.ResolverBuilder).SetDialTimeout(timeout)
		}
	}
}

func ResolverWithLogger(logger *logrus.Logger) func(r Resolver) {
	return func(r Resolver) {
		switch r.(type) {
		case *etcd.ResolverBuilder:
			r.(*etcd.ResolverBuilder).SetLogger(logger)
		}
	}
}
