package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

func FileExists(ctx context.Context, fileName string) (bool, error) {
	svc := s3.New(createSession(ctx))

	bucketName := ctx.Value("bucket_name").(string)

	_, fileExistsErr := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})
	if fileExistsErr != nil {
		if s3Err, ok := fileExistsErr.(awserr.Error); ok && s3Err.Code() == "NotFound" {
			return false, nil
		} else {
			return false, fmt.Errorf("failed to check if pokemon history exists already: %w", s3Err)
		}
	}

	return true, nil
}

func CreateFile(ctx context.Context, fileName string) error {
	svc := s3.New(createSession(ctx))

	bucketName := ctx.Value("bucket_name").(string)

	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return fmt.Errorf("failed to update published pokemon history: %w", err)
	}

	return nil
}
