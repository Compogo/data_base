package data_base

import "context"

type HookManager[T any] struct {
	hooks []Hook[T]
}

func NewHookManager[T any](hooks ...Hook[T]) *HookManager[T] {
	return &HookManager[T]{hooks: hooks}
}

func (m *HookManager[T]) RegisterHooks(hooks ...Hook[T]) {
	m.hooks = append(m.hooks, hooks...)
}

func (m *HookManager[T]) RegisterHook(hook Hook[T]) {
	m.hooks = append(m.hooks, hook)
}

func (m *HookManager[T]) Count() int {
	return len(m.hooks)
}

func (m *HookManager[T]) Fire(ctx context.Context, c Client, ts ...*T) (err error) {
	for _, hook := range m.hooks {
		if err = hook(ctx, c, ts...); err != nil {
			return err
		}
	}

	return nil
}
