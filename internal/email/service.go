package email

import (
	"context"
	"fmt"
)

type Service struct {
	renderer Renderer
	sender   Sender
	from     Address
}

func NewService(renderer Renderer, sender Sender, from Address) (*Service, error) {
	if renderer == nil {
		return nil, fmt.Errorf("renderer is required")
	}
	if sender == nil {
		return nil, fmt.Errorf("sender is required")
	}
	if err := from.Validate(); err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}

	return &Service{
		renderer: renderer,
		sender:   sender,
		from:     from,
	}, nil
}

func (s *Service) ComposeBookingMessage(templateType TemplateType, recipients []Address, data TemplateData) (Message, error) {
	rendered, err := s.renderer.Render(templateType, data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to compose booking email: %w", err)
	}

	message := Message{
		From:        s.from,
		To:          recipients,
		Subject:     rendered.Subject,
		Body:        rendered.Body,
		ContentType: DefaultContentType(),
	}
	if err := message.Validate(); err != nil {
		return Message{}, fmt.Errorf("failed to compose booking email: %w", err)
	}

	return message, nil
}

func (s *Service) SendBookingEmail(ctx context.Context, templateType TemplateType, recipients []Address, data TemplateData) error {
	message, err := s.ComposeBookingMessage(templateType, recipients, data)
	if err != nil {
		return err
	}
	if err := s.sender.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send booking email: %w", err)
	}
	return nil
}
