package helper

import (
	"context"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ak-repo/stream-hub/config"
)

// =============================================================================
// CONFIG & ENVIRONMENT
// =============================================================================

// OverrideLocal ensures localhost usage in local development mode
func OverrideLocal(cfg *config.Config) {
	if cfg.App.Environment != "development" {
		return
	}

	cfg.Services.Gateway.Host = "localhost"
	cfg.Services.Auth.Host = "localhost"
	cfg.Services.Chat.Host = "localhost"
	cfg.Services.File.Host = "localhost"
	cfg.Services.Payment.Host = "localhost"
	cfg.Services.Admin.Host = "localhost"
	cfg.Redis.Host = "localhost"
	cfg.MinIO.Endpoint = "localhost:9000"
}

// GetEnv retrieves an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

// =============================================================================
// TIME HELPERS
// =============================================================================

// TimeToString converts time to a standard database-friendly string format
func TimeToString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	const layout = "2006-01-02 15:04:05"
	return t.Format(layout)
}

// StringToTime parses a string back to time object
func StringToTime(str string) (time.Time, error) {
	const layout = "2006-01-02 15:04:05"
	return time.Parse(layout, str)
}

// =============================================================================
// CONTEXT & TIMEOUTS
// =============================================================================

// WithGRPCTimeout returns a context with a standard timeout (default 5s)
// Use this for internal gRPC calls to prevent hanging routines.
func WithGRPCTimeout() (context.Context, context.CancelFunc) {

	return context.WithTimeout(context.Background(), 5*time.Second)
}

// =============================================================================
// TYPE CONVERSIONS
// =============================================================================

// StringToInt safely converts string to int, returning 0 on error
func StringToInt(str string) int {
	n, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return n
}

// StringToInt64 safely converts string to int64, returning 0 on error
func StringToInt64(str string) int64 {
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// StringToInt64 safely converts string to int64, returning 0 on error
func StringToInt32(str string) int32 {
	n, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0
	}
	return int32(n)
}

// IntToString converts int to string
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// Int64ToString converts int64 to string
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// =============================================================================
// DATA STRUCTURE HELPERS
// =============================================================================

// ContainsString checks if a slice contains a specific string
func ContainsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// CalculateOffset helps with pagination logic in Repositories
// page: 1-based index
func CalculateOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10 // Default limit
	}
	return (page - 1) * limit
}

// ParsePaginationDefaults ensures valid page/limit values for API requests
func ParsePaginationDefaults(page, limit int32) (int, int) {
	p := int(page)
	l := int(limit)

	if p <= 0 {
		p = 1
	}
	if l <= 0 || l > 100 {
		l = 10 // Default limit, max 100
	}
	return p, l
}

// SanitizeString removes leading/trailing spaces and lowers case (good for emails/usernames)
func SanitizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func BytesToMB(bytes int64) int64 {
	if bytes <= 0 {
		return 0
	}
	return int64(math.Ceil(float64(bytes) / 1_000_000))
}
