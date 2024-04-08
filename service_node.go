/**
 * grpc服务节点信息
 *
 * @title service_node
 * @projectName discovery
 * @author niuk
 * @date 2022/8/25 16:40
 */

package discovery

import (
	"errors"
	"fmt"
	"strings"
)

type ServiceNode struct {
	Name    string `json:"name"`    // 名称
	Addr    string `json:"addr"`    // 地址
	Version string `json:"version"` // 版本
	Weight  int64  `json:"weight"`  // 权重
}

func (s ServiceNode) transName() string {
	return strings.ReplaceAll(s.Name, ".", "/")
}

func (s ServiceNode) BuildPath() string {
	if len(s.Version) <= 0 {
		return fmt.Sprintf("/%s/%s", s.transName(), s.Addr)
	}
	return fmt.Sprintf("/%s/%s/%s", s.transName(), s.Version, s.Addr)
}

func (s ServiceNode) BuildPrefix() string {
	if len(s.Version) <= 0 {
		return fmt.Sprintf("/%s", s.transName())
	}
	return fmt.Sprintf("/%s/%s", s.transName(), s.Version)
}

func ParseKeyToAddr(key string) (string, error) {
	index := strings.LastIndex(key, "/")
	if index < 0 || !strings.Contains(key[index:], ":") {
		return "", errors.New("key format error")
	}
	return key[index+1:], nil
}
