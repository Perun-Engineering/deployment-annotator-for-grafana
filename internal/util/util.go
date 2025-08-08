package util

import "strings"

// SanitizeForLog removes characters that could be used for log injection attacks
func SanitizeForLog(input string) string {
	sanitized := strings.ReplaceAll(input, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\t", "")
	return sanitized
}

// ExtractImageTag returns a human-friendly version tag from an image reference.
// Handles tags, digests, and registries with ports.
func ExtractImageTag(imageRef string) string {
	// Digest format: repo@sha256:abcdef...
	if at := strings.LastIndex(imageRef, "@"); at != -1 {
		digest := imageRef[at+1:]
		// Return short digest
		if colon := strings.Index(digest, ":"); colon != -1 && len(digest) > colon+7 {
			return digest[colon+1 : colon+8]
		}
		return digest
	}

	// Tag format: repo[:port]/name:tag
	lastSlash := strings.LastIndex(imageRef, "/")
	lastColon := strings.LastIndex(imageRef, ":")
	if lastColon != -1 && lastColon > lastSlash {
		return imageRef[lastColon+1:]
	}
	return "latest"
}
