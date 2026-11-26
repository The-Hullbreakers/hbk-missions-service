package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// RequestEvidenceReadHandler issues presigned read grants for stored evidence
// objects. Only privileged reviewers may read evidence media.
type RequestEvidenceReadHandler struct {
	contracts ContractReader
	objects   EvidenceObjectReader
}

// NewRequestEvidenceReadHandler creates a handler with the required ports.
func NewRequestEvidenceReadHandler(contracts ContractReader, objects EvidenceObjectReader) *RequestEvidenceReadHandler {
	return &RequestEvidenceReadHandler{contracts: contracts, objects: objects}
}

// Handle authorizes the caller as a reviewer, validates that the media
// reference belongs to the contract and exists, and returns a read grant.
func (h *RequestEvidenceReadHandler) Handle(identity domain.CallerIdentity, contractID, mediaURI string) (ReadGrant, error) {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return ReadGrant{}, err
	}

	if !isReviewer(identity) {
		return ReadGrant{}, domain.ErrCallerNotAuthorized
	}

	uri, err := domain.NewMediaURI(mediaURI, contract.ContractID)
	if err != nil {
		return ReadGrant{}, err
	}

	exists, err := h.objects.Exists(uri)
	if err != nil {
		return ReadGrant{}, err
	}
	if !exists {
		return ReadGrant{}, domain.ErrEvidenceMediaInvalid
	}

	return h.objects.PresignRead(uri)
}

// isReviewer allows only privileged roles to read evidence media.
func isReviewer(identity domain.CallerIdentity) bool {
	return identity.HasRole("admin") || identity.HasRole("moderator")
}
