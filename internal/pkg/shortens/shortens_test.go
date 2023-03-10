package shortens

import "testing"

func TestShort(t *testing.T) {
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
			want: "3f",
		},
		{
			name: "4",
			args: 90,
			want: "2i",
		},
		{
			name: "5",
			args: 10000,
			want: "7ps",
		},
		{
			name: "6",
			args: 1000000,
			want: "lfls",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Short(tt.args); got != tt.want {
				t.Errorf("Short() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOriginal(t *testing.T) {
	tests := []struct {
		name    string
		want    int
		args    string
		wantErr bool
	}{
		{
			name:    "1",
			want:    0,
			args:    "0",
			wantErr: false,
		},
		{
			name:    "2",
			want:    11,
			args:    "b",
			wantErr: false,
		},
		{
			name:    "3",
			want:    123,
			args:    "3f",
			wantErr: false,
		},
		{
			name:    "4",
			want:    90,
			args:    "2i",
			wantErr: false,
		},
		{
			name:    "5",
			want:    10000,
			args:    "7ps",
			wantErr: false,
		},
		{
			name:    "6",
			want:    1000000,
			args:    "lfls",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Original(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Original() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Original() got = %v, want %v", got, tt.want)
			}
		})
	}
}
