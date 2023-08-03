package templatemap

import "fmt"

// TemplateMap is a typemap for the HTML templates.
type TemplateMap map[string]any

// Title sets the title on the template map. If a title already exists, the new
// value is prepended.
func (m TemplateMap) Title(f string, args ...any) {
	if f == "" {
		return
	}

	s := f
	if len(args) > 0 {
		s = fmt.Sprintf(f, args...)
	}

	if current := m["title"]; current != nil && current != "" {
		m["title"] = fmt.Sprintf("%s | %s", s, current)
		return
	}

	m["title"] = s
}
