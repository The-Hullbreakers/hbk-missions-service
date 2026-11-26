package domain

import "strings"

// MediaURI is a validated reference to a stored evidence object.
// Valid form: evidence/{contract_id}/{object_id}; never client-shaped URLs.
type MediaURI string

const evidenceKeyPrefix = "evidence/"

// NewMediaURI validates raw as an evidence object reference owned by contractID.
// It rejects empty values, foreign-contract prefixes, traversal segments, and
// any reference whose object id is missing or spans additional path levels.
func NewMediaURI(raw, contractID string) (MediaURI, error) {
	objectID, ok := strings.CutPrefix(raw, mediaPrefix(contractID))
	if !ok || objectID == "" || strings.Contains(objectID, "/") || strings.Contains(objectID, "..") {
		return "", ErrEvidenceMediaInvalid
	}
	return MediaURI(raw), nil
}

// AllowedMediaType reports whether a content type may be uploaded as evidence
// media: images and videos only.
func AllowedMediaType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")
}

// MediaPrefix returns the storage prefix for this contract's evidence objects.
func (c *MissionContract) MediaPrefix() string {
	return mediaPrefix(c.ContractID)
}

func mediaPrefix(contractID string) string {
	return evidenceKeyPrefix + contractID + "/"
}
