package mocks

import (
	"context"
)

// MockExtractor はExtractorのモック実装です
type MockExtractor struct {
	ExtractedFiles map[string][]byte
	Error          error
	CallCount      int
}

// ExtractFiles はモック実装です
func (m *MockExtractor) ExtractFiles(ctx context.Context, archivePath string, archiveType int, targetFiles []string) (map[string][]byte, error) {
	m.CallCount++
	if m.Error != nil {
		return nil, m.Error
	}
	return m.ExtractedFiles, nil
}
