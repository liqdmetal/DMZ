package signal

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The browser mirror and this package read the same fixture. A case added here must pass in both
// readers, so a rule cannot be tightened on one side and quietly left loose on the other.
type sharedVectors struct {
	Version int `json:"version"`
	Records []struct {
		Name      string          `json:"name"`
		Record    json.RawMessage `json:"record"`
		Valid     bool            `json:"valid"`
		Field     string          `json:"field"`
		Canonical string          `json:"canonical"`
		Live      []struct {
			Now    int64  `json:"now"`
			Height uint64 `json:"height"`
			Valid  bool   `json:"valid"`
			Field  string `json:"field"`
		} `json:"live"`
	} `json:"records"`
	JSON []struct {
		Name  string `json:"name"`
		Raw   string `json:"raw"`
		Valid bool   `json:"valid"`
		Field string `json:"field"`
	} `json:"json"`
	Expiry []struct {
		Input string `json:"input"`
		Valid bool   `json:"valid"`
		Kind  string `json:"kind"`
		Value string `json:"value"`
	} `json:"expiry"`
}

func TestSharedVectors(t *testing.T) {
	b, err := os.ReadFile("testdata/signal_v1_vectors.json")
	if err != nil {
		t.Fatalf("shared vectors are part of this package and must be present: %v", err)
	}
	var vf sharedVectors
	if err := json.Unmarshal(b, &vf); err != nil {
		t.Fatalf("read shared vectors: %v", err)
	}
	if vf.Version != Version {
		t.Fatalf("shared vectors are for v%d, this reader accepts v%d", vf.Version, Version)
	}

	for _, tc := range vf.Records {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			in, err := Parse(tc.Record)
			if !tc.Valid {
				if err == nil {
					t.Fatalf("Parse accepted an invalid record: %s", tc.Record)
				}
				if tc.Field != "" && !strings.Contains(err.Error(), tc.Field) {
					t.Fatalf("Parse error = %q, want it to mention %q", err, tc.Field)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			got, err := in.Canonical()
			if err != nil {
				t.Fatalf("Canonical: %v", err)
			}
			if string(got) != tc.Canonical {
				t.Fatalf("Canonical:\n got: %s\nwant: %s", got, tc.Canonical)
			}
			for _, lc := range tc.Live {
				err := in.Live(time.Unix(lc.Now, 0), lc.Height)
				if lc.Valid {
					if err != nil {
						t.Fatalf("Live(now=%d, height=%d): %v", lc.Now, lc.Height, err)
					}
					continue
				}
				if err == nil {
					t.Fatalf("Live accepted a dead record at now=%d, height=%d", lc.Now, lc.Height)
				}
				if lc.Field != "" && !strings.Contains(err.Error(), lc.Field) {
					t.Fatalf("Live error = %q, want it to mention %q", err, lc.Field)
				}
			}
		})
	}

	for _, tc := range vf.JSON {
		t.Run("json/"+tc.Name, func(t *testing.T) {
			_, err := Parse([]byte(tc.Raw))
			if tc.Valid {
				if err != nil {
					t.Fatalf("Parse: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Parse accepted invalid JSON")
			}
			if tc.Field != "" && !strings.Contains(err.Error(), tc.Field) {
				t.Fatalf("Parse error = %q, want it to mention %q", err, tc.Field)
			}
		})
	}

	for _, tc := range vf.Expiry {
		t.Run("expiry/"+tc.Input, func(t *testing.T) {
			got, err := ParseExpiry(tc.Input)
			if !tc.Valid {
				if err == nil {
					t.Fatalf("ParseExpiry accepted %q", tc.Input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseExpiry: %v", err)
			}
			if string(got.Kind) != tc.Kind || strconv.FormatUint(got.Value, 10) != tc.Value {
				t.Fatalf("got %s:%d, want %s:%s", got.Kind, got.Value, tc.Kind, tc.Value)
			}
		})
	}
}
