package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadOptions configures the download operation
type DownloadOptions struct {
	// URL to download from
	URL string

	// Expected checksum for verification (format: "algorithm:hash")
	Checksum string

	// Directory to save the downloaded file
	DestDir string

	// Filename to save as (if empty, derived from URL)
	Filename string

	// Whether to show progress
	ShowProgress bool
}

// Result contains information about the downloaded file
type Result struct {
	// Full path to the downloaded file
	FilePath string

	// Size of the downloaded file in bytes
	Size int64

	// Calculated checksum of the file
	Checksum string
}

// Download downloads a file from a URL with progress reporting and checksum verification
func Download(opts DownloadOptions) (*Result, error) {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(opts.DestDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Determine filename from URL if not specified
	if opts.Filename == "" {
		opts.Filename = filepath.Base(opts.URL)
	}

	// Full path to the downloaded file
	destPath := filepath.Join(opts.DestDir, opts.Filename)

	// Create the file
	out, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Initialize variables for checksum calculation
	var hasher hash.Hash
	var resultChecksum string
	var writer io.Writer = out

	// Set up checksum verification if requested
	if opts.Checksum != "" {
		parts := strings.Split(opts.Checksum, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid checksum format, expected 'algorithm:hash'")
		}

		algorithm := strings.ToLower(parts[0])
		if algorithm != "sha256" {
			return nil, fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
		}

		// Create SHA-256 hasher
		hasher = sha256.New()
		// Write to both file and hasher
		writer = io.MultiWriter(out, hasher)
	}

	// Copy data with optional progress reporting
	size, err := io.Copy(writer, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Verify checksum if provided
	if opts.Checksum != "" && hasher != nil {
		parts := strings.Split(opts.Checksum, ":")
		expectedChecksum := parts[1]
		actualChecksum := hex.EncodeToString(hasher.Sum(nil))
		resultChecksum = actualChecksum

		if !strings.EqualFold(actualChecksum, expectedChecksum) {
			// Remove the file if checksum verification fails
			os.Remove(destPath)
			return nil, fmt.Errorf("checksum verification failed: expected %s, got %s",
				expectedChecksum, actualChecksum)
		}
	}

	return &Result{
		FilePath: destPath,
		Size:     size,
		Checksum: resultChecksum,
	}, nil
}
