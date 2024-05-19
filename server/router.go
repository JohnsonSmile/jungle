package server

import (
	"log"
	"net/http"
	"strings"
)

type router struct {
	tree            map[string]*node
	supportedMethod map[string]bool
}

func newRouter() *router {
	return &router{
		tree: make(map[string]*node),
		supportedMethod: map[string]bool{
			http.MethodGet:     true,
			http.MethodHead:    true,
			http.MethodPost:    true,
			http.MethodPut:     true,
			http.MethodPatch:   true,
			http.MethodDelete:  true,
			http.MethodConnect: true,
			http.MethodOptions: true,
			http.MethodTrace:   true,
		},
	}
}

func (r *router) AddRoute(method string, path string, handles ...HandleFunc) {
	// validate the method must be one of the http.Method
	if !r.supportedMethod[method] {
		log.Fatalf("method: %s not supported", method)
	}
	root, ok := r.tree[method]
	if !ok {
		root = &node{
			path:     "/",
			children: make([]*node, 0),
		}
		r.tree[method] = root
	}
	// 切割path
	segs := strings.Split(path, "/")
	cur := root
	total := len(segs)
	for idx, seg := range segs {
		// 去除有问题的seg
		if seg == "" {
			continue
		}
		// 如果是 *
		if seg == "*" {
			next := &node{
				path:     seg,
				children: make([]*node, 0),
			}
			if cur.starChild == nil {
				cur.starChild = next
			}
			if total == idx+1 {
				cur.starChild.handleChains = handles
			}
			cur = cur.starChild
			continue
		}

		// 应该和当前节点的children来比较
		insertIndex := -1
		continueIndex := -1
		children := cur.children
		for idx, child := range children {
			// 大于需要插入
			if child.path > seg {
				insertIndex = idx
				break
			} else if child.path == seg {
				// 相等, 需要 找他的children,将cur设置成他就可以了
				continueIndex = idx
				break
			}
		}
		if continueIndex >= 0 {
			// 下一个
			cur = cur.children[continueIndex]
			// 这个地方可能会忽略掉,导致出问题.
			if total == idx+1 {
				cur.handleChains = handles
			}
			continue
		}

		// 需要insert
		if insertIndex >= 0 {

			// 创建
			next := &node{
				path:     seg,
				children: make([]*node, 0),
			}

			// 是叶子节点,绑定调用链
			if total == idx+1 {
				next.handleChains = handles
			}
			if insertIndex != 0 {
				insertIndex = insertIndex - 1
			}
			// 排序
			before := cur.children[:insertIndex]
			after := cur.children[insertIndex:]
			children := make([]*node, 0, len(cur.children)+1)
			children = append(children, before...)
			children = append(children, next)
			children = append(children, after...)
			cur.children = children
			cur = next
			continue
		}

		// 需要添加到队尾
		// 创建
		next := &node{
			path:     seg,
			children: make([]*node, 0),
		}

		// 是叶子节点,绑定调用链
		if total == idx+1 {
			next.handleChains = handles
		}
		cur.children = append(cur.children, next)
		cur = next
	}
}

func (r *router) FindRoute(method string, path string) (n *node, found bool) {

	node, ok := r.tree[method]
	if !ok {
		return nil, false
	}

	// 根路径
	if node.path == path {
		if len(node.handleChains) > 0 {
			return node, true
		}
		return nil, false
	}

	segs := strings.Split(path, "/")
	total := len(segs)
	for idx, seg := range segs {
		if seg == "" {
			continue
		}

		// TODO: 二分查找...因为是排序了的...
		// 查找所有的children
		shouldMatchStar := true
		for _, child := range node.children {
			if child.path == seg {
				if len(child.handleChains) > 0 &&
					idx+1 == total {
					return child, true
				}
				shouldMatchStar = false
				node = child
			}
		}
		if node.starChild != nil && shouldMatchStar {
			if len(node.starChild.handleChains) > 0 &&
				idx+1 == total {
				return node.starChild, true
			}
			node = node.starChild
			continue
		}
	}
	return nil, false
}

// 定义api接口.
// 定义测试.
// 添加测试用例.
// 实现,并确保通过测试用例.
// 重复添加测试用力,并尽量满足所有的场景.
// 实现,并确保通过所有的测试用例.

// 通配符 /user/* 和 /user/id 的 优先级
// 一般设计成 /user/id > /user/* /user/* 作为兜底
type node struct {
	path     string
	children []*node
	// 通配符
	starChild *node
	// 责任链
	handleChains []HandleFunc
}
