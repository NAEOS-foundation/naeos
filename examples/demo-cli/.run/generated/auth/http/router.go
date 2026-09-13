// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package http

import "net/http"

type Router struct {
	handler Handler
}

func NewRouter() *Router {
	return &Router{handler: Handler{}}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}
