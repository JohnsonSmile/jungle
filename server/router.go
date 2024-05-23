package server

import (
	"log"
	"net/http"
	"net/url"
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

func (r *router) AddRoute(method string, path string, servMiddlewares []HandleFunc, handler HandleFunc, middlewares ...HandleFunc) {

	// middle 和 handlers 组合
	handlerChain := append(servMiddlewares, middlewares...)
	handlerChain = append(handlerChain, handler)
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
			// 不允许同时存在 * 和 :
			if cur.paramChild != nil {
				panic("already a param child exists")
			}
			next := &node{
				path:     seg,
				children: make([]*node, 0),
			}
			if cur.starChild == nil {
				cur.starChild = next
			}
			if total == idx+1 {
				cur.starChild.matchedPath = path
				cur.starChild.handlerChains = handlerChain
			}
			cur = cur.starChild
			continue
		} else if strings.HasPrefix(seg, ":") {

			// 不允许同时存在 * 和 :
			if cur.starChild != nil {
				panic("already a star child exists")
			}
			next := &node{
				path: strings.Trim(seg, ":"),
				// path:     seg,
				children: make([]*node, 0),
			}
			if cur.paramChild == nil {
				cur.paramChild = next
			}
			if total == idx+1 {
				cur.paramChild.matchedPath = path
				cur.paramChild.handlerChains = handlerChain
			}
			cur = cur.paramChild
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
				cur.matchedPath = path
				cur.handlerChains = handlerChain
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
				next.matchedPath = path
				next.handlerChains = handlerChain
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
			next.matchedPath = path
			next.handlerChains = handlerChain
		}
		cur.children = append(cur.children, next)
		cur = next
	}
}

// 这种路径参数只支持,同名,且唯一一个
func (r *router) FindRoute(method string, path string) (n *MatchNode, found bool) {

	node, ok := r.tree[method]
	if !ok {
		return nil, false
	}

	var pathParams = make(url.Values)
	// 根路径
	if node.path == path {
		if len(node.handlerChains) > 0 {
			return &MatchNode{
				node:       node,
				pathParams: pathParams,
			}, true
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
		shouldMatchOther := true
		for _, child := range node.children {
			if child.path == seg {
				if len(child.handlerChains) > 0 &&
					idx+1 == total {
					return &MatchNode{
						node:       child,
						pathParams: pathParams,
					}, true
				}
				shouldMatchOther = false
				node = child
			}
		}

		// 路径参数匹配优先级 > *匹配
		if node.paramChild != nil && shouldMatchOther {
			// 记录pathParams,传递给最终的叶子节点.
			if !pathParams.Has(node.paramChild.path) {
				pathParams[node.paramChild.path] = make([]string, 0)
			}
			pathParams.Add(node.paramChild.path, seg)
			if len(node.paramChild.handlerChains) > 0 &&
				idx+1 == total {
				return &MatchNode{
					node:       node.paramChild,
					pathParams: pathParams,
				}, true
			}
			node = node.paramChild
			continue
		}

		// *匹配
		if node.starChild != nil && shouldMatchOther {
			if len(node.starChild.handlerChains) > 0 &&
				idx+1 == total {
				return &MatchNode{
					node:       node.starChild,
					pathParams: pathParams,
				}, true
			}
			node = node.starChild
			continue
		}
	}
	return &MatchNode{
		node: nil,
	}, false
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
	path        string
	matchedPath string
	children    []*node
	// 通配符
	starChild *node
	// 路径参数
	paramChild *node
	// 责任链
	handlerChains []HandleFunc
}

type MatchNode struct {
	node       *node
	pathParams url.Values
}
