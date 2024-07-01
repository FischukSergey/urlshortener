package dbstorage

import (
	"database/sql"
	"fmt"

	"github.com/FischukSergey/urlshortener.git/config"
	 _ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}
/*
var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)
*/
func NewDB(dbConfig config.DBConfig) (*Storage, error) {

	dbconn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	
	db, err := sql.Open("pgx", dbconn)
	if err != nil {
		return nil, fmt.Errorf("%s, unable to creat connection db:%s", err, dbConfig.Database)
	}
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS urlshort
	  (id uuid primary key,
    alias varchar NOT NULL UNIQUE,
    url varchar NOT NULL);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s, unable to prepare query", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s, unable to execute query", err)
	}

	return &Storage{db: db}, nil
}

// GetPingDB() метод проверки соединения с базой
func (s *Storage) GetPingDB() error {
	err := s.db.Ping()
	if err != nil {
		return fmt.Errorf("соединение с базой отсутствует %s", err)
	}
	return nil
}

func (s *Storage) Close() {
	s.db.Close()
}