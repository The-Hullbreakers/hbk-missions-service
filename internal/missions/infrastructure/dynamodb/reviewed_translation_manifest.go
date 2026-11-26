package dynamodb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ValidateLocalBackfillEndpoint prevents the offline publisher from using AWS.
func ValidateLocalBackfillEndpoint(endpoint string) error {
	target, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid local loopback endpoint: %w", err)
	}
	host := target.Hostname()
	if target.Scheme != "http" ||
		(host != "localhost" && host != "127.0.0.1" && host != "::1") ||
		target.User != nil || (target.Path != "" && target.Path != "/") ||
		target.RawQuery != "" || target.Fragment != "" || target.Opaque != "" {
		return fmt.Errorf("backfill requires a plain HTTP loopback endpoint")
	}
	port, err := strconv.Atoi(target.Port())
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("backfill requires a loopback endpoint with an explicit port")
	}
	return nil
}

// LoadReviewedTranslationManifests rejects malformed and duplicate records
// before the caller creates a database client.
func LoadReviewedTranslationManifests(folder string) ([]ReviewedMissionTranslation, error) {
	if strings.TrimSpace(folder) == "" {
		return nil, fmt.Errorf("reviewed translation directory is required")
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("failed to read reviewed translations: %w", err)
	}
	records := make([]ReviewedMissionTranslation, 0)
	seen := make(map[string]bool)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if !entry.Type().IsRegular() {
			return nil, fmt.Errorf("translation manifest is not a regular file: %s", entry.Name())
		}
		path := filepath.Join(folder, entry.Name())
		fileRecords, err := readReviewedManifest(path)
		if err != nil {
			return nil, err
		}
		for _, record := range fileRecords {
			if err := validateReviewedTranslation(record); err != nil {
				return nil, fmt.Errorf("invalid translation %s#%s in %s: %w",
					record.DefinitionCode, record.DefinitionVersion, path, err)
			}
			identity := record.DefinitionCode + "#" + record.DefinitionVersion
			if seen[identity] {
				return nil, fmt.Errorf("duplicate reviewed translation: %s", identity)
			}
			seen[identity] = true
			records = append(records, record)
		}
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("reviewed translation directory has no records")
	}
	return records, nil
}

func readReviewedManifest(path string) ([]ReviewedMissionTranslation, error) {
	source, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open reviewed translation manifest: %w", err)
	}
	defer source.Close()
	var manifest struct {
		Translations []ReviewedMissionTranslation `json:"translations"`
	}
	decoder := json.NewDecoder(source)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("invalid reviewed translation manifest %s: %w", path, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("reviewed translation manifest %s has trailing content", path)
	}
	if len(manifest.Translations) == 0 {
		return nil, fmt.Errorf("reviewed translation manifest %s is empty", path)
	}
	return manifest.Translations, nil
}
