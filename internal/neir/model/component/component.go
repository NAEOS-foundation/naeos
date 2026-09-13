// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package component

type ComponentKind string

const (
	KindHandler    ComponentKind = "handler"
	KindService    ComponentKind = "service"
	KindRepository ComponentKind = "repository"
	KindMiddleware ComponentKind = "middleware"
	KindModel      ComponentKind = "model"
	KindConfig     ComponentKind = "config"
	KindWorker     ComponentKind = "worker"
	KindScheduler  ComponentKind = "scheduler"
)

type Component struct {
	Name         string            `json:"name"`
	Kind         ComponentKind     `json:"kind,omitempty"`
	Module       string            `json:"module,omitempty"`
	Description  string            `json:"description,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}
