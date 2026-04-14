package email

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
)

type TemplateType string

const (
	TemplateTypeNewBooking   TemplateType = "new_booking"
	TemplateTypeCancellation TemplateType = "cancellation"
	TemplateTypeModification TemplateType = "modification"
)

type TemplateData struct {
	GuestName       string
	GuestEmail      string
	RoomNumber      string
	StartDate       string
	EndDate         string
	SpecialRequests string
}

func (d TemplateData) Validate() error {
	if strings.TrimSpace(d.GuestName) == "" {
		return fmt.Errorf("guest name is required")
	}
	if strings.TrimSpace(d.GuestEmail) == "" {
		return fmt.Errorf("guest email is required")
	}
	if strings.TrimSpace(d.RoomNumber) == "" {
		return fmt.Errorf("room number is required")
	}
	if strings.TrimSpace(d.StartDate) == "" {
		return fmt.Errorf("start date is required")
	}
	if strings.TrimSpace(d.EndDate) == "" {
		return fmt.Errorf("end date is required")
	}

	return nil
}

type RenderedTemplate struct {
	Subject string
	Body    string
}

type Renderer interface {
	Render(templateType TemplateType, data TemplateData) (RenderedTemplate, error)
}

type Address struct {
	Email string
	Name  string
}

func (a Address) HeaderValue() string {
	email := strings.TrimSpace(a.Email)
	name := strings.TrimSpace(a.Name)
	if name == "" {
		return email
	}
	return (&mail.Address{
		Name:    name,
		Address: email,
	}).String()
}

func (a Address) Validate() error {
	name := strings.TrimSpace(a.Name)
	email := strings.TrimSpace(a.Email)

	if strings.ContainsAny(name, "\r\n") || strings.ContainsAny(email, "\r\n") {
		return fmt.Errorf("email address contains invalid characters")
	}

	if email == "" {
		return fmt.Errorf("email address is required")
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || parsed.Name != "" {
		return fmt.Errorf("email address is invalid")
	}

	return nil
}

type Message struct {
	From        Address
	To          []Address
	BCC         []Address
	Subject     string
	Body        string
	ContentType string
}

func (m Message) Validate() error {
	if err := m.From.Validate(); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}
	if len(m.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	for i := range m.To {
		if err := m.To[i].Validate(); err != nil {
			return fmt.Errorf("invalid recipient %d: %w", i, err)
		}
	}
	for i := range m.BCC {
		if err := m.BCC[i].Validate(); err != nil {
			return fmt.Errorf("invalid bcc recipient %d: %w", i, err)
		}
	}
	if strings.TrimSpace(m.Subject) == "" {
		return fmt.Errorf("subject is required")
	}
	if strings.TrimSpace(m.Body) == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

type Sender interface {
	Send(ctx context.Context, message Message) error
}
