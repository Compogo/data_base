package data_base

import "context"

type Hook[T any] func(context.Context, Client, ...*T) error
