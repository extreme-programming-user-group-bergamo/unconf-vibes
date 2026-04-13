package common

import "fmt"

// RenderHeader renders a reusable room explorer header with conference context.
func RenderHeader(styles Styles, conferenceSlug string) string {
	contextValue := conferenceSlug
	if contextValue == "" {
		contextValue = "none selected"
	}

	return styles.Header.Render("UNCONF Room Explorer") + "\n" +
		styles.Subtle.Render(fmt.Sprintf("Conference: %s", contextValue))
}
