package pjsekaioverlay

import (
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	wapi "github.com/iamacarpet/go-win64api"
	so "github.com/iamacarpet/go-win64api/shared"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

//go:embed sekai.obj
var sekaiObj []byte

func TryInstallObject() bool {
	processes, _ := wapi.ProcessList()
	var aviutlProcess *so.Process
	for _, process := range processes {
		if process.Executable == "aviutl.exe" {
			aviutlProcess = &process
			break
		}
	}
	if aviutlProcess == nil {
		return false
	}
	var aviutlPath string
	aviutlPath = filepath.Dir(aviutlProcess.Fullpath)
	var exeditRoot string
	if _, err := os.Stat(filepath.Join(aviutlPath, "exedit.auf")); err == nil {
		exeditRoot = filepath.Join(aviutlPath)
	} else if _, err := os.Stat(filepath.Join(aviutlPath, "Plugins", "exedit.auf")); err == nil {
		exeditRoot = filepath.Join(aviutlPath, "Plugins")
	} else {
		return false
	}

	os.MkdirAll(filepath.Join(exeditRoot, "script"), 0755)

	var sekaiObjPath = filepath.Join(exeditRoot, "script", "@pjsekai-overlay.obj")
	if _, err := os.Stat(sekaiObjPath); err == nil {
		var sekaiObjFile, _ = os.OpenFile(sekaiObjPath, os.O_RDONLY, 0755)
		defer sekaiObjFile.Close()
		var sekaiObjDecoder = japanese.ShiftJIS.NewDecoder()
		var existingSekaiObj, _ = io.ReadAll(transform.NewReader(sekaiObjFile, sekaiObjDecoder))
		if strings.Contains(string(existingSekaiObj), "--version: "+Version) && Version != "0.0.0" {
			return false
		}
	}
	err := os.MkdirAll(filepath.Join(exeditRoot, "script"), 0755)
	if err != nil {
		return false
	}
	sekaiObjFile, err := os.Create(sekaiObjPath)
	if err != nil {
		return false
	}
	defer sekaiObjFile.Close()

	var sekaiObjWriter = transform.NewWriter(sekaiObjFile, japanese.ShiftJIS.NewEncoder())

	version_numbers := strings.Split(Version, ".")
	og_version_numbers := slices.Delete(version_numbers, 3, len(version_numbers))
	OGVersion := strings.Join(og_version_numbers, ".")

	strings.NewReader(strings.NewReplacer(
		"\r\n", "\r\n",
		"\r", "\r\n",
		"\n", "\r\n",
		"{version}", OGVersion,
	).Replace(string(sekaiObj))).WriteTo(sekaiObjWriter)
	return true
}
