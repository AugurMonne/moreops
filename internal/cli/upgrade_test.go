package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"0.2.0", "0.1.0", 1},
		{"0.1.0", "0.2.0", -1},
		{"0.1.0", "0.1.0", 0},
		{"1.0.0", "0.9.9", 1},
		{"0.10.0", "0.9.0", 1},
		{"0.1.1", "0.1.0", 1},
		{"v0.2.0", "0.1.0", 1},
	}

	for _, tt := range tests {
		got := compareVersions(tt.v1, tt.v2)
		if got != tt.expected {
			t.Errorf("compareVersions(%q, %q) = %d; want %d", tt.v1, tt.v2, got, tt.expected)
		}
	}
}

func TestExtractBinaryFromTarGz(t *testing.T) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	content := []byte("#!/bin/sh\necho test\n")
	hdr := &tar.Header{
		Name: "moreops",
		Mode: 0755,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	tw.Close()
	gzw.Close()

	extracted, err := extractBinaryFromTarGz(&buf, "moreops")
	if err != nil {
		t.Fatalf("extractBinaryFromTarGz failed: %v", err)
	}

	if !bytes.Equal(extracted, content) {
		t.Errorf("extracted content does not match: got %s, want %s", extracted, content)
	}
}
