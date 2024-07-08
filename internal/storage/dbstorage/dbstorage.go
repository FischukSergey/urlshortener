package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrURLExists = errors.New("url exists")

type Storage struct {
	db *sql.DB
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

// NewDB() создаем объект базы данных postgres
func NewDB(dbConfig config.DBConfig) (*Storage, error) {

	// dbconn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
	// 	dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)

	db, err := sql.Open("pgx", config.FlagDatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("%w, unable to create connection db:%s", err, dbConfig.Database)
	}
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS urlshort
	  (id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    alias varchar NOT NULL UNIQUE,
    url varchar NOT NULL);
	`)
	if err != nil {
		return nil, fmt.Errorf("%w, unable to prepare query", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%w, unable to execute query", err)
	}
	_, err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS url_idx ON urlshort (url);") //создаем уникальный индекс по оригинальному url
	if err != nil {
		return nil, fmt.Errorf("%w, unable to create index", err)
	}

	return &Storage{db: db}, nil
}

// GetPingDB() метод проверки соединения с базой
func (s *Storage) GetPingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("соединение с базой отсутствует %w", err)
	}
	return nil
}

// GetStorageURL() метод получения URL по алиасу
func (s *Storage) GetStorageURL(ctx context.Context, alias string) (string, bool) {
	const where = "dbstorage.GetStorageURL"
	log = log.With(slog.String("method from", where))

	stmt, err := s.db.PrepareContext(ctx, "SELECT url FROM urlshort WHERE alias=$1")
	if err != nil {
		log.Error("unable to prepare query")
		return "", false
	}
	defer stmt.Close()

	var resURL string

	err = stmt.QueryRowContext(ctx, alias).Scan(&resURL)
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
func (s *Storage) SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error {
	const op = "dbstorage.SaveStorageURL"
	//начинаем транзакцию записи в БД
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: не удалось начать транзакцию записи в базу %w", op, err)
	}
	defer tx.Rollback()

	//готовим запрос на вставку
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urlshort (alias,url) VALUES($1,$2);") 
	if err != nil {
		return fmt.Errorf("%s: не удалось подготовить транзакцию записи в базу %w", op, err)
	}
	defer stmt.Close()

	//пишем слайс urlов в базу данных
	for _, ss := range saveURL {

		_, err := stmt.ExecContext(ctx, ss.ShortURL, ss.OriginalURL)
		//обработка ошибки вставки url
		if err != nil {
			//если url неуникальный
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				var shorturl string
				err = s.db.QueryRowContext(ctx, "SELECT alias FROM urlshort WHERE url=$1", ss.OriginalURL).Scan(&shorturl)
				if errors.Is(err, sql.ErrNoRows) {
					log.Error("url not found")
					return fmt.Errorf("%s: %w", op, ErrURLExists)
				}
				return fmt.Errorf("%s: %w", shorturl, ErrURLExists) 
			}
			return fmt.Errorf("%s: не удалось выполнить транзакцию записи в базу %w", op, err)
		}
	}
	return tx.Commit()
}

func (s *Storage) Close() {
	s.db.Close()
}
