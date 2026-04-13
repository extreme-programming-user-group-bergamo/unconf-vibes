package common

import "strings"

// RenderFooter renders keyboard help and exit guidance.
func RenderFooter(styles Styles, hints []string) string {
	if len(hints) == 0 {
		hints = []string{"q: quit"}
	}

	parts := make([]string, 0, len(hints))
	for _, hint := range hints {
		segments := strings.SplitN(hint, ":", 2)
		if len(segments) != 2 {
			parts = append(parts, styles.Footer.Render(strings.TrimSpace(hint)))
			continue
		}

		key := styles.HelpKey.Render(strings.TrimSpace(segments[0]))
		label := styles.Footer.Render(strings.TrimSpace(segments[1]))
		parts = append(parts, key+": "+label)
	}

	return strings.Join(parts, "  •  ")
}
