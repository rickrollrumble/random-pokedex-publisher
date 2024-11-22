package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	bucketName         = os.Getenv("BUCKET_NAME")
	awsRegion          = os.Getenv("AWS_DEFAULT_REGION")
	awsAccessKey       = os.Getenv("AWS_ACCESS_KEY")
	awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
)

type Bucket struct{}

func createSession() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretAccessKey, ""),
	}))
}

func (b *Bucket) FileExists(ctx context.Context, fileName string) (bool, error) {
	svc := s3.New(createSession())

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

func (b *Bucket) CreateFile(ctx context.Context, fileName string) error {
	svc := s3.New(createSession())

	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return fmt.Errorf("failed to update published pokemon history: %w", err)
	}

	return nil
}
