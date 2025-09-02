package substitute

import (
	"bytes"
	"io"
	"os"
	"text/template"
)

// ContentsOf substitutes contents of the file at the filename with the substitutions.
func ContentsOf(path string, subs map[string]string) (string, error) {
	asBytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("map").Parse(string(asBytes))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	var w io.Writer = &buf

	err = tmpl.Execute(w, subs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
