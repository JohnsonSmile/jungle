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
			method  string
			path    string
			handler HandleFunc
		}
		wantTree map[string]*node
	}{
		{
			name: "only get method",
			routes: []struct {
				method  string
				path    string
				handler HandleFunc
			}{
				{
					method:  http.MethodGet,
					path:    "/user/home",
					handler: mockHandler,
				},
				{
					method:  http.MethodGet,
					path:    "/home",
					handler: mockHandler,
				},
			},
			wantTree: map[string]*node{
				http.MethodGet: {
					path: "/",
					children: []*node{
						{
							path:     "home",
							children: make([]*node, 0),
							handleChains: []HandleFunc{
								mockHandler,
							},
						},
						{
							path: "user",
							children: []*node{
								{
									path:     "home",
									children: make([]*node, 0),
									handleChains: []HandleFunc{
										mockHandler,
									},
								},
							},
						},
					},
				},
			},
		}, {
			name: "get, post, put, delete method with same parent node",
			routes: []struct {
				method  string
				path    string
				handler HandleFunc
			}{
				{
					method:  http.MethodGet,
					path:    "/users",
					handler: mockHandler,
				},
				{
					method:  http.MethodGet,
					path:    "/users/id",
					handler: mockHandler,
				},
				{
					method:  http.MethodPost,
					path:    "/users/id",
					handler: mockHandler,
				},
				{
					method:  http.MethodPut,
					path:    "/users/id",
					handler: mockHandler,
				},
				{
					method:  http.MethodDelete,
					path:    "/users/id",
					handler: mockHandler,
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
									handleChains: []HandleFunc{
										mockHandler,
									},
								},
							},
							handleChains: []HandleFunc{
								mockHandler,
							},
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
									handleChains: []HandleFunc{
										mockHandler,
									},
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
									handleChains: []HandleFunc{
										mockHandler,
									},
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
									handleChains: []HandleFunc{
										mockHandler,
									},
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
				method  string
				path    string
				handler HandleFunc
			}{
				{
					method:  http.MethodGet,
					path:    "/user/home/id",
					handler: mockHandler1,
				},
				{
					method:  http.MethodGet,
					path:    "/user/home",
					handler: mockHandler,
				},
				{
					method:  http.MethodGet,
					path:    "/user",
					handler: mockHandler1,
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
											handleChains: []HandleFunc{
												mockHandler1,
											},
										},
									},
									handleChains: []HandleFunc{
										mockHandler,
									},
								},
							},
							handleChains: []HandleFunc{
								mockHandler1,
							},
						},
					},
				},
			},
		},
	}

	// mock 方案
	for idx, tc := range testCases {
		r := newRouter()
		t.Log(tc.name)
		for _, route := range tc.routes {
			r.AddRoute(route.method, route.path, route.handler)
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
			method  string
			path    string
			handler HandleFunc
		}
		wantNode    *node
		shouldFound bool
	}{
		{
			name:   "get node that exists with first level",
			method: http.MethodHead,
			path:   "/user",
			routes: []struct {
				method  string
				path    string
				handler HandleFunc
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
						handleChains: []HandleFunc{
							mockHandler1,
						},
					},
				},
				handleChains: []HandleFunc{
					mockHandler,
				},
			},
			shouldFound: true,
		},
		{
			name:   "get node that exists with second level",
			method: http.MethodHead,
			path:   "/user/id",
			routes: []struct {
				method  string
				path    string
				handler HandleFunc
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
				path:     "id",
				children: make([]*node, 0),
				handleChains: []HandleFunc{
					mockHandler1,
				},
			},
			shouldFound: true,
		},
		{
			name:   "get node that exists with third level and more node exists",
			method: http.MethodHead,
			path:   "/user/home/id",
			routes: []struct {
				method  string
				path    string
				handler HandleFunc
			}{
				{
					method:  http.MethodHead,
					path:    "/user",
					handler: mockHandler,
				},
				{
					method:  http.MethodHead,
					path:    "/user/home",
					handler: mockHandler1,
				},
				{
					method:  http.MethodHead,
					path:    "/user/home/id",
					handler: mockHandler1,
				},
				{
					method:  http.MethodPost,
					path:    "/admin/home",
					handler: mockHandler1,
				},
				{
					method:  http.MethodGet,
					path:    "/admin/home",
					handler: mockHandler1,
				},
			},
			wantNode: &node{
				path:     "id",
				children: make([]*node, 0),
				handleChains: []HandleFunc{
					mockHandler1,
				},
			},
			shouldFound: true,
		},
		{
			name:   "get node that not exists",
			method: http.MethodHead,
			path:   "/user/xxx",
			routes: []struct {
				method  string
				path    string
				handler HandleFunc
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
			wantNode:    nil,
			shouldFound: false,
		},
	}

	// mock 方案
	for _, tc := range testCases {
		r := newRouter()
		t.Log(tc.name)
		for _, route := range tc.routes {
			r.AddRoute(route.method, route.path, route.handler)
		}
		node, found := r.FindRoute(tc.method, tc.path)
		// 断言
		require.NoError(t, nodeEqual(tc.wantNode, node))
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

	if len(src.handleChains) != len(dst.handleChains) {
		return fmt.Errorf("src handleChains length: [%+v] not equal to dst handleChains length: [%+v]", src, dst)
	}

	for i := 0; i < len(src.children); i++ {
		err := nodeEqual(src.children[i], dst.children[i])
		if err != nil {
			return err
		}
	}

	for idx, handler := range src.handleChains {
		if reflect.TypeOf(handler) != reflect.TypeOf(dst.handleChains[idx]) {
			return fmt.Errorf("src handler: [%+v] not equal to dst handler: [%+v]", src, dst)
		}
	}
	return nil
}
