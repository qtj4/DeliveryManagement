package email

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
)

// RenderTemplate renders an HTML email template with the given data.
func RenderTemplate(templateName string, data interface{}) (string, error) {
	tmplPath := filepath.Join("templates", templateName)
	tmpl, err := template.New(templateName).Funcs(template.FuncMap{
		"upper": func(s string) string { return strings.ToUpper(s) },
	}).ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
