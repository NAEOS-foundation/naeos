// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package kernel

type EventBus interface {
	Publish(topic string, payload any)
	Subscribe(topic string, handler func(any)) error
}
