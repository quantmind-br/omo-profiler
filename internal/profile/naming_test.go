package profile

import "testing"

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "myprofile", false},
		{"valid with dash", "my-profile", false},
		{"valid with underscore", "my_profile", false},
		{"valid with numbers", "profile123", false},
		{"valid mixed", "My-Profile_123", false},
		{"empty", "", true},
		{"with spaces", "my profile", true},
		{"with slash", "my/profile", true},
		{"with dots", "my.profile", true},
		{"with special chars", "my@profile!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"already valid", "myprofile", "myprofile"},
		{"with spaces", "my profile", "myprofile"},
		{"with dots", "my.profile.json", "myprofilejson"},
		{"with slashes", "path/to/profile", "pathtoprofile"},
		{"with special chars", "my@profile!", "myprofile"},
		{"leading dash", "-profile", "profile"},
		{"trailing underscore", "profile_", "profile"},
		{"mixed", "  --my.profile!!  ", "myprofile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
