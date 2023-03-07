package shortens

import "testing"

func TestShortens(t *testing.T) {
	tests := []struct {
		name string
		args int
		want string
	}{
		{
			name: "1",
			args: 0,
			want: "0",
		},
		{
			name: "2",
			args: 11,
			want: "b",
		},
		{
			name: "3",
			args: 123,
			want: "7b",
		},
		{
			name: "4",
			args: 90,
			want: "5a",
		},
		{
			name: "5",
			args: 10000,
			want: "2710",
		},
		{
			name: "6",
			args: 1000000,
			want: "f4240",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Shortens(tt.args); got != tt.want {
				t.Errorf("Shortens() = %v, want %v", got, tt.want)
			}
		})
	}
}
