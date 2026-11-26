// Package s3 provides the AWS S3 adapter for evidence blob ports.
// Objects live in a private bucket; access flows only through short-lived
// presigned URLs bound to the exact key and content type.
package s3

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

const (
	uploadGrantTTL = 5 * time.Minute
	readGrantTTL   = 15 * time.Minute
)

// EvidenceStore implements EvidenceUploadPresigner and EvidenceObjectReader
// against a private S3 bucket.
type EvidenceStore struct {
	bucket  string
	client  *awss3.Client
	presign *awss3.PresignClient
}

// NewEvidenceStore creates the adapter for the given bucket and client.
func NewEvidenceStore(bucket string, client *awss3.Client) *EvidenceStore {
	return &EvidenceStore{
		bucket:  bucket,
		client:  client,
		presign: awss3.NewPresignClient(client),
	}
}

// PresignUpload returns a presigned PUT URL bound to the server-generated
// key and the declared content type, valid for five minutes.
func (s *EvidenceStore) PresignUpload(key, contentType string, _ int64) (application.UploadGrant, error) {
	out, err := s.presign.PresignPutObject(context.Background(), &awss3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, awss3.WithPresignExpires(uploadGrantTTL))
	if err != nil {
		return application.UploadGrant{}, fmt.Errorf("failed to presign evidence upload: %w", err)
	}
	return application.UploadGrant{URL: out.URL, MediaURI: domain.MediaURI(key)}, nil
}

// Exists reports whether the referenced object is stored in the bucket.
func (s *EvidenceStore) Exists(uri domain.MediaURI) (bool, error) {
	_, err := s.client.HeadObject(context.Background(), &awss3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(string(uri)),
	})
	if err == nil {
		return true, nil
	}
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check evidence object: %w", err)
}

// PresignRead returns a presigned GET URL for the stored object, valid for
// fifteen minutes.
func (s *EvidenceStore) PresignRead(uri domain.MediaURI) (application.ReadGrant, error) {
	out, err := s.presign.PresignGetObject(context.Background(), &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(string(uri)),
	}, awss3.WithPresignExpires(readGrantTTL))
	if err != nil {
		return application.ReadGrant{}, fmt.Errorf("failed to presign evidence read: %w", err)
	}
	return application.ReadGrant{URL: out.URL}, nil
}
