package common

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	testCases := []struct {
		in   string
		want string
		unit string
	}{
		{
			in:   "1m10s",
			want: "70.00",
			unit: "s",
		},
		{
			in:   "3s720ms",
			want: "3720",
			unit: "ms",
		},
		{
			in:   "720ms",
			want: "0.72",
			unit: "s",
		},
	}
	for _, tc := range testCases {
		d, err := time.ParseDuration(tc.in)
		if err != nil {
			t.Fatalf("error parsing %s: %v", tc.in, err)
		}
		if got := FormatDuration(d, tc.unit); got != tc.want {
			t.Errorf("want %s, got %s for %s", tc.want, got, tc.in)
		}
	}
}
