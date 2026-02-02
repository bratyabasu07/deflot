package sources

import (
	"bufio"
	"context"
	"os"

	appCtx "github.com/bratyabasu07/deflot/internal/context"
)

type FileSource struct {
	path string
}

func NewFileSource(path string) *FileSource {
	return &FileSource{path: path}
}

func (s *FileSource) Name() string {
	return "file"
}

func (s *FileSource) NeedsKey() bool {
	return false
}

func (s *FileSource) Run(ctx context.Context, results chan<- appCtx.ScanRecord) {
	if s.path == "" {
		return
	}

	f, err := os.Open(s.path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()
			results <- appCtx.ScanRecord{
				URL:      line,
				Source:   "file",
				Category: "none",
			}
		}
	}
}
