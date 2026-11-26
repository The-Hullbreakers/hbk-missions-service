package application

import (
	"crypto/rand"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// RequestEvidenceUploadHandler issues a presigned upload grant scoped to the
// caller's contract. The object key is generated here; clients never name it.
type RequestEvidenceUploadHandler struct {
	contracts ContractReader
	presigner EvidenceUploadPresigner
	maxBytes  int64
}

// NewRequestEvidenceUploadHandler creates a handler with the required ports
// and the upload size bound enforced in the signed policy.
func NewRequestEvidenceUploadHandler(contracts ContractReader, presigner EvidenceUploadPresigner, maxBytes int64) *RequestEvidenceUploadHandler {
	return &RequestEvidenceUploadHandler{contracts: contracts, presigner: presigner, maxBytes: maxBytes}
}

// Handle authorizes the caller against the contract, checks the contract is
// IN_PROGRESS, validates the declared content type, and returns a grant
// whose key lives under the contract's media prefix.
func (h *RequestEvidenceUploadHandler) Handle(identity domain.CallerIdentity, contractID, contentType string, sizeBytes int64) (UploadGrant, error) {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return UploadGrant{}, err
	}

	if !canAdjudicate(identity, contract) {
		return UploadGrant{}, domain.ErrCallerNotAuthorized
	}

	if contract.State != domain.ContractInProgress {
		return UploadGrant{}, domain.ErrContractNotInProgress
	}

	if !domain.AllowedMediaType(contentType) {
		return UploadGrant{}, domain.ErrMediaTypeNotAllowed
	}

	if sizeBytes > h.maxBytes {
		return UploadGrant{}, domain.ErrEvidenceUploadTooLarge
	}

	return h.presigner.PresignUpload(contract.MediaPrefix()+newObjectID(), contentType, h.maxBytes)
}

// newObjectID produces a random object id segment for an evidence key.
func newObjectID() string {
	return rand.Text()
}
