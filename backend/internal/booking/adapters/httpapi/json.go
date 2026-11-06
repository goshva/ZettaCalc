package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// bytesReader оборачивает срез байт в io.Reader.
func bytesReader(b []byte) io.Reader { return bytes.NewReader(b) }

// quote сериализует строку как JSON без HTML-экранирования.
func quote(s string) string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return `""`
	}
	out := strings.TrimRight(buf.String(), "\n")
	return out
}

// stringList сериализует список строк компактно.
func stringList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i, item := range items {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(quote(item))
	}
	sb.WriteByte(']')
	return sb.String()
}

func boolLiteral(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func trimSpace(s string) string { return strings.TrimSpace(s) }
