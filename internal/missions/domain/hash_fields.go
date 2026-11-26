package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

// hashFields digests length-framed string components into one SHA-256 hex
// value. Framing each component ("<len>:<value>") makes field boundaries
// unambiguous, so concatenated content can never collide across fields or
// across a variable-length tail such as member lists.
func hashFields(fields []string) string {
	h := sha256.New()
	for _, f := range fields {
		h.Write([]byte(strconv.Itoa(len(f))))
		h.Write([]byte{':'})
		h.Write([]byte(f))
	}
	return hex.EncodeToString(h.Sum(nil))
}
