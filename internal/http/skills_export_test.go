package http

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestParseSkillExportRequestFormatsAndIDs(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/skills/export?format=tgz&id=11111111-1111-1111-1111-111111111111&ids=22222222-2222-2222-2222-222222222222,33333333-3333-3333-3333-333333333333&include_system=true", nil)

	parsed, err := parseSkillExportRequest(req)
	if err != nil {
		t.Fatalf("parseSkillExportRequest() error = %v", err)
	}
	if parsed.Format.Canonical != "tar.gz" || parsed.Format.Extension != ".tar.gz" {
		t.Fatalf("format = %#v, want tar.gz alias", parsed.Format)
	}
	if len(parsed.IDs) != 3 {
		t.Fatalf("ids = %v, want 3 ids", parsed.IDs)
	}
	if !parsed.IncludeSystem {
		t.Fatal("include_system was not parsed")
	}
}

func TestParseSkillExportRequestRejectsUnsupportedFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/skills/export?format=rar", nil)

	if _, err := parseSkillExportRequest(req); err == nil {
		t.Fatal("parseSkillExportRequest() error = nil, want unsupported format error")
	}
}

func TestSkillExportArchiveWriters(t *testing.T) {
	for _, tc := range []struct {
		name       string
		format     string
		contentTyp string
		wantEntry  func(t *testing.T, data []byte)
	}{
		{
			name:       "zip",
			format:     "zip",
			contentTyp: "application/zip",
			wantEntry: func(t *testing.T, data []byte) {
				t.Helper()
				zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
				if err != nil {
					t.Fatalf("zip.NewReader: %v", err)
				}
				names := make([]string, 0, len(zr.File))
				for _, f := range zr.File {
					names = append(names, f.Name)
				}
				if !slices.Contains(names, "skills/demo/SKILL.md") {
					t.Fatalf("zip entries = %v, want SKILL.md", names)
				}
			},
		},
		{
			name:       "tar.gz",
			format:     "tar.gz",
			contentTyp: "application/gzip",
			wantEntry: func(t *testing.T, data []byte) {
				t.Helper()
				gr, err := gzip.NewReader(bytes.NewReader(data))
				if err != nil {
					t.Fatalf("gzip.NewReader: %v", err)
				}
				defer gr.Close()
				tr := tar.NewReader(gr)
				for {
					hdr, err := tr.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatalf("tar.Next: %v", err)
					}
					if hdr.Name == "skills/demo/SKILL.md" {
						return
					}
				}
				t.Fatal("tar missing skills/demo/SKILL.md")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer, err := newSkillArchiveWriter(&buf, mustSkillExportFormat(t, tc.format))
			if err != nil {
				t.Fatalf("newSkillArchiveWriter() error = %v", err)
			}
			if writer.ContentType() != tc.contentTyp {
				t.Fatalf("ContentType() = %q, want %q", writer.ContentType(), tc.contentTyp)
			}
			if err := writer.AddFile("skills/demo/SKILL.md", []byte("# Demo")); err != nil {
				t.Fatalf("AddFile: %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}
			tc.wantEntry(t, buf.Bytes())
		})
	}
}

func TestAddSkillDirectoryToArchiveSkipsSymlinksAndResources(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "SKILL.md"), "# Demo")
	mustWriteFile(t, filepath.Join(root, "references", "guide.md"), "guide")
	mustWriteFile(t, filepath.Join(root, "scripts", "run.sh"), "#!/bin/sh")
	mustWriteFile(t, filepath.Join(root, "assets", "logo.txt"), "logo")
	mustWriteFile(t, filepath.Join(root, ".DS_Store"), "junk")
	if err := os.Symlink(filepath.Join(root, "SKILL.md"), filepath.Join(root, "linked.md")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	var buf bytes.Buffer
	writer, err := newSkillArchiveWriter(&buf, mustSkillExportFormat(t, "zip"))
	if err != nil {
		t.Fatalf("newSkillArchiveWriter() error = %v", err)
	}
	if err := addSkillDirectoryToArchive(writer, root, "skills/demo/"); err != nil {
		t.Fatalf("addSkillDirectoryToArchive() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	for _, want := range []string{
		"skills/demo/SKILL.md",
		"skills/demo/references/guide.md",
		"skills/demo/scripts/run.sh",
		"skills/demo/assets/logo.txt",
	} {
		if !slices.Contains(names, want) {
			t.Fatalf("entries = %v, missing %s", names, want)
		}
	}
	if slices.Contains(names, "skills/demo/.DS_Store") || slices.Contains(names, "skills/demo/linked.md") {
		t.Fatalf("entries include skipped artifacts: %v", names)
	}
}

func mustSkillExportFormat(t *testing.T, raw string) skillExportFormat {
	t.Helper()
	format, err := parseSkillExportFormat(raw)
	if err != nil {
		t.Fatalf("parseSkillExportFormat(%q): %v", raw, err)
	}
	return format
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
