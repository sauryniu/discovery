/**
 * TODO:文件描述
 *
 * @title node
 * @projectName grpcDemo
 * @author niuk
 * @date 2024/4/17 16:18
 */

package discovery

import (
	"fmt"
	"strings"
)

type Node struct {
	Name string `json:"name"` // 名称
	Addr string `json:"addr"` // 地址
}

// 把服务名中的 . 转换为 /
func (s Node) transName() string {
	return strings.ReplaceAll(s.Name, ".", "/")
}

// 构建节点 key
func (s Node) buildKey() string {
	return fmt.Sprintf("/%s/%s", s.transName(), s.Addr)
}

// 构建节点前缀
func (s Node) buildPrefix() string {
	return fmt.Sprintf("/%s", s.transName())
}
