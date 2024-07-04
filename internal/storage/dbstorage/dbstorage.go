package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/FischukSergey/urlshortener.git/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

// NewDB() создаем объект базы данных postgres
func NewDB(dbConfig config.DBConfig) (*Storage, error) {

	dbconn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)

	db, err := sql.Open("pgx", dbconn)
	if err != nil {
		return nil, fmt.Errorf("%s, unable to create connection db:%s", err, dbConfig.Database)
	}
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS urlshort
	  (id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    alias varchar NOT NULL UNIQUE,
    url varchar NOT NULL);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s, unable to prepare query", err)
	}
	defer stmt.Close()

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

// GetStorageURL() метод получения URL по алиасу
func (s *Storage) GetStorageURL(ctx context.Context, alias string) (string, bool) {
	const where = "dbstorage.GetStorageURL"
	log = log.With(slog.String("method from", where))

	stmt, err := s.db.Prepare("SELECT url FROM urlshort WHERE alias=$1")
	if err != nil {
		log.Error("unable to prepare query")
		return "", false
	}
	defer stmt.Close()

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		log.Error("row not found")
		return "", false
	}
	if err != nil {
		log.Error("unable to execute query")
		return "", false
	}
	return resURL, true
}

// SaveStorage() метод сохранения alias в BD
func (s *Storage) SaveStorageURL(ctx context.Context, alias, URL string) error {
	const op = "dbstorage.SaveStorageURL"
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s не удалось начать транзакцию записи в базу %s", op, err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urlshort (alias,url) VALUES($1,$2)")
	if err != nil {
		return fmt.Errorf("%s не удалось подготовить транзакцию записи в базу %s", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, alias, URL)
	if err != nil {
		return fmt.Errorf("%s не удалось выполнить транзакцию записи в базу %s", op, err)
	}

	return tx.Commit()
}

func (s *Storage) Close() {
	s.db.Close()
}
