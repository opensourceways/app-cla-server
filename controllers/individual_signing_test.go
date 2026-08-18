package controllers

import "testing"

func TestTokenPrefix(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"short", "short"},
		{"12345678", "12345678"},
		{"123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0", "12345678"},
	}

	for _, c := range cases {
		if got := tokenPrefix(c.in); got != c.want {
			t.Errorf("tokenPrefix(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
