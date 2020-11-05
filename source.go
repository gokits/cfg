package cfg

import (
	"context"
)

type Source interface {
	Start() error
	Next(ctx context.Context, oldversion int64) (content []byte, version int64, ok bool)
	Close()
	String() string
}
