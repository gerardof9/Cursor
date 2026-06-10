package mysql

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-mysql-org/go-mysql/replication"
)

// SourceState tracks binlog source lifecycle.
type SourceState string

const (
	StateOpening  SourceState = "opening"
	StateIndexing SourceState = "indexing"
	StateReady    SourceState = "ready"
	StateError    SourceState = "error"
)

// Source is an opened MySQL binlog file.
type Source struct {
	ID          string
	Path        string
	file        *os.File
	FileSize    int64
	State       SourceState
	Error       string
	IndexedCount int
	BytesRead   int64
	parser      *replication.BinlogParser
}

// OpenSource validates and opens a binlog file for indexing and detail reload.
func OpenSource(path string) (*Source, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory")
	}

	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	parser := replication.NewBinlogParser()
	parser.SetVerifyChecksum(true)

	src := &Source{
		ID:       sourceID(abs),
		Path:     abs,
		file:     f,
		FileSize: info.Size(),
		State:    StateOpening,
		parser:   parser,
	}

	// Validate by reading format description event.
	if _, err := parser.Parse(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("parse binlog header: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		f.Close()
		return nil, fmt.Errorf("rewind file: %w", err)
	}
	src.parser = replication.NewBinlogParser()
	src.parser.SetVerifyChecksum(true)
	src.BytesRead = 0

	src.State = StateIndexing
	return src, nil
}

func sourceID(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:6])
}

// File returns the open file handle for indexing/detail reads.
func (s *Source) File() *os.File {
	return s.file
}

// Parser returns the binlog parser instance for this source.
func (s *Source) Parser() *replication.BinlogParser {
	return s.parser
}

// MarkError transitions the source to error state.
func (s *Source) MarkError(err error) {
	s.State = StateError
	if err != nil {
		s.Error = err.Error()
	}
}

// MarkReady marks indexing complete.
func (s *Source) MarkReady() {
	s.State = StateReady
}

// Close closes the underlying file.
func (s *Source) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// IndexProgress returns 0-100 indexing percent estimate.
func (s *Source) IndexProgress() int {
	if s.FileSize <= 0 {
		if s.State == StateReady {
			return 100
		}
		return 0
	}
	if s.State == StateReady {
		return 100
	}
	pct := int(s.BytesRead * 100 / s.FileSize)
	if pct > 99 && s.State != StateReady {
		return 99
	}
	return pct
}
