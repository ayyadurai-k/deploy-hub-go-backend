package config

import (
	"os"
	"strings"
)

func CORSAllowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = "http://localhost:5173"
	}

	origins := make([]string, 0)
	for o := range strings.SplitSeq(raw, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed == "" {
			continue
		}
		// "*" cannot be combined with AllowCredentials: true — browsers reject it.
		if trimmed == "*" {
			continue
		}
		origins = append(origins, trimmed)
	}
	return origins
}
