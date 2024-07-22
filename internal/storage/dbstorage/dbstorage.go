package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/getuserallurl"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrURLExists = errors.New("url exists")

type Storage struct {
	db *pgxpool.Pool
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

// NewDB() создаем объект базы данных postgres
func NewDB(dbConfig *pgconn.Config) (*Storage, error) {

	dbconn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		dbConfig.User, dbConfig.Password, dbConfig.Host, strconv.Itoa(int(dbConfig.Port)), dbConfig.Database)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := pgxpool.New(ctx, dbconn) //config.FlagDatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("%w, unable to create connection db:%s", err, dbConfig.Database)
	}
	query := `
	CREATE TABLE IF NOT EXISTS urlshort
	  (id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    alias VARCHAR NOT NULL UNIQUE,
    url VARCHAR NOT NULL,
		userid INTEGER DEFAULT 0,
		deletedflag BOOLEAN DEFAULT FALSE);
	`
	_, err = db.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w, unable to execute query", err)
	}
	_, err = db.Exec(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS url_idx ON urlshort (url);") //создаем уникальный индекс по оригинальному url
	if err != nil {
		return nil, fmt.Errorf("%w, unable to create index", err)
	}

	return &Storage{db: db}, nil
}

// GetPingDB() метод проверки соединения с базой
func (s *Storage) GetPingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("соединение с базой отсутствует %w", err)
	}
	return nil
}

// GetStorageURL() метод получения URL по алиасу
func (s *Storage) GetStorageURL(ctx context.Context, alias string) (string, bool) {
	const where = "dbstorage.GetStorageURL"
	log = log.With(slog.String("method from", where))

	query := "SELECT url, deletedflag FROM urlshort WHERE alias=$1;"

	var resURL string
	var resDeleted bool
	err := s.db.QueryRow(ctx, query, alias).Scan(&resURL, &resDeleted)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Error("row not found")
		return "", false
	}
	if err != nil {
		log.Error("unable to execute query")
		return "", false
	}
	if resDeleted { //если алиас есть, но помечен на удаление
		return resURL, false
	}

	return resURL, true
}

// SaveStorage() метод сохранения alias в BD
func (s *Storage) SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error {
	const op = "dbstorage.SaveStorageURL"

	id := ctx.Value(auth.CtxKeyUser).(int)

	//готовим запрос на вставку
	query := `INSERT INTO urlshort (alias,url,userid) VALUES($1,$2,$3);`

	//пишем слайс urlов в базу данных
	for _, ss := range saveURL {

		_, err := s.db.Exec(ctx, query, ss.ShortURL, ss.OriginalURL, id)
		//обработка ошибки вставки url
		if err != nil {
			//если url неуникальный
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				var shorturl string
				err = s.db.QueryRow(ctx, "SELECT alias FROM urlshort WHERE url=$1", ss.OriginalURL).Scan(&shorturl)
				if errors.Is(err, sql.ErrNoRows) {
					log.Error("url not found")
					return fmt.Errorf("%s: %w", op, ErrURLExists)
				}
				return fmt.Errorf("%s: %w", shorturl, ErrURLExists)
			}
			return fmt.Errorf("%s: не удалось выполнить транзакцию записи в базу %w", op, err)
		}
	}
	return nil
}

func (s *Storage) Close() {
	s.db.Close()
}

// GetAllUserURL осуществляет выборку всех записей, сделанных пользователем ID
// Принимает ID пользователя, возвращает слайс сокращенных и оригинальных URL
func (s *Storage) GetAllUserURL(ctx context.Context, userID int) ([]getuserallurl.AllURLUserID, error) {
	const op = "dbstorage.GetAllUserURL"
	log = log.With(slog.String("method from", op))

	var getUserURLs []getuserallurl.AllURLUserID

	query := `SELECT alias,url FROM urlshort WHERE userid=$1`

	result, err := s.db.Query(ctx, query, userID)
	if err != nil {
		log.Error("unable to execute query")
		return getUserURLs, fmt.Errorf("unable to execute query: %w", err)
	}
	if result.Err() != nil {
		log.Error("unable to execute query")
		return getUserURLs, fmt.Errorf("unable to execute query: %w", err)

	}
	defer result.Close()

	for result.Next() {
		var res getuserallurl.AllURLUserID
		err = result.Scan(&res.ShortURL, &res.OriginalURL)
		if err != nil {
			log.Error("unable to read row of query")
			return getUserURLs, fmt.Errorf("unable to read row of query: %w", err)
		}
		getUserURLs = append(getUserURLs, res)
	}

	return getUserURLs, nil
}

// DeleteBatch метод удаления записей по списку сокращенных URl сделанных определенным пользователем
func (s *Storage) DeleteBatch(ctx context.Context, aliases []string) error {
	const op = "dbstorage.DeleteBatch"
	log = log.With(slog.String("method from", op))

	id := ctx.Value(auth.CtxKeyUser).(int) //получаем id пользователя

	query := `UPDATE urlshort SET deletedflag=true WHERE alias=$1 AND userid=$2;`

	batch := &pgx.Batch{} //формируем пакет запросов
	for _, alias := range aliases {
		batch.Queue(query, alias, id)
	}
	br := s.db.SendBatch(ctx, batch)

	_, err := br.Exec()
	if err != nil {
		log.Error("unable to execute update batch of query")
		return fmt.Errorf("unable to execute update batch of query: %w", err)
	}

	err = br.Close() //в этот момент происходит обновление
	if err != nil {
		log.Error("unable to close  batch of query")
		return fmt.Errorf("unable to close  batch of query: %w", err)
	}

	return nil
}
