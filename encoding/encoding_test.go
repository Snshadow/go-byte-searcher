package encoding

import (
	"os"
	"path/filepath"
	"testing"
)

var (
	utf8        = []byte("This is a utf-8 text.")
	utf8WithBOM = []byte("\ufeff" + "This is a utf-8 text with BOM.")
	utf16LE     = []byte{0xff, 0xfe, 0x54, 0x00, 0x68, 0x00, 0x69, 0x00, 0x73, 0x00, 0x20, 0x00,
		0x69, 0x00, 0x73, 0x00, 0x20, 0x00, 0x61, 0x00, 0x20, 0x00, 0x75, 0x00,
		0x74, 0x00, 0x66, 0x00, 0x2d, 0x00, 0x31, 0x00, 0x36, 0x00, 0x20, 0x00,
		0x4c, 0x00, 0x45, 0x00, 0x20, 0x00, 0x74, 0x00, 0x65, 0x00, 0x78, 0x00,
		0x74, 0x00, 0x2e, 0x00}
	utf16BE = []byte{0xfe, 0xff, 0x00, 0x54, 0x00, 0x68, 0x00, 0x69, 0x00, 0x73, 0x00, 0x20,
		0x00, 0x69, 0x00, 0x73, 0x00, 0x20, 0x00, 0x61, 0x00, 0x20, 0x00, 0x75,
		0x00, 0x74, 0x00, 0x66, 0x00, 0x2d, 0x00, 0x31, 0x00, 0x36, 0x00, 0x20,
		0x00, 0x42, 0x00, 0x45, 0x00, 0x20, 0x00, 0x74, 0x00, 0x65, 0x00, 0x78,
		0x00, 0x74, 0x00, 0x2e}
)

func TestGetFileEncodingTool(t *testing.T) {
	texts := [][]byte{utf8, utf8WithBOM, utf16LE, utf16BE}

	expected := []string{"UTF-8", "UTF-8 with BOM", "UTF-16LE", "UTF-16BE"}

	temp := t.TempDir()

	tempFile := filepath.Join(temp, "test.txt")

	for i, data := range texts {
		err := os.WriteFile(tempFile, data, 0644)
		if err != nil {
			t.Fatalf("Writing file %s failed with err %v\n", tempFile, err)
			continue
		}
		file, err := os.Open(tempFile)
		if err != nil {
			t.Fatalf("Failed to open file %s to get encoding type with err %v\n", tempFile, err)
		}

		_, encType, err := GetFileEncoder(file)
		if err != nil {
			t.Fatalf("Failed to get encoder for file %v with err %v\n", data, err)
		}

        if encType != expected[i] {
            t.Fatalf("Returned encoding type was not %q but %q\n", expected[i], encType)
        }

		file.Close()
	}
}
