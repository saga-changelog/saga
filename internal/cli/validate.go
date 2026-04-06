package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/saga-changelog/saga/internal/config"
	"github.com/saga-changelog/saga/internal/dispatch"
	"github.com/saga-changelog/saga/internal/feat"
	"github.com/saga-changelog/saga/pkg/courier"
)

// ValidateCmd validates all configuration and feat files.
type ValidateCmd struct{}

func (cmd *ValidateCmd) Run() error {
	dir, cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("saga validate: %w", err)
	}

	var allErrs []error

	// Validate config structure.
	allErrs = append(allErrs, cfg.Validate()...)

	// Build valid audience set.
	validAudiences := make(map[string]bool)
	for _, a := range cfg.Audiences {
		validAudiences[a.Name] = true
	}

	// Validate pending feats.
	pendingDir := filepath.Join(dir, "feats", "pending")
	pendingErrs := validateFeatsInDir(dir, pendingDir, validAudiences)
	allErrs = append(allErrs, pendingErrs...)

	// Validate chaptered feats.
	chaptersDir := filepath.Join(dir, "feats", "chapters")
	chapterEntries, err := os.ReadDir(chaptersDir)
	if err != nil && !os.IsNotExist(err) {
		allErrs = append(allErrs, fmt.Errorf("reading chapters directory: %w", err))
	}
	for _, e := range chapterEntries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if e.Name() == "pending" {
			allErrs = append(allErrs, fmt.Errorf("chapter name %q is reserved", e.Name()))
			continue
		}
		chapterDir := filepath.Join(chaptersDir, e.Name())
		chapterErrs := validateFeatsInDir(dir, chapterDir, validAudiences)
		for _, err := range chapterErrs {
			allErrs = append(allErrs, fmt.Errorf("chapter %s: %w", e.Name(), err))
		}
	}

	// Validate courier bindings and per-route config values.
	allErrs = append(allErrs, validateCouriers(cfg)...)

	if len(allErrs) > 0 {
		fmt.Fprintf(os.Stderr, "Validation failed with %d error(s):\n\n", len(allErrs))
		for _, e := range allErrs {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
		return fmt.Errorf("validation failed")
	}

	fmt.Println("All files valid.")
	return nil
}

func validateFeatsInDir(sagaDir, dir string, validAudiences map[string]bool) []error {
	feats, err := feat.LoadDir(sagaDir, dir)
	if err != nil {
		return []error{err}
	}

	var errs []error
	for slug, f := range feats {
		errs = append(errs, f.Validate(slug, validAudiences)...)
	}
	return errs
}

// validateCouriers resolves each route's courier, checks that the binary is
// installed and responds to "info", and verifies that the route's config
// values match the regex patterns declared by the courier. Info is cached
// so a courier referenced by multiple routes is only invoked once.
func validateCouriers(cfg *config.Config) []error {
	type compiledKey struct {
		key courier.ConfigKey
		re  *regexp.Regexp // nil when regex is empty or invalid (invalid recorded as error)
	}
	type cacheEntry struct {
		info *courier.Info
		keys []compiledKey
		err  error
	}
	cache := make(map[string]cacheEntry)

	var errs []error
	for _, ar := range cfg.AllRoutes() {
		name := ar.Route.Courier.Name
		entry, cached := cache[name]
		if !cached {
			info, infoErr := dispatch.Info(name)
			entry = cacheEntry{info: info, err: infoErr}
			if infoErr == nil {
				for _, key := range info.ConfigKeys {
					ck := compiledKey{key: key}
					if key.Regex != "" {
						re, err := regexp.Compile(key.Regex)
						if err != nil {
							errs = append(errs, fmt.Errorf("courier %q config key %q has invalid regex %q: %v",
								name, key.Name, key.Regex, err))
						} else {
							ck.re = re
						}
					}
					entry.keys = append(entry.keys, ck)
				}
			}
			cache[name] = entry
		}

		qualified := ar.QualifiedName()
		if entry.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", qualified, entry.err))
			continue
		}

		for _, ck := range entry.keys {
			if ck.re == nil {
				continue
			}
			value := ar.Route.Courier.Config[ck.key.Name]
			if ck.re.MatchString(value) {
				continue
			}
			if value == "" {
				errs = append(errs, fmt.Errorf("%s: config key %q is required (must match %s)",
					qualified, ck.key.Name, ck.key.Regex))
			} else {
				errs = append(errs, fmt.Errorf("%s: config key %q value %q does not match required pattern %s",
					qualified, ck.key.Name, value, ck.key.Regex))
			}
		}
	}
	return errs
}
