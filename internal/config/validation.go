package config

import "fmt"

// Validate checks the config for structural errors.
// It returns all errors found rather than stopping at the first.
func (c *Config) Validate() []error {
	var errs []error

	// Check audience names are non-empty and unique.
	audienceNames := make(map[string]bool)

	for i, a := range c.Audiences {
		if a.Name == "" {
			errs = append(errs, fmt.Errorf("audiences[%d]: name is empty", i))
			continue
		}
		if audienceNames[a.Name] {
			errs = append(errs, fmt.Errorf("audiences[%d]: duplicate name %q", i, a.Name))
		}
		audienceNames[a.Name] = true

		// Check routes are unique within this audience.
		routeNames := make(map[string]bool)
		for j, r := range a.Routes {
			if r.Name == "" {
				errs = append(errs, fmt.Errorf("%s/routes[%d]: name is empty", a.Name, j))
			} else if routeNames[r.Name] {
				errs = append(errs, fmt.Errorf("%s/%s: duplicate route name", a.Name, r.Name))
			}
			routeNames[r.Name] = true

			if r.Courier.Name == "" {
				errs = append(errs, fmt.Errorf("%s/%s: courier name is empty", a.Name, r.Name))
			}
		}
	}

	return errs
}
