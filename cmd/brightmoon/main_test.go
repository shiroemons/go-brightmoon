package main

import (
	"path/filepath"
	"testing"
)

func TestSafeOutputPath(t *testing.T) {
	tests := []struct {
		name      string
		entryName string
		want      string
		wantErr   bool
	}{
		{
			name:      "normal nested path",
			entryName: "bgm/th10_01.wav",
			want:      filepath.Join("out", "bgm", "th10_01.wav"),
		},
		{
			name:      "backslash path is normalized",
			entryName: "bgm\\th10_01.wav",
			want:      filepath.Join("out", "bgm", "th10_01.wav"),
		},
		{
			name:      "parent traversal is rejected",
			entryName: "../escape.txt",
			wantErr:   true,
		},
		{
			name:      "absolute path is rejected",
			entryName: "/tmp/escape.txt",
			wantErr:   true,
		},
		{
			name:      "drive path is rejected",
			entryName: "C:/escape.txt",
			wantErr:   true,
		},
		{
			name:      "empty name is rejected",
			entryName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeOutputPath("out", tt.entryName)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("safeOutputPath returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("safeOutputPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
