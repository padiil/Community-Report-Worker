package storage

import (
	"bytes"
	"context"
)

type StorageProvider interface {
	Save(ctx context.Context, reportType, filename string, file *bytes.Buffer) (string, error)
}
