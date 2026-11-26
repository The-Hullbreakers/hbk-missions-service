package s3_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"testing"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	infras3 "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/s3"
)

func s3TestEndpoint() string {
	if ep := os.Getenv("S3_TEST_ENDPOINT"); ep != "" {
		return ep
	}
	return "http://localhost:14566"
}

func newEvidenceStore(t *testing.T, bucket string) (*infras3.EvidenceStore, *awss3.Client) {
	t.Helper()
	ctx := context.Background()
	client, err := infras3.NewS3Client(ctx, "us-east-1", s3TestEndpoint())
	require.NoError(t, err)
	_, err = client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: &bucket})
	var owned *types.BucketAlreadyOwnedByYou
	var exists *types.BucketAlreadyExists
	require.True(t, err == nil || errors.As(err, &owned) || errors.As(err, &exists))
	return infras3.NewEvidenceStore(bucket, client), client
}

// TestEvidenceStore_RoundTrip proves the full grant flow against a real
// S3 endpoint: presigned PUT accepts the bytes, Exists sees the object,
// and the presigned GET returns identical bytes.
func TestEvidenceStore_RoundTrip(t *testing.T) {
	store, _ := newEvidenceStore(t, "hbk-evidence-test")
	key := "evidence/contract-x/obj-1"
	payload := []byte("fake-png-bytes")

	grant, err := store.PresignUpload(key, "image/png", int64(len(payload)))
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPut, grant.URL, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "image/png")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	exists, err := store.Exists(domain.MediaURI(key))
	require.False(t, err != nil || !exists)

	read, err := store.PresignRead(domain.MediaURI(key))
	require.NoError(t, err)
	getResp, err := http.Get(read.URL)
	require.NoError(t, err)
	got, err := io.ReadAll(getResp.Body)
	getResp.Body.Close()
	require.NoError(t, err)
	require.False(t, !bytes.Equal(got, payload))
}

// TestEvidenceStore_ExistsPhantom reports false (not an error) for a key
// that was never uploaded.
func TestEvidenceStore_ExistsPhantom(t *testing.T) {
	store, _ := newEvidenceStore(t, "hbk-evidence-test")

	exists, err := store.Exists(domain.MediaURI("evidence/contract-x/never-uploaded"))
	require.NoError(t, err)
	require.False(t, exists, "exists = true for phantom key")
}
