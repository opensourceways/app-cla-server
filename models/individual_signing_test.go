package models

import (
	"encoding/json"
	"testing"
)

// TestIndividualSignedJSONSerialization verifies that the Check API response
// always includes signed, version_matched and status fields so that legacy
// callers keep working and the new authoritative status field is always present.
//
// Requirement (#909): the response must look like
//
//	{"data":{"type":"corporation","signed":true,"version_matched":true,"status":"valid"}}
//
// for all three states (not_signed / valid / expired), signed and
// version_matched must be present even when false.
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
				Status:         "valid",
			},
			want: map[string]interface{}{
				"type":            "corporation",
				"signed":          true,
				"version_matched": true,
				"status":          "valid",
			},
		},
		{
			name: "expired keeps version_matched:false",
			in: IndividualSigned{
				Type:           "corporation",
				Signed:         true,
				VersionMatched: false,
				Status:         "expired",
			},
			want: map[string]interface{}{
				"type":            "corporation",
				"signed":          true,
				"version_matched": false,
				"status":          "expired",
			},
		},
		{
			name: "not_signed keeps signed:false and version_matched:false",
			in: IndividualSigned{
				Type:           "individual",
				Signed:         false,
				VersionMatched: false,
				Status:         "not_signed",
			},
			want: map[string]interface{}{
				"type":            "individual",
				"signed":          false,
				"version_matched": false,
				"status":          "not_signed",
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
		Status:         "valid",
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
