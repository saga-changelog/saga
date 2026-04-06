package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/saga-changelog/saga/internal/theme"
)

// AudiencesCmd manages audiences.
type AudiencesCmd struct {
	List AudiencesListCmd `cmd:"" default:"withargs" help:"List configured audiences."`
}

// AudiencesListCmd lists configured audiences.
type AudiencesListCmd struct{}

func (cmd *AudiencesListCmd) Run() error {
	_, cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Audiences) == 0 {
		fmt.Println(theme.Faint.Render("No audiences configured."))
		return nil
	}

	fmt.Println(theme.Emphasis.Render("Audiences:"))
	fmt.Println()
	for _, a := range cfg.Audiences {
		fmt.Printf("  %s\n", theme.Emphasis.Render(a.Name))
		fmt.Printf("    %s\n", a.Description)
		fmt.Printf("    %s  %s\n", theme.Faint.Render("interest:"), a.Interest)
		fmt.Printf("    %s      %s\n", theme.Faint.Render("tone:"), a.Tone)
		if len(a.Routes) > 0 {
			fmt.Printf("    %s\n", theme.Faint.Render("routes:"))
			for _, r := range a.Routes {
				fmt.Printf("      %s  (courier: %s", r.Name, theme.CourierName.Render(r.Courier.Name))
				if len(r.Courier.Config) > 0 {
					fmt.Printf(", %s", formatConfig(r.Courier.Config))
				}
				fmt.Println(")")
			}
		}
		fmt.Println()
	}
	fmt.Println(theme.Faint.Render(fmt.Sprintf("%d audiences", len(cfg.Audiences))))

	return nil
}

func formatConfig(cfg map[string]string) string {
	keys := slices.Sorted(maps.Keys(cfg))
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + cfg[k]
	}
	return strings.Join(parts, ", ")
}
