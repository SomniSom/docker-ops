package cli

import (
	"reflect"
	"testing"
)

func TestLogsComposeExtras(t *testing.T) {
	tests := []struct {
		tail, tty  bool
		wantExtra  []string
		wantRunTTY bool
	}{
		{tail: true, tty: false, wantExtra: []string{"--tail", "200"}, wantRunTTY: false},
		{tail: true, tty: true, wantExtra: []string{"--tail", "200"}, wantRunTTY: false},
		{tail: false, tty: true, wantExtra: []string{"-f"}, wantRunTTY: true},
		{tail: false, tty: false, wantExtra: []string{"--tail", "200"}, wantRunTTY: false},
	}
	for _, tt := range tests {
		ex, tty := logsComposeExtras(tt.tail, tt.tty)
		if !reflect.DeepEqual(ex, tt.wantExtra) || tty != tt.wantRunTTY {
			t.Errorf("logsComposeExtras(tail=%v, tty=%v) = (%v, %v), want (%v, %v)",
				tt.tail, tt.tty, ex, tty, tt.wantExtra, tt.wantRunTTY)
		}
	}
}
