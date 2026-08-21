package models

import (
	"encoding/json"
	"testing"
)

// TestIndividualSignedJSONSerialization verifies that the Check API response
// always includes signed and version_matched fields so that legacy callers
// keep working.
//
// The signing states are fully covered by the two fields (refactor b8e3107):
//
//	{"data":{"type":"corporation","signed":true,"version_matched":true}}
//
// signed and version_matched must be present even when false, for all three
// states (not_signed / valid / expired).
func TestIndividualSignedJSONSerialization(t *testing.T) {
	tests := []struct {
		name string
		in   IndividualSigned
		want map[string]interface{}
	}{
		{
			name: "valid",
			in: IndividualSigned{
				Type:           "corporation",
				Signed:         true,
				VersionMatched: true,
			},
			want: map[string]interface{}{
				"type":            "corporation",
				"signed":          true,
				"version_matched": true,
			},
		},
		{
			name: "expired keeps version_matched:false",
			in: IndividualSigned{
				Type:           "corporation",
				Signed:         true,
				VersionMatched: false,
			},
			want: map[string]interface{}{
				"type":            "corporation",
				"signed":          true,
				"version_matched": false,
			},
		},
		{
			name: "not_signed keeps signed:false and version_matched:false",
			in: IndividualSigned{
				Type:           "individual",
				Signed:         false,
				VersionMatched: false,
			},
			want: map[string]interface{}{
				"type":            "individual",
				"signed":          false,
				"version_matched": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("json.Marshal returned unexpected error: %v", err)
			}

			var got map[string]interface{}
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("json.Unmarshal returned unexpected error: %v", err)
			}

			for k, wantVal := range tt.want {
				gotVal, ok := got[k]
				if !ok {
					t.Errorf("missing field %q in response: %s", k, string(raw))
					continue
				}
				if gotVal != wantVal {
					t.Errorf("field %q = %v (%T), want %v (%T)",
						k, gotVal, gotVal, wantVal, wantVal)
				}
			}

			if _, ok := got["_debug"]; ok {
				t.Errorf("_debug should be omitted when DebugInfo is nil, got: %s", string(raw))
			}
		})
	}
}

// TestIndividualSignedDebugInfoOmittedWhenNil ensures the debug payload is
// only exposed when populated (i.e. debug=true).
func TestIndividualSignedDebugInfoOmittedWhenNil(t *testing.T) {
	in := IndividualSigned{
		Type:           "individual",
		Signed:         true,
		VersionMatched: true,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("json.Marshal returned unexpected error: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal returned unexpected error: %v", err)
	}

	if _, ok := got["_debug"]; ok {
		t.Errorf("_debug must be omitted when nil, got: %s", string(raw))
	}
}
