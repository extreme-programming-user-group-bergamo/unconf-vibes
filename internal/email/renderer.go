package email

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

const defaultContentType = "text/plain; charset=utf-8"

var (
	//go:embed templates/*.tmpl
	templateFS embed.FS

	templateFileByType = map[TemplateType]string{
		TemplateTypeNewBooking:   "templates/new_booking.tmpl",
		TemplateTypeCancellation: "templates/cancellation.tmpl",
		TemplateTypeModification: "templates/modification.tmpl",
	}

	templateSubjectByType = map[TemplateType]string{
		TemplateTypeNewBooking:   "New Booking Notification",
		TemplateTypeCancellation: "Booking Cancellation Notification",
		TemplateTypeModification: "Booking Modification Notification",
	}
)

type TemplateRenderer struct{}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

func (r *TemplateRenderer) Render(templateType TemplateType, data TemplateData) (RenderedTemplate, error) {
	if err := data.Validate(); err != nil {
		return RenderedTemplate{}, fmt.Errorf("failed to render %s template: %w", templateType, err)
	}

	templateFile, ok := templateFileByType[templateType]
	if !ok {
		return RenderedTemplate{}, fmt.Errorf("failed to render template: unsupported template type %q", templateType)
	}

	subject := templateSubjectByType[templateType]
	parsedTemplate, err := template.ParseFS(templateFS, templateFile)
	if err != nil {
		return RenderedTemplate{}, fmt.Errorf("failed to parse template %s: %w", templateFile, err)
	}

	var bodyBuilder bytes.Buffer
	if err := parsedTemplate.Execute(&bodyBuilder, data); err != nil {
		return RenderedTemplate{}, fmt.Errorf("failed to execute template %s: %w", templateFile, err)
	}

	return RenderedTemplate{
		Subject: subject,
		Body:    strings.TrimSpace(bodyBuilder.String()),
	}, nil
}

func DefaultContentType() string {
	return defaultContentType
}
