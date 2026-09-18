package utils

import "testing"

func TestParsePgDumpMajor(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    int
		wantErr bool
	}{
		{
			name:   "debian bookworm client",
			output: "pg_dump (PostgreSQL) 15.18 (Debian 15.18-0+deb12u1)\n",
			want:   15,
		},
		{
			name:   "pgdg client 18",
			output: "pg_dump (PostgreSQL) 18.0\n",
			want:   18,
		},
		{
			name:   "pgdg client with extra text",
			output: "pg_dump (PostgreSQL) 17.11 (Debian 17.11-1.pgdg12+1)\n",
			want:   17,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
		{
			name:    "unrelated output",
			output:  "not a pg_dump version string",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePgDumpMajor(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got major %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got major %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPgMajorFromServerVersionNum(t *testing.T) {
	tests := []struct {
		versionNum int
		want       int
	}{
		{versionNum: 170011, want: 17},
		{versionNum: 180000, want: 18},
		{versionNum: 150018, want: 15},
		{versionNum: 100001, want: 10},
	}

	for _, tt := range tests {
		got := pgMajorFromServerVersionNum(tt.versionNum)
		if got != tt.want {
			t.Errorf("pgMajorFromServerVersionNum(%d) = %d, want %d", tt.versionNum, got, tt.want)
		}
	}
}

func TestPgDumpSupportsServer(t *testing.T) {
	tests := []struct {
		dumpMajor   int
		serverMajor int
		want        bool
	}{
		{dumpMajor: 18, serverMajor: 17, want: true},
		{dumpMajor: 18, serverMajor: 18, want: true},
		{dumpMajor: 15, serverMajor: 17, want: false},
		{dumpMajor: 15, serverMajor: 15, want: true},
	}

	for _, tt := range tests {
		got := pgDumpSupportsServer(tt.dumpMajor, tt.serverMajor)
		if got != tt.want {
			t.Errorf("pgDumpSupportsServer(%d, %d) = %v, want %v", tt.dumpMajor, tt.serverMajor, got, tt.want)
		}
	}
}
