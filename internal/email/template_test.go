package email

import (
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	cases := []struct {
		name string
		tmpl string
		data interface{}
	}{
		{"delivery_created", "delivery_created.html", map[string]interface{}{"RecipientName": "Alice", "DeliveryID": 42}},
		{"delivery_delivered", "delivery_delivered.html", map[string]interface{}{"RecipientName": "Bob", "DeliveryID": 99}},
		{"damage_reported", "damage_reported.html", map[string]interface{}{"RecipientName": "Carol", "DeliveryID": 123}},
		{"export_ready", "export_ready.html", map[string]interface{}{"RecipientName": "Dave", "ExportURL": "https://example.com/export.csv"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			html, err := RenderTemplate(c.tmpl, c.data)
			if err != nil {
				t.Errorf("failed to render %s: %v", c.tmpl, err)
			}
			if len(html) == 0 {
				t.Errorf("empty output for %s", c.tmpl)
			}
		})
	}
}
