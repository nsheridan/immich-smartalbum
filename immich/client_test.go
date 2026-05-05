package immich

import (
	"errors"
	"testing"
)

func TestAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "no error",
			err:  nil,
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("broken"),
			want: false,
		},
		{
			name: "non-auth api error",
			err:  &APIError{StatusCode: 404},
			want: false,
		},
		{
			name: "auth error",
			err:  &APIError{StatusCode: 401},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AuthError(test.err)
			if result != test.want {
				t.Errorf("wanted %v, got %v", test.want, result)
			}
		})
	}
}
