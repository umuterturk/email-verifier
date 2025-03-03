package validatortest

import (
	"emailvalidator/pkg/validator"
	"testing"
)

func TestSyntaxValidatorValidate(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "Valid standard email",
			email: "user@example.com",
			want:  true,
		},
		{
			name:  "Valid email with plus",
			email: "user+tag@example.com",
			want:  true,
		},
		{
			name:  "Valid email with unicode characters",
			email: "用户@例子.广告",
			want:  true,
		},
		{
			name:  "Valid email with unicode characters (Hindi)",
			email: "अजय@डाटा.भारत",
			want:  true,
		},
		{
			name:  "Valid email with single character local part",
			email: "J@example.com",
			want:  true,
		},
		{
			name:  "Valid email with single unicode character local part",
			email: "İ@example.com",
			want:  true,
		},
		{
			name:  "Invalid email starts with dot",
			email: ".user@example.com",
			want:  false,
		},
		{
			name:  "Invalid email ends with dot",
			email: "user.@example.com",
			want:  false,
		},
		{
			name:  "Invalid email - only dot in local part",
			email: ".@example.com",
			want:  false,
		},
		{
			name:  "Invalid email - no @",
			email: "invalid-email",
			want:  false,
		},
		{
			name:  "Invalid email - double dots",
			email: "john..doe@example.com",
			want:  false,
		},
		{
			name:  "Invalid email - spaces in quotes",
			email: "\"john doe\"@example.com",
			want:  false,
		},
		{
			name:  "Invalid email - empty",
			email: "",
			want:  false,
		},
		{
			name:  "Invalid email - multiple @",
			email: "user@domain@example.com",
			want:  false,
		},
	}

	validator := validator.NewSyntaxValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.Validate(tt.email)
			if got != tt.want {
				t.Errorf("SyntaxValidator.Validate(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}
