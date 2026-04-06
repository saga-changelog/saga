package cli

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/saga-changelog/saga/internal/dispatch"
	"github.com/saga-changelog/saga/internal/theme"
	"github.com/saga-changelog/saga/pkg/courier"
)

// CouriersCmd manages courier plugins.
type CouriersCmd struct {
	List CouriersListCmd `cmd:"" default:"withargs" help:"List available courier plugins."`
}

// CouriersListCmd lists available courier plugins.
type CouriersListCmd struct{}

// discovered represents a courier binary found on PATH.
type discovered struct {
	name string
	path string
	info *courier.Info
	err  error
}

func (cmd *CouriersListCmd) Run() error {
	_, cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Collect courier names referenced by routes.
	referenced := make(map[string][]string) // courier name -> qualified route names
	for _, ar := range cfg.AllRoutes() {
		referenced[ar.Route.Courier.Name] = append(referenced[ar.Route.Courier.Name], ar.QualifiedName())
	}

	// Discover installed couriers on PATH.
	installed := make(map[string]discovered)
	for _, d := range discoverCouriers() {
		installed[d.name] = d
	}

	// Build sorted list of all known courier names (referenced + installed).
	names := make(map[string]bool)
	for n := range referenced {
		names[n] = true
	}
	for n := range installed {
		names[n] = true
	}
	sorted := slices.Sorted(maps.Keys(names))

	if len(sorted) == 0 {
		fmt.Println("No couriers referenced or installed.")
		return nil
	}

	fmt.Println("Couriers:")
	fmt.Println()

	var missing, errored int
	for _, name := range sorted {
		d, isInstalled := installed[name]
		routes := referenced[name]

		fmt.Printf("  %s\n", name)

		switch {
		case !isInstalled:
			missing++
			fmt.Printf("    %s", theme.Error.Render("not installed"))
			if len(routes) > 0 {
				fmt.Printf(" (required by: %s)", strings.Join(routes, ", "))
			}
			fmt.Println()
		case d.err != nil:
			errored++
			fmt.Printf("    %s at %s\n", theme.Warning.Render("info failed"), d.path)
			fmt.Printf("    error: %s\n", d.err)
		default:
			fmt.Printf("    %s at %s\n", theme.Success.Render("installed"), d.path)
			if d.info != nil && d.info.Description != "" {
				fmt.Printf("    %s\n", d.info.Description)
			}
			if d.info != nil && len(d.info.ConfigKeys) > 0 {
				fmt.Printf("    config keys:\n")
				for _, k := range d.info.ConfigKeys {
					required := ""
					if k.Regex != "" && !matchesEmpty(k.Regex) {
						required = " (required)"
					}
					fmt.Printf("      %s%s - %s\n", k.Name, required, k.Description)
				}
			}
			if len(routes) > 0 {
				fmt.Printf("    used by: %s\n", strings.Join(routes, ", "))
			}
		}
		fmt.Println()
	}

	total := len(sorted)
	installedCount := total - missing - errored
	fmt.Printf("%d installed, %d missing, %d errored (of %d referenced/installed)\n",
		installedCount, missing, errored, total)

	if missing > 0 || errored > 0 {
		return fmt.Errorf("%d courier(s) not usable", missing+errored)
	}
	return nil
}

// discoverCouriers scans PATH for saga-courier-* binaries and invokes each
// with the "info" subcommand. Results are sorted by name. Duplicate binaries
// across PATH entries are reported only once (first match wins).
func discoverCouriers() []discovered {
	const prefix = "saga-courier-"
	seen := make(map[string]bool)
	var results []discovered

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // unreadable PATH entries are ignored
		}
		for _, e := range entries {
			name := e.Name()
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			courierName := strings.TrimPrefix(name, prefix)
			if courierName == "" || seen[courierName] {
				continue
			}

			full := filepath.Join(dir, name)
			fi, statErr := os.Stat(full)
			if statErr != nil || fi.IsDir() || fi.Mode()&0o111 == 0 {
				continue
			}

			seen[courierName] = true

			d := discovered{name: courierName, path: full}
			info, err := dispatch.Info(courierName)
			if err != nil {
				d.err = err
			} else {
				d.info = info
			}
			results = append(results, d)
		}
	}

	slices.SortFunc(results, func(a, b discovered) int {
		return strings.Compare(a.name, b.name)
	})

	return results
}

// matchesEmptyCache caches compiled regexes for matchesEmpty.
var matchesEmptyCache = make(map[string]bool)

// matchesEmpty reports whether the regex matches an empty string,
// which marks the config key as optional.
func matchesEmpty(pattern string) bool {
	if result, ok := matchesEmptyCache[pattern]; ok {
		return result
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		matchesEmptyCache[pattern] = false
		return false
	}
	result := re.MatchString("")
	matchesEmptyCache[pattern] = result
	return result
}
