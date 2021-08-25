package testutil

import (
	"errors"
	"testing"
)

func TestEqualErrorMessage(t *testing.T) {
	testError := errors.New("test-error")
	tt := []struct {
		desc      string
		got       error
		want      error
		wantEqual bool
	}{
		{
			desc:      "both nil",
			got:       nil,
			want:      nil,
			wantEqual: true,
		},
		{
			desc:      "same",
			got:       testError,
			want:      testError,
			wantEqual: true,
		},
		{
			desc:      "same text",
			got:       errors.New("test-text"),
			want:      errors.New("test-text"),
			wantEqual: true,
		},
		{
			desc:      "got nil",
			got:       nil,
			want:      testError,
			wantEqual: false,
		},
		{
			desc:      "want nil",
			got:       testError,
			want:      nil,
			wantEqual: false,
		},
		{
			desc:      "different text",
			got:       errors.New("text1"),
			want:      errors.New("text2"),
			wantEqual: false,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			equal := EqualErrorMessage(tc.got, tc.want)
			if equal != tc.wantEqual {
				t.Errorf("got equal %v, want %v", equal, tc.wantEqual)
			}
		})
	}
}
