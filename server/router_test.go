package server

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// go test -v server/*.go -run TestRouter_AddRoute
func TestRouter_AddRoute(t *testing.T) {

	var mockHandler HandleFunc = func(ctx *Context) {}
	var mockHandler1 HandleFunc = func(ctx *Context) {}
	testCases := []struct {
		name   string
		routes []struct {
			method      string
			path        string
			handler     HandleFunc
			middlewares []HandleFunc
		}
		wantTree    map[string]*node
		shouldPanic bool
		panicText   string
	}{
		{
			name: "only get method",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodGet,
					path:        "/user/home",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/home",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
			},
			wantTree: map[string]*node{
				http.MethodGet: {
					path: "/",
					children: []*node{
						{
							path:     "home",
							children: make([]*node, 0),
							handlerChains: []HandleFunc{
								mockHandler,
							},
							matchedPath: "/home",
						},
						{
							path: "user",
							children: []*node{
								{
									path:     "home",
									children: make([]*node, 0),
									handlerChains: []HandleFunc{
										mockHandler,
									},
									matchedPath: "/user/home",
								},
							},
						},
					},
				},
			},
		}, {
			name: "get, post, put, delete method with same parent node",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodGet,
					path:        "/users",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/users/id",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodPost,
					path:        "/users/id",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodPut,
					path:        "/users/id",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodDelete,
					path:        "/users/id",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
			},
			wantTree: map[string]*node{
				http.MethodGet: {
					path: "/",
					children: []*node{
						{
							path: "users",
							children: []*node{
								{
									path:     "id",
									children: make([]*node, 0),
									handlerChains: []HandleFunc{
										mockHandler,
									},
									matchedPath: "/users/id",
								},
							},
							handlerChains: []HandleFunc{
								mockHandler,
							},
							matchedPath: "/users",
						},
					},
				},
				http.MethodPost: {
					path: "/",
					children: []*node{
						{
							path: "users",
							children: []*node{
								{
									path:     "id",
									children: make([]*node, 0),
									handlerChains: []HandleFunc{
										mockHandler,
									},
									matchedPath: "/users/id",
								},
							},
						},
					},
				},
				http.MethodPut: {
					path: "/",
					children: []*node{
						{
							path: "users",
							children: []*node{
								{
									path:     "id",
									children: make([]*node, 0),
									handlerChains: []HandleFunc{
										mockHandler,
									},
									matchedPath: "/users/id",
								},
							},
						},
					},
				},
				http.MethodDelete: {
					path: "/",
					children: []*node{
						{
							path: "users",
							children: []*node{
								{
									path:     "id",
									children: make([]*node, 0),
									handlerChains: []HandleFunc{
										mockHandler,
									},
									matchedPath: "/users/id",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "only get method with third level same parent node",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodGet,
					path:        "/user/home/id",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/user/home",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/user",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
			},
			wantTree: map[string]*node{
				http.MethodGet: {
					path: "/",
					children: []*node{
						{
							path: "user",
							children: []*node{
								{
									path: "home",
									children: []*node{
										{
											path:     "id",
											children: make([]*node, 0),
											handlerChains: []HandleFunc{
												mockHandler1,
											},
											matchedPath: "/user/home/id",
										},
									},
									handlerChains: []HandleFunc{
										mockHandler,
									},
									matchedPath: "/user/home",
								},
							},
							handlerChains: []HandleFunc{
								mockHandler1,
							},
							matchedPath: "/user",
						},
					},
				},
			},
		},
		{
			name: "star method should be created properly",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodGet,
					path:        "/user/home",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/user/*",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/user/*/id",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
			},
			wantTree: map[string]*node{
				http.MethodGet: {
					path: "/",
					children: []*node{
						{
							path: "user",
							children: []*node{
								{
									path:     "home",
									children: make([]*node, 0),
									handlerChains: []HandleFunc{
										mockHandler,
									},
									matchedPath: "/user/home",
								},
							},
							starChild: &node{
								path: "*",
								children: []*node{
									{
										path:     "id",
										children: make([]*node, 0),
										handlerChains: []HandleFunc{
											mockHandler,
										},
										matchedPath: "/user/*/id",
									},
								},
								handlerChains: []HandleFunc{
									mockHandler,
								},
								matchedPath: "/user/*",
							},
						},
					},
				},
			},
		},
		{
			name: "only get method with path param",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodGet,
					path:        "/users",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/users/:id",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/users/:id/article",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
			},
			wantTree: map[string]*node{
				http.MethodGet: {
					path: "/",
					children: []*node{
						{
							path:     "users",
							children: make([]*node, 0),
							paramChild: &node{
								path: "id",
								children: []*node{
									{
										path:     "article",
										children: make([]*node, 0),
										handlerChains: []HandleFunc{
											mockHandler,
										},
										matchedPath: "/users/:id/article",
									},
								},
								handlerChains: []HandleFunc{
									mockHandler,
								},
								matchedPath: "/users/:id",
							},
							handlerChains: []HandleFunc{
								mockHandler,
							},
							matchedPath: "/users",
						},
					},
				},
			},
		},
		{
			name: "should panic when create param node but already has a star node",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodGet,
					path:        "/users",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/users/*",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/users/:id",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
			},
			shouldPanic: true,
			panicText:   "already a star child exists",
		},
		{
			name: "should panic when create star node but already has a param node",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodGet,
					path:        "/users",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/users/:id",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/users/*",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
			},
			shouldPanic: true,
			panicText:   "already a param child exists",
		},
	}

	// mock 方案
	for idx, tc := range testCases {
		r := newRouter()
		t.Log(tc.name)
		if tc.shouldPanic {
			require.Panicsf(t, func() {
				for _, route := range tc.routes {
					r.AddRoute(route.method, route.path, route.middlewares, route.handler)
				}
			}, tc.panicText)
			continue
		}
		for _, route := range tc.routes {
			r.AddRoute(route.method, route.path, route.middlewares, route.handler)
		}
		if idx == 0 {
			continue
		}
		// 断言
		require.NoError(t, treeEqual(tc.wantTree, r.tree))
	}
}

// go test -v server/*.go -run TestRouter_FindRoute
func TestRouter_FindRoute(t *testing.T) {

	var mockHandler HandleFunc = func(ctx *Context) {}
	var mockHandler1 HandleFunc = func(ctx *Context) {}
	testCases := []struct {
		name   string
		method string
		path   string
		routes []struct {
			method      string
			path        string
			handler     HandleFunc
			middlewares []HandleFunc
		}
		wantNode    *node
		shouldFound bool
	}{
		{
			name:   "get node that exists with first level",
			method: http.MethodHead,
			path:   "/user",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:  http.MethodHead,
					path:    "/user",
					handler: mockHandler,
				},
				{
					method:  http.MethodHead,
					path:    "/user/id",
					handler: mockHandler1,
				},
			},
			wantNode: &node{
				path: "user",
				children: []*node{
					{
						path:     "id",
						children: make([]*node, 0),
						handlerChains: []HandleFunc{
							mockHandler1,
						},
						matchedPath: "/user/id",
					},
				},
				handlerChains: []HandleFunc{
					mockHandler,
				},
				matchedPath: "/user",
			},
			shouldFound: true,
		},
		{
			name:   "get node that exists with second level",
			method: http.MethodHead,
			path:   "/user/id",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodHead,
					path:        "/user",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/id",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
			},
			wantNode: &node{
				path:     "id",
				children: make([]*node, 0),
				handlerChains: []HandleFunc{
					mockHandler1,
				},
				matchedPath: "/user/id",
			},
			shouldFound: true,
		},
		{
			name:   "get node that exists with third level and more node exists",
			method: http.MethodHead,
			path:   "/user/home/id",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodHead,
					path:        "/user",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/home",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/home/id",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodPost,
					path:        "/admin/home",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodGet,
					path:        "/admin/home",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
			},
			wantNode: &node{
				path:     "id",
				children: make([]*node, 0),
				handlerChains: []HandleFunc{
					mockHandler1,
				},
				matchedPath: "/user/home/id",
			},
			shouldFound: true,
		},
		{
			name:   "get node that not exists",
			method: http.MethodHead,
			path:   "/user/xxx",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodHead,
					path:        "/user",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/id",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
			},
			wantNode:    nil,
			shouldFound: false,
		},
		{
			name:   "find route that exists with *",
			method: http.MethodHead,
			path:   "/user/1",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodHead,
					path:        "/user",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/*",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
			},
			wantNode: &node{
				path:     "*",
				children: make([]*node, 0),
				handlerChains: []HandleFunc{
					mockHandler1,
				},
				matchedPath: "/user/*",
			},
			shouldFound: true,
		},
		{
			name:   "find route that exists with complex *",
			method: http.MethodHead,
			path:   "/user/2",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodHead,
					path:        "/user",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/*",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/*/id",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
			},
			wantNode: &node{
				path: "*",
				children: []*node{
					{
						path:     "id",
						children: make([]*node, 0),
						handlerChains: []HandleFunc{
							mockHandler,
						},
						matchedPath: "/user/*/id",
					},
				},
				handlerChains: []HandleFunc{
					mockHandler1,
				},
				matchedPath: "/user/*",
			},
			shouldFound: true,
		},
		{
			name:   "find route that exists with route param",
			method: http.MethodHead,
			path:   "/user/2",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:  http.MethodHead,
					path:    "/user",
					handler: mockHandler,
				},
				{
					method:  http.MethodHead,
					path:    "/user/:id",
					handler: mockHandler1,
				},
				{
					method:  http.MethodHead,
					path:    "/user/:id/article",
					handler: mockHandler,
				},
			},
			wantNode: &node{
				path: "id",
				children: []*node{
					{
						path:     "article",
						children: make([]*node, 0),
						handlerChains: []HandleFunc{
							mockHandler,
						},
						matchedPath: "/user/:id/article",
					},
				},
				handlerChains: []HandleFunc{
					mockHandler1,
				},
				matchedPath: "/user/:id",
			},
			shouldFound: true,
		},
		{
			name:   "find route that exists with route param 1",
			method: http.MethodHead,
			path:   "/user/2/article",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodHead,
					path:        "/user",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/:id",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/:id/article",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
			},
			wantNode: &node{
				path:     "article",
				children: make([]*node, 0),
				handlerChains: []HandleFunc{
					mockHandler,
				},
				matchedPath: "/user/:id/article",
			},
			shouldFound: true,
		},
		{
			name:   "find route that exists with route param 2",
			method: http.MethodHead,
			path:   "/user/2",
			routes: []struct {
				method      string
				path        string
				handler     HandleFunc
				middlewares []HandleFunc
			}{
				{
					method:      http.MethodHead,
					path:        "/user",
					handler:     mockHandler,
					middlewares: []HandleFunc{},
				},
				{
					method:      http.MethodHead,
					path:        "/user/:id",
					handler:     mockHandler1,
					middlewares: []HandleFunc{},
				},
			},
			wantNode: &node{
				path:     "id",
				children: make([]*node, 0),
				handlerChains: []HandleFunc{
					mockHandler1,
				},
				matchedPath: "/user/:id",
			},
			shouldFound: true,
		},
	}

	// mock 方案
	for _, tc := range testCases {
		r := newRouter()
		t.Log(tc.name)
		for _, route := range tc.routes {
			r.AddRoute(route.method, route.path, route.middlewares, route.handler)
		}
		matchInfo, found := r.FindRoute(tc.method, tc.path)
		// 断言
		t.Log("tc.wantNode", tc.wantNode)
		t.Log("matchInfo", matchInfo.node)
		require.NoError(t, nodeEqual(tc.wantNode, matchInfo.node))
		require.True(t, found == tc.shouldFound)
	}
}

func treeEqual(src map[string]*node, dst map[string]*node) error {

	if len(src) != len(dst) {
		return errors.New("length of src should be equal to dst")
	}
	for key, srcNode := range src {
		dstNode, ok := dst[key]
		if !ok {
			return fmt.Errorf("node with key: [%s] not exists in dst", key)
		}
		if err := nodeEqual(srcNode, dstNode); err != nil {
			return err
		}
	}
	return nil
}

func nodeEqual(src *node, dst *node) error {
	if src == nil && dst == nil {
		return nil
	}
	if src == nil || dst == nil {
		return errors.New("src or dst is nil")
	}
	if src.path != dst.path {
		return fmt.Errorf("src path: [%s] not equal to dst path: [%s]", src.path, dst.path)
	}
	if len(src.children) != len(dst.children) {
		return fmt.Errorf("src children length: [%+v] not equal to dst children length: [%+v]", src, dst)
	}

	if len(src.handlerChains) != len(dst.handlerChains) {
		return fmt.Errorf("src handleChains length: [%+v] not equal to dst handleChains length: [%+v]", src, dst)
	}
	if src.matchedPath != dst.matchedPath {
		return fmt.Errorf("src matchedPath: [%s] not equal to dst matchedPath: [%s]", src.matchedPath, dst.matchedPath)
	}

	if err := nodeEqual(src.paramChild, dst.paramChild); err != nil {
		return err
	}

	if err := nodeEqual(src.starChild, dst.starChild); err != nil {
		return err
	}

	for i := 0; i < len(src.children); i++ {
		err := nodeEqual(src.children[i], dst.children[i])
		if err != nil {
			return err
		}
	}

	for idx, handler := range src.handlerChains {
		if reflect.TypeOf(handler) != reflect.TypeOf(dst.handlerChains[idx]) {
			return fmt.Errorf("src handler: [%+v] not equal to dst handler: [%+v]", src, dst)
		}
	}
	return nil
}
