package feat

import (
	"fmt"
	"strings"

	"github.com/saga-changelog/saga/internal/taletext"
)

// Validate checks a feat for structural errors against the given set of valid audience names.
func (f *Feat) Validate(slug string, validAudiences map[string]bool) []error {
	var errs []error

	if len(f.Tales) == 0 {
		errs = append(errs, fmt.Errorf("%s: no tales defined", slug))
	}

	seenAudiences := make(map[string]bool)
	for i, t := range f.Tales {
		if t.Audience == "" {
			errs = append(errs, fmt.Errorf("%s: tales[%d]: audience is empty", slug, i))
		} else {
			if seenAudiences[t.Audience] {
				errs = append(errs, fmt.Errorf("%s: tales[%d]: duplicate audience %q", slug, i, t.Audience))
			}
			seenAudiences[t.Audience] = true

			if !validAudiences[t.Audience] {
				errs = append(errs, fmt.Errorf("%s: tales[%d]: audience %q is not defined in config", slug, i, t.Audience))
			}
		}

		title := strings.TrimSpace(t.Title)
		if title == "" {
			errs = append(errs, fmt.Errorf("%s: tales[%d]: title is empty", slug, i))
		} else if strings.ContainsAny(title, "\r\n") {
			errs = append(errs, fmt.Errorf("%s: tales[%d]: title must be a single line", slug, i))
		}

		if strings.TrimSpace(t.Text) == "" {
			errs = append(errs, fmt.Errorf("%s: tales[%d]: text is empty", slug, i))
		} else if _, err := taletext.Parse(t.Text); err != nil {
			errs = append(errs, fmt.Errorf("%s: tales[%d]: text: %w", slug, i, err))
		}
	}

	return errs
}
