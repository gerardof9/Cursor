package mysql

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-mysql-org/go-mysql/replication"

	"db-log-explorer/internal/events"
)

// SourceState tracks binlog source lifecycle.
type SourceState string

const (
	StateOpening   SourceState = "opening"
	StateAnalyzing SourceState = "analyzing"
	StateIndexing  SourceState = "indexing"
	StateReady     SourceState = "ready"
	StateError     SourceState = "error"
)

// Source is an opened MySQL binlog file.
type Source struct {
	ID           string
	Path         string
	file         *os.File
	FileSize     int64
	State        SourceState
	Error        string
	IndexedCount int
	BytesRead    int64
	Analysis     *events.FileAnalysisResult
	ActiveScope  *events.InvestigationScope
	parser       *replication.BinlogParser
	mu           sync.Mutex
}

// newParser returns a fresh binlog parser configured for this source.
func (s *Source) newParser() *replication.BinlogParser {
	p := replication.NewBinlogParser()
	p.SetVerifyChecksum(true)
	return p
}

// withFileLock serializes all reads on the shared binlog file handle.
func (s *Source) withFileLock(fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn()
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

	if err := validateBinlogHeader(f, parser); err != nil {
		f.Close()
		return nil, fmt.Errorf("parse binlog header: %w", err)
	}

	if pos, err := f.Seek(0, io.SeekCurrent); err == nil {
		src.BytesRead = pos
	}

	src.State = StateAnalyzing
	return src, nil
}

func validateBinlogHeader(f *os.File, parser *replication.BinlogParser) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind file: %w", err)
	}

	hdr := make([]byte, 4)
	if _, err := io.ReadFull(f, hdr); err != nil {
		return fmt.Errorf("read magic: %w", err)
	}
	if !bytes.Equal(hdr, replication.BinLogFileHeader) {
		return fmt.Errorf("not a valid MySQL binlog file")
	}

	var gotFormat bool
	_, err := parser.ParseSingleEvent(f, func(ev *replication.BinlogEvent) error {
		if _, ok := ev.Event.(*replication.FormatDescriptionEvent); ok {
			gotFormat = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !gotFormat {
		return fmt.Errorf("missing format description event")
	}
	return nil
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

// AnalysisProgress returns 0-100 analysis percent estimate.
func (s *Source) AnalysisProgress() int {
	if s.FileSize <= 0 {
		return 0
	}
	pct := int(s.BytesRead * 100 / s.FileSize)
	if pct > 99 {
		return 99
	}
	return pct
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
