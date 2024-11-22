package gcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

var bucket = os.Getenv("GCP_BUCKET")

type Bucket struct{}

// getMetadata prints all of the object attributes.
func (b *Bucket) FileExists(ctx context.Context, object string) (bool, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to create google cloud client: %w", err)
	}
	defer client.Close()

	o := client.Bucket(bucket).Object(object)
	_, err = o.Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("failed to get object from bucket: %w", err)
	}
	return true, nil
}

func (b *Bucket) CreateFile(ctx context.Context, object string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create google cloud client: %w", err)
	}
	defer client.Close()

	f, err := os.Create(object)
	if err != nil {
		return fmt.Errorf("failed to create file %s to save in bucket: %w", object, err)
	}
	defer f.Close()
	defer os.Remove(object)

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := client.Bucket(bucket).Object(object)

	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("failed to copy object %s to bucket: %w", bucket, err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close writer after copying: %w", err)
	}

	return nil
}
