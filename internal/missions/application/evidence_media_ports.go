package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// UploadGrant is a short-lived, policy-constrained grant to upload one
// evidence object. URL is the presigned endpoint; MediaURI is the
// server-generated reference the client echoes back at confirmation.
type UploadGrant struct {
	URL      string
	MediaURI domain.MediaURI
}

// ReadGrant is a short-lived grant to retrieve one evidence object.
type ReadGrant struct {
	URL string
}

// EvidenceUploadPresigner issues upload grants for evidence objects.
type EvidenceUploadPresigner interface {
	PresignUpload(key, contentType string, maxBytes int64) (UploadGrant, error)
}

// EvidenceObjectReader verifies stored evidence and issues read grants.
type EvidenceObjectReader interface {
	Exists(uri domain.MediaURI) (bool, error)
	PresignRead(uri domain.MediaURI) (ReadGrant, error)
}

// EvidenceBlobStore is the combined storage port the composition injects
// into the evidence use cases.
type EvidenceBlobStore interface {
	EvidenceUploadPresigner
	EvidenceObjectReader
}
