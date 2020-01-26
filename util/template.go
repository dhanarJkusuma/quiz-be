package util

import (
	"errors"
	"fmt"
	goTemplate "html/template"
	"net/http"
	"strings"
)

var (
	ErrTemplateNotFound = errors.New("html template not found")
)

type TemplateHandler struct {
	TemplatePath string
	Template     *goTemplate.Template
}

func ConcatString(strs ...string) string {
	return strings.Trim(strings.Join(strs, ""), " ")
}

func NewTemplateHandler(templatePath string) *TemplateHandler {
	funcMap := goTemplate.FuncMap{
		"concat": ConcatString,
	}

	templates := goTemplate.Must(
		goTemplate.
			New("").
			Funcs(funcMap).
			ParseGlob(
				fmt.Sprintf(
					"%s/*.gohtml",
					templatePath,
	)))

	return &TemplateHandler{
		Template:     templates,
		TemplatePath: templatePath,
	}
}

func (h *TemplateHandler) ServeTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	foundTemplate := h.Template.Lookup(templateName)
	if foundTemplate != nil {
		return foundTemplate.Execute(w, data)
	}
	return ErrTemplateNotFound
}
