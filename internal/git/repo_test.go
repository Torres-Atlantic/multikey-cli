package git

import "testing"

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantHost string
		wantOrg  string
		wantRepo string
		wantErr  bool
	}{
		{
			name:     "git@ with .git suffix",
			url:      "git@github.com:torres-atlantic/multikey-cli.git",
			wantHost: "github.com",
			wantOrg:  "torres-atlantic",
			wantRepo: "multikey-cli",
		},
		{
			name:     "git@ without .git suffix",
			url:      "git@github.com:torres-atlantic/multikey-cli",
			wantHost: "github.com",
			wantOrg:  "torres-atlantic",
			wantRepo: "multikey-cli",
		},
		{
			name:     "git@ with host alias",
			url:      "git@github-work:acme/website.git",
			wantHost: "github-work",
			wantOrg:  "acme",
			wantRepo: "website",
		},
		{
			name:     "https with .git suffix",
			url:      "https://github.com/torres-atlantic/multikey-cli.git",
			wantHost: "github.com",
			wantOrg:  "torres-atlantic",
			wantRepo: "multikey-cli",
		},
		{
			name:     "https without .git suffix",
			url:      "https://github.com/torres-atlantic/multikey-cli",
			wantHost: "github.com",
			wantOrg:  "torres-atlantic",
			wantRepo: "multikey-cli",
		},
		{
			name:    "ssh:// scheme is unsupported",
			url:     "ssh://git@github.com/torres-atlantic/multikey-cli.git",
			wantErr: true,
		},
		{
			name:    "git@ missing colon",
			url:     "git@github.com",
			wantErr: true,
		},
		{
			name:    "git@ missing repo segment",
			url:     "git@github.com:torres-atlantic",
			wantErr: true,
		},
		{
			name:    "git@ with too many path segments",
			url:     "git@github.com:torres-atlantic/multikey-cli/extra",
			wantErr: true,
		},
		{
			name:    "https missing repo segment",
			url:     "https://github.com/torres-atlantic",
			wantErr: true,
		},
		{
			name:    "empty string",
			url:     "",
			wantErr: true,
		},
		{
			name:    "unrecognized scheme",
			url:     "ftp://github.com/torres-atlantic/multikey-cli",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, org, repo, err := ParseRemoteURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got host=%q org=%q repo=%q", tt.url, host, org, repo)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.url, err)
			}
			if host != tt.wantHost || org != tt.wantOrg || repo != tt.wantRepo {
				t.Errorf("ParseRemoteURL(%q) = (%q, %q, %q), want (%q, %q, %q)",
					tt.url, host, org, repo, tt.wantHost, tt.wantOrg, tt.wantRepo)
			}
		})
	}
}
