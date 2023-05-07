package storage

import (
	"log"
	"strconv"
	"testing"

	"main/internal/app/config"
)

func TestAddAndGet(t *testing.T) {
	conf := config.Conf

	c, _, _, err := StartStorage(conf)
	if err != nil {
		log.Print(err)
	}

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
			gotAdd, err := c.Add(tt.url, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAdd != tt.want {
				t.Errorf("Add() got = %v, want %v", gotAdd, tt.want)
			}
			gotGet, _, err := c.Get(tt.want)
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
