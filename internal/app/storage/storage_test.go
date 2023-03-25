package storage

import (
	"strconv"
	"testing"
)

func TestAddAndGet(t *testing.T) {
	for i := 0; i < 25; i++ {
		tt := struct {
			name    string
			url     string
			want    string
			wantErr bool
		}{
			name:    strconv.Itoa(i),
			url:     "https://github.com/chazari-x?tab=overview&from=2023-03-01&to=" + strconv.Itoa(i),
			want:    strconv.FormatInt(int64(i), 36),
			wantErr: false,
		}
		t.Run(tt.name, func(t *testing.T) {
			gotAdd, err := Add(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAdd != tt.want {
				t.Errorf("Add() got = %v, want %v", gotAdd, tt.want)
			}
			gotGet, err := GetOriginal(tt.want)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOriginal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotGet != tt.url {
				t.Errorf("GetOriginal() got = %v, want %v", gotGet, tt.url)
			}
		})
	}
}
