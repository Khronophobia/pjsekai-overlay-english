package pjsekaioverlay

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func encodeString(str string) string {
	bytes := utf16.Encode([]rune(str))
	encoded := make([]string, 1024)
	if len(str) > 1024 {
		panic("too long string")
	}
	for i := range encoded {
		var hex string
		if i >= len(bytes) {
			hex = fmt.Sprintf("%04x", 0)
		} else {
			hex = fmt.Sprintf("%02x%02x", bytes[i]&0xff, bytes[i]>>8)
		}

		encoded[i] = hex
	}

	return strings.Join(encoded, "")
}

//go:embed main.exo
var rawBaseExo []byte

//go:embed main_1080p.exo
var rawBaseExoHD []byte

func WriteExoFiles(assets string, destDir string, title string, description string) error {
	baseExo := string(rawBaseExo)
	baseExoHD := string(rawBaseExoHD)
	replacedExo := baseExo
	replacedExoHD := baseExoHD
	mapping := []string{
		"{assets}", strings.ReplaceAll(assets, "\\", "/"),
		"{dist}", strings.ReplaceAll(destDir, "\\", "/"),
		"{text:difficulty}", encodeString("MASTER"),
		"{text:title}", encodeString(title),
		"{text:description}", encodeString(description),
	}
	for i := range mapping {
		if i%2 == 0 {
			continue
		}
		if !strings.Contains(replacedExo, mapping[i-1]) {
			panic(fmt.Sprintf("failed to generate exo file (%s not found)", mapping[i-1]))
		}
		if !strings.Contains(replacedExoHD, mapping[i-1]) {
			panic(fmt.Sprintf("failed to generate exo file (%s not found)", mapping[i-1]))
		}
		replacedExo = strings.ReplaceAll(replacedExo, mapping[i-1], mapping[i])
		replacedExoHD = strings.ReplaceAll(replacedExoHD, mapping[i-1], mapping[i])
	}
	replacedExo = strings.ReplaceAll(replacedExo, "\n", "\r\n")
	replacedExoHD = strings.ReplaceAll(replacedExoHD, "\n", "\r\n")
	encodedExo, err := io.ReadAll(transform.NewReader(
		strings.NewReader(replacedExo), japanese.ShiftJIS.NewEncoder()))
	if err != nil {
		return fmt.Errorf("encoding failed (%w)", err)
	}
	encodedExoHD, err := io.ReadAll(transform.NewReader(
		strings.NewReader(replacedExoHD), japanese.ShiftJIS.NewEncoder()))
	if err != nil {
		return fmt.Errorf("encoding failed (%w)", err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "main.exo"),
		encodedExo,
		0644); err != nil {
		return fmt.Errorf("failed to write file (%w)", err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "main_1080p.exo"),
		encodedExoHD,
		0644); err != nil {
		return fmt.Errorf("failed to write file (%w)", err)
	}

	return nil
}
