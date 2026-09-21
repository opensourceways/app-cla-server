package controllers

import "testing"

// TestParseCorpAutoApprovalOption covers AC9: the PUT /auto-approval body must
// carry an "enabled" boolean field. A missing field, an unknown-fields-only
// body, an empty body, invalid JSON, or a non-bool value all must yield
// HTTP 400 with errCode error_parsing_api_body. A present bool (true/false)
// is parsed successfully.
func TestParseCorpAutoApprovalOption(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    bool
		wantErr bool
	}{
		{"enabled true", `{"enabled":true}`, true, false},
		{"enabled false", `{"enabled":false}`, false, false},
		{"missing field", `{}`, false, true},
		{"unknown field only", `{"foo":"bar"}`, false, true},
		{"non-bool value", `{"enabled":"yes"}`, false, true},
		{"non-bool number", `{"enabled":1}`, false, true},
		{"empty body", ``, false, true},
		{"invalid json", `{"enabled":`, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, fr := parseCorpAutoApprovalOption([]byte(tt.body))
			if tt.wantErr {
				if fr == nil {
					t.Fatalf("expected error, got nil and value %v", got)
				}
				if fr.errCode != errParsingApiBody {
					t.Errorf("errCode: got %q, want %q", fr.errCode, errParsingApiBody)
				}
				if fr.statusCode != 400 {
					t.Errorf("statusCode: got %d, want 400", fr.statusCode)
				}
				return
			}
			if fr != nil {
				t.Fatalf("unexpected error: %v (statusCode %d)", fr.reason, fr.statusCode)
			}
			if got != tt.want {
				t.Errorf("enabled: got %v, want %v", got, tt.want)
			}
		})
	}
}
