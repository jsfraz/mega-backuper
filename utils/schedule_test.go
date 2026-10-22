package utils

import "testing"

func TestParseSchedule(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		wantSeconds bool
		wantErr     bool
	}{
		{name: "standard 5 fields", expression: "0 12 * * *"},
		{name: "6 fields with seconds", expression: "30 0 12 * * *", wantSeconds: true},
		{name: "descriptor daily", expression: "@daily"},
		{name: "descriptor every", expression: "@every 1h30m"},
		{name: "surrounding whitespace", expression: "  0 12 * * *  "},
		{name: "7 fields with year", expression: "0 0 12 * * * 2030", wantErr: true},
		{name: "4 fields", expression: "0 12 * *", wantErr: true},
		{name: "empty", expression: "", wantErr: true},
		{name: "garbage", expression: "not a cron", wantErr: true},
		{name: "out of range minute", expression: "99 12 * * *", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSeconds, err := ParseSchedule(tt.expression)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got seconds=%v", tt.expression, gotSeconds)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.expression, err)
			}
			if gotSeconds != tt.wantSeconds {
				t.Fatalf("ParseSchedule(%q) seconds = %v, want %v", tt.expression, gotSeconds, tt.wantSeconds)
			}
		})
	}
}
