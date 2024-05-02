package pprofile

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

func Expand(tmpl string) string {

	var data struct {
		PID        string
		TMP        string
		Executable string
	}

	data.PID = fmt.Sprintf("%d", os.Getpid())
	data.TMP = os.TempDir()

	exe, err := os.Executable()
	if err != nil {
		if dbglog() {
			log.Printf("[ERROR] %v", err)
		}
		if fail() {
			panic(err)
		}
		return ""
	}
	data.Executable = filepath.Base(exe)

	t := template.Must(template.New("template").Parse(tmpl))

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		if dbglog() {
			log.Printf("[ERROR] %v", err)
		}
		if fail() {
			panic(err)
		}
		return ""
	}
	return buf.String()
}
