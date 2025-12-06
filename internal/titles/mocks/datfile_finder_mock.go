package mocks

// MockDatFileFinder はDatFileFinderのモック実装です
type MockDatFileFinder struct {
	FoundFile string
	Error     error
}

// Find はモック実装です
func (m *MockDatFileFinder) Find() (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	return m.FoundFile, nil
}
