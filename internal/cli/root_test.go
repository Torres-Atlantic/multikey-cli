package cli

import "testing"

func TestFirstRunEligible(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		// Commands that must work without config never trigger onboarding.
		{"version subcommand", []string{"version"}, false},
		{"help subcommand", []string{"help"}, false},
		{"long help flag", []string{"--help"}, false},
		{"short help flag", []string{"-h"}, false},
		{"help flag on a subcommand", []string{"scan", "--help"}, false},
		{"completion generation", []string{"completion", "zsh"}, false},
		// Bare invocation and functional commands still onboard on first run.
		{"bare invocation", []string{}, true},
		{"functional command scan", []string{"scan", "/tmp"}, true},
		{"functional command profile add", []string{"profile", "add"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstRunEligible(tt.args); got != tt.want {
				t.Errorf("firstRunEligible(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
