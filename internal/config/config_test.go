package config

import "testing"

func TestValidateDatabaseConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  DatabaseConfig
		wantErr bool
	}{
		{
			name: "sqlite",
			config: DatabaseConfig{
				Driver: "sqlite",
				SQLite: &SQLiteConfig{Path: "phoenix.db"},
			},
		},
		{
			name: "sqlite3",
			config: DatabaseConfig{
				Driver: "sqlite3",
				SQLite: &SQLiteConfig{Path: "phoenix.db"},
			},
		},
		{
			name: "sqlite missing path",
			config: DatabaseConfig{
				Driver: "sqlite",
				SQLite: &SQLiteConfig{},
			},
			wantErr: true,
		},
		{
			name: "postgres",
			config: DatabaseConfig{
				Driver: "postgres",
				Postgres: &PostgresConfig{
					Host:   "localhost",
					Port:   5432,
					DBName: "phoenix",
				},
			},
		},
		{
			name: "postgres with auth",
			config: DatabaseConfig{
				Driver: "postgres",
				Postgres: &PostgresConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "password",
					DBName:   "phoenix",
				},
			},
		},
		{
			name: "postgres missing password with user",
			config: DatabaseConfig{
				Driver: "postgres",
				Postgres: &PostgresConfig{
					Host:   "localhost",
					Port:   5432,
					User:   "postgres",
					DBName: "phoenix",
				},
			},
			wantErr: true,
		},
		{
			name: "postgres invalid port",
			config: DatabaseConfig{
				Driver: "postgres",
				Postgres: &PostgresConfig{
					Host:   "localhost",
					DBName: "phoenix",
				},
			},
			wantErr: true,
		},
		{
			name: "postgres port out of range",
			config: DatabaseConfig{
				Driver: "postgres",
				Postgres: &PostgresConfig{
					Host:   "localhost",
					Port:   65536,
					DBName: "phoenix",
				},
			},
			wantErr: true,
		},
		{
			name: "postgres invalid host",
			config: DatabaseConfig{
				Driver: "postgres",
				Postgres: &PostgresConfig{
					Host:   "not a host",
					Port:   5432,
					DBName: "phoenix",
				},
			},
			wantErr: true,
		},
		{
			name: "postgres invalid user",
			config: DatabaseConfig{
				Driver: "postgres",
				Postgres: &PostgresConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "post\ngres",
					Password: "password",
					DBName:   "phoenix",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate.Struct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
