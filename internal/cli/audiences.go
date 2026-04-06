package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"
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
		fmt.Println("No audiences configured.")
		return nil
	}

	fmt.Println("Audiences:")
	fmt.Println()
	for _, a := range cfg.Audiences {
		fmt.Printf("  %s\n", a.Name)
		fmt.Printf("    %s\n", a.Description)
		fmt.Printf("    interest:  %s\n", a.Interest)
		fmt.Printf("    tone:      %s\n", a.Tone)
		if len(a.Routes) > 0 {
			fmt.Printf("    routes:\n")
			for _, r := range a.Routes {
				fmt.Printf("      %s  (courier: %s", r.Name, r.Courier.Name)
				if len(r.Courier.Config) > 0 {
					fmt.Printf(", %s", formatConfig(r.Courier.Config))
				}
				fmt.Println(")")
			}
		}
		fmt.Println()
	}
	fmt.Printf("%d audiences\n", len(cfg.Audiences))

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
