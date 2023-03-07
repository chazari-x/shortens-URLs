package storage

import "testing"

func TestAddAndGet(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "1",
			url:     "https://github.com/chazari-x?tab=overview&from=2023-03-01&to=2023-03-07",
			want:    "0/000000",
			wantErr: false,
		},
		{
			name:    "2",
			url:     "https://pkg.go.dev/net/http@go1.17.2",
			want:    "0/000001",
			wantErr: false,
		},
		{
			name:    "3",
			url:     "https://github.com/golang-standards/project-layout/blob/master/README_ru.md",
			want:    "0/000002",
			wantErr: false,
		},
		{
			name:    "4",
			url:     "https://account.jetbrains.com/licenses",
			want:    "0/000003",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdd, err := Add(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAdd != tt.want {
				t.Errorf("Add() got = %v, want %v", gotAdd, tt.want)
			}
			gotGet, err := Get(tt.want)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotGet != tt.url {
				t.Errorf("Get() got = %v, want %v", gotGet, tt.url)
			}
		})
	}
}
