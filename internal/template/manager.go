package template

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"sync"
)

type Manager struct {
	templates map[string]*template.Template
	mutex     sync.RWMutex
}

// TemplateNames represents all available templates
const (
	LoginPage           = "login"
	RegisterPage        = "register"
	VerifyPage          = "verify"
	VerificationPending = "verification-pending"
)

// Create a new template manager
func NewManager() (*Manager, error) {
	m := &Manager{
		templates: make(map[string]*template.Template),
	}

	// Load all templates
	templates := map[string]string{
		LoginPage:           "templates/login.html",
		RegisterPage:        "templates/register.html",
		VerifyPage:          "templates/verify.html",
		VerificationPending: "templates/verification-pending.html",
	}

	for name, path := range templates {
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %v", name, err)
		}
		m.templates[name] = tmpl
	}

	return m, nil
}

// Render executes the template with the given name and data
func (m *Manager) Render(name string, data interface{}) (string, error) {
	m.mutex.RLock()
	tmpl, exists := m.templates[name]
	m.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("template %s not found", name)
	}

	// Create a buffer to store the rendered template
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template %s: %v", name, err)
	}

	return buf.String(), nil
}

// Get returns the template with the given name
func (m *Manager) Get(name string) (*template.Template, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	tmpl, exists := m.templates[name]
	if !exists {
		return nil, fmt.Errorf("template %s not found", name)
	}

	return tmpl, nil
}

// RenderToResponse renders the template directly to http.ResponseWriter
func (m *Manager) RenderToResponse(w http.ResponseWriter, name string, data interface{}) error {
	m.mutex.RLock()
	tmpl, exists := m.templates[name]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("template %s not found", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}
