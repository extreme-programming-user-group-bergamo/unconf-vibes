package service

import (
	"context"
	"strings"
)

type sessionContextKey string

const sessionClientMetadataContextKey sessionContextKey = "session_client_metadata"

func WithSessionClientMetadata(ctx context.Context, clientMetadata string) context.Context {
	if strings.TrimSpace(clientMetadata) == "" {
		return ctx
	}

	return context.WithValue(ctx, sessionClientMetadataContextKey, clientMetadata)
}

func sessionClientMetadataFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	value, ok := ctx.Value(sessionClientMetadataContextKey).(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(value)
}
