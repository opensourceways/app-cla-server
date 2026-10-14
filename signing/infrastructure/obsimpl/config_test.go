package obsimpl

import "testing"

func TestConfigSetDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty prefix gets default", "", "cla-pdfs"},
		{"custom prefix preserved", "my-pdfs", "my-pdfs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{PathPrefix: tt.input}
			cfg.SetDefault()
			if cfg.PathPrefix != tt.expected {
				t.Errorf("SetDefault: PathPrefix got %q, want %q", cfg.PathPrefix, tt.expected)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := &Config{PathPrefix: ""}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: unexpected error %v", err)
	}
	if cfg.PathPrefix != "cla-pdfs" {
		t.Errorf("Validate: PathPrefix got %q, want %q", cfg.PathPrefix, "cla-pdfs")
	}
}

func TestConfigObjectKey(t *testing.T) {
	cfg := &Config{PathPrefix: "cla-pdfs"}

	key := cfg.objectKey("507f1f77bcf86cd799439011")
	expected := "cla-pdfs/corp-signing/507f1f77bcf86cd799439011.pdf"
	if key != expected {
		t.Errorf("objectKey: got %q, want %q", key, expected)
	}
}

func TestConfigObjectKeyCustomPrefix(t *testing.T) {
	cfg := &Config{PathPrefix: "custom-prefix"}

	key := cfg.objectKey("abc123")
	expected := "custom-prefix/corp-signing/abc123.pdf"
	if key != expected {
		t.Errorf("objectKey: got %q, want %q", key, expected)
	}
}
