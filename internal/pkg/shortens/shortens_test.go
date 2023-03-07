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
			want: "0/000000",
		},
		{
			name: "2",
			args: 11,
			want: "0/00000b",
		},
		{
			name: "3",
			args: 123,
			want: "1/0003ff",
		},
		{
			name: "4",
			args: 4572,
			want: "50/0cffff",
		},
		{
			name: "5",
			args: 10000,
			want: "111/00000a",
		},
		{
			name: "6",
			args: 1000000,
			want: "11111/00000a",
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
