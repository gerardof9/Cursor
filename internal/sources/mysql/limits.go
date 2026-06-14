package mysql

const (
	// LargeFileSizeBytes triggers Entire-file warning (1 GiB).
	LargeFileSizeBytes int64 = 1 << 30
	// LargeFileEventCount triggers Entire-file warning.
	LargeFileEventCount int64 = 500_000
)
