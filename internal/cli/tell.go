package cli

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/saga-changelog/saga/internal/chapter"
	"github.com/saga-changelog/saga/internal/dispatch"
	"github.com/saga-changelog/saga/internal/feat"
	"github.com/saga-changelog/saga/internal/taletext"
	"github.com/saga-changelog/saga/pkg/courier"
)

// TellCmd dispatches couriers for a chapter.
type TellCmd struct {
	Version  string   `arg:"" help:"Chapter version to tell."`
	DryRun   bool     `help:"Preview what would be sent without dispatching." name:"dry-run"`
	Audience string   `help:"Only tell tales for a specific audience." name:"audience"`
	Routes   []string `help:"Only dispatch specific routes within the audience (requires --audience)." name:"routes"`
}

// taleDelivery describes a single tale to be sent via a single route.
// Saga invokes the courier once per delivery so that each tale becomes
// its own message on the destination platform.
type taleDelivery struct {
	audience      string
	routeName     string
	qualifiedName string
	courierName   string
	payload       *courier.Payload
}

// deliveryID identifies a delivery in error messages: "audience/route [feat]".
func (d taleDelivery) deliveryID() string {
	return fmt.Sprintf("%s [%s]", d.qualifiedName, d.payload.Tale.Feat)
}

func (cmd *TellCmd) Run() error {
	if len(cmd.Routes) > 0 && cmd.Audience == "" {
		return fmt.Errorf("--route requires --audience")
	}

	// Load config and chapter.
	dir, cfg, err := loadConfig()
	if err != nil {
		return err
	}

	ch, err := chapter.Load(dir, cmd.Version)
	if err != nil {
		return err
	}

	// Validate.
	var allErrs []error
	allErrs = append(allErrs, cfg.Validate()...)

	validAudiences := make(map[string]bool)
	for _, a := range cfg.Audiences {
		validAudiences[a.Name] = true
	}
	for slug, f := range ch.Feats {
		allErrs = append(allErrs, f.Validate(slug, validAudiences)...)
	}
	if len(allErrs) > 0 {
		fmt.Fprintf(os.Stderr, "Validation failed with %d error(s):\n", len(allErrs))
		for _, e := range allErrs {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
		return fmt.Errorf("validation failed")
	}

	// Build route filter.
	routeFilter := make(map[string]bool, len(cmd.Routes))
	for _, r := range cmd.Routes {
		routeFilter[r] = true
	}

	// Collect which audiences have tales in this chapter.
	audiencesWithTales := make(map[string]bool)
	for _, f := range ch.Feats {
		for _, t := range f.Tales {
			audiencesWithTales[t.Audience] = true
		}
	}

	// Build one delivery per (route, tale) pair.
	var deliveries []taleDelivery
	for _, ar := range cfg.AllRoutes() {
		if !audiencesWithTales[ar.Audience] {
			continue
		}
		if cmd.Audience != "" && ar.Audience != cmd.Audience {
			continue
		}
		if len(routeFilter) > 0 && !routeFilter[ar.Route.Name] {
			continue
		}

		tales, err := gatherTales(ch.Feats, ar.Audience)
		if err != nil {
			return fmt.Errorf("%s: %w", ar.QualifiedName(), err)
		}
		if len(tales) == 0 {
			continue
		}

		for _, t := range tales {
			deliveries = append(deliveries, taleDelivery{
				audience:      ar.Audience,
				routeName:     ar.Route.Name,
				qualifiedName: ar.QualifiedName(),
				courierName:   ar.Route.Courier.Name,
				payload: &courier.Payload{
					Chapter: ch.Version,
					Route: courier.Route{
						Name:     ar.Route.Name,
						Audience: ar.Audience,
						Config:   ar.Route.Courier.Config,
					},
					Tale: t,
				},
			})
		}
	}

	if len(deliveries) == 0 {
		fmt.Println("No tales to tell for the given filters.")
		return nil
	}

	// Dry run: print what would be sent.
	if cmd.DryRun {
		cmd.printDryRun(deliveries)
		return nil
	}

	// Validate each route's credentials and connectivity before dispatching.
	fmt.Fprintf(os.Stderr, "Validating routes...\n")
	validated := make(map[string]bool)
	for _, d := range deliveries {
		if validated[d.qualifiedName] {
			continue
		}
		if err := dispatch.ValidateRoute(d.courierName, &d.payload.Route); err != nil {
			return fmt.Errorf("%s: %w", d.qualifiedName, err)
		}
		validated[d.qualifiedName] = true
	}

	// Dispatch: one courier invocation per tale.
	var failed []string
	for _, d := range deliveries {
		fmt.Fprintf(os.Stderr, "Telling %s...\n", d.deliveryID())
		if err := dispatch.Tell(d.courierName, d.payload); err != nil {
			fmt.Fprintf(os.Stderr, "  FAILED: %s\n", err)
			failed = append(failed, d.deliveryID())
		}
	}

	// Report.
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%d tales dispatched", len(deliveries))
	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, ", %d failed: %s", len(failed), strings.Join(failed, ", "))
	}
	fmt.Fprintln(os.Stderr)

	if len(failed) > 0 {
		return fmt.Errorf("%d tale(s) failed", len(failed))
	}
	return nil
}

func (cmd *TellCmd) printDryRun(deliveries []taleDelivery) {
	// Group deliveries by qualified route name for readable output.
	byRoute := make(map[string][]taleDelivery)
	var routeOrder []string
	courierSet := make(map[string]bool)
	for _, d := range deliveries {
		if _, seen := byRoute[d.qualifiedName]; !seen {
			routeOrder = append(routeOrder, d.qualifiedName)
		}
		byRoute[d.qualifiedName] = append(byRoute[d.qualifiedName], d)
		courierSet[d.courierName] = true
	}

	fmt.Printf("Chapter %s - dry run\n", cmd.Version)
	fmt.Println()

	for _, qname := range routeOrder {
		group := byRoute[qname]
		first := group[0]

		// Find a representative config value for display.
		configHint := ""
		for _, v := range first.payload.Route.Config {
			configHint = v
			break
		}

		fmt.Printf("  Route: %s\n", qname)
		if configHint != "" {
			fmt.Printf("    Courier: %s → %s\n", first.courierName, configHint)
		} else {
			fmt.Printf("    Courier: %s\n", first.courierName)
		}
		fmt.Printf("    Tales:\n")
		for _, d := range group {
			title := d.payload.Tale.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			fmt.Printf("      - %s: %q\n", d.payload.Tale.Feat, title)
		}
		fmt.Println()
	}

	fmt.Printf("%d tales across %d routes, %d couriers\n",
		len(deliveries), len(routeOrder), len(courierSet))
}

// gatherTales collects tales for a specific audience across all feats in
// a chapter, parsing each tale's body into the structured form couriers
// consume. Tales are returned in deterministic feat-slug order.
func gatherTales(feats map[string]*feat.Feat, audience string) ([]courier.Tale, error) {
	slugs := slices.Sorted(maps.Keys(feats))

	var tales []courier.Tale
	for _, slug := range slugs {
		f := feats[slug]
		for _, t := range f.Tales {
			if t.Audience != audience {
				continue
			}
			text, err := taletext.Parse(t.Text)
			if err != nil {
				return nil, fmt.Errorf("%s: tale for %q: %w", slug, audience, err)
			}
			tales = append(tales, courier.Tale{
				Feat:  slug,
				Title: strings.TrimSpace(t.Title),
				Text:  text,
			})
		}
	}
	return tales, nil
}
