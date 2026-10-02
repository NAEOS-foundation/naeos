// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}
