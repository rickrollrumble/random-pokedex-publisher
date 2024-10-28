package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func createSession(ctx context.Context) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(ctx.Value("aws_region").(string)),
		Credentials: credentials.NewStaticCredentials(ctx.Value("aws_key_id").(string), ctx.Value("aws_secret").(string), ""),
	}))
}

func SaveFile(ctx context.Context, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	svc := s3.New(createSession(ctx))

	bucketName := ctx.Value("bucket_name").(string)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   file,
	})

	return err
}

func GetFile(ctx context.Context, fileName string) error {
	svc := s3.New(createSession(ctx))

	file, err := os.Create("published_pokemon.txt")
	if err != nil {
		return fmt.Errorf("failed to create file to store s3 object: %q: %w", fileName, err)
	}

	defer file.Close()

	objectKey := ctx.Value("object_key").(string)
	bucketName := ctx.Value("bucket_name").(string)

	_, err = svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("unable to download item %q from bucket %q: %w", objectKey, bucketName, err)
	}

	return nil
}
