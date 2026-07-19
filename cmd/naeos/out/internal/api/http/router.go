// Copyright 2026 NAEOS Foundation
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

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
