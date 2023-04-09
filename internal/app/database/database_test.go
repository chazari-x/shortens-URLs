package database

import (
	"fmt"
	"testing"

	"main/internal/app/config"
)

type DBConfig struct {
	Host string `yaml:"host"` // Хост
	Port string `yaml:"port"` // Порт
	User string `yaml:"user"` // Пользователь
	Pass string `yaml:"pass"` // Пароль
	Name string `yaml:"name"` // Название
}

func TestStartDB(t *testing.T) {
	conf, _ := config.ParseConfig()

	var db = DBConfig{
		Host: "localhost",
		Port: "32768",
		User: "postgres",
		Pass: "postgrespw",
		Name: "postgres",
	}

	conf.DataBaseDSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Pass, db.Name)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "one",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := StartDB(conf); (err != nil) != tt.wantErr {
				t.Errorf("StartDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
