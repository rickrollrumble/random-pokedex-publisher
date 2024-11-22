package cloud

import "context"

type FileBucket interface {
	FileExists(ctx context.Context, object string) (bool, error)
	CreateFile(ctx context.Context, object string) error
}
