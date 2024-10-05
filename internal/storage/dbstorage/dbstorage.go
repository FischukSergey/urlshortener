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

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/getuserallurl"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
)

// ErrURLExists ошибка, если url уже существует
var ErrURLExists = errors.New("url exists")

// Storage структура для работы с базой данных
type Storage struct {
	DB      *pgxpool.Pool
	DelChan chan config.DeletedRequest //канал для записи отложенных запросов на удаление
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

// NewDB создаем объект базы данных postgres
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

	instance := &Storage{
		DB:      db,
		DelChan: make(chan config.DeletedRequest, 1024), //устанавливаем каналу буфер
	}

	go instance.flushDeletes() //горутина фонового сохранения данных на удаление

	return instance, nil
}

// flushMessages постоянно отправляет несколько сообщений в хранилище с определённым интервалом
func (s *Storage) flushDeletes() {
	// будем отправлять сообщения, накопленные за последние 10 секунд
	ticker := time.NewTicker(10 * time.Second)

	var delmsges []config.DeletedRequest

	for {
		select {
		case msg := <-s.DelChan:
			// добавим сообщение в слайс для последующей отправки на удаление
			delmsges = append(delmsges, msg)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(delmsges) == 0 {
				continue
			}
			//отправим на удаление все пришедшие сообщения одновременно
			err := s.DeleteBatch(context.TODO(), delmsges...)
			if err != nil {
				log.Debug("cannot save messages", logger.Err(err))
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			delmsges = nil
		}
	}
}

// GetPingDB метод проверки соединения с базой
func (s *Storage) GetPingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.DB.Ping(ctx)
	if err != nil {
		return fmt.Errorf("соединение с базой отсутствует %w", err)
	}
	return nil
}

// GetStorageURL метод получения URL по алиасу
func (s *Storage) GetStorageURL(ctx context.Context, alias string) (string, bool) {
	const where = "dbstorage.GetStorageURL"
	log = log.With(slog.String("method from", where))

	query := "SELECT url, deletedflag FROM urlshort WHERE alias=$1;"

	var resURL string
	var resDeleted bool
	err := s.DB.QueryRow(ctx, query, alias).Scan(&resURL, &resDeleted)
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

// SaveStorage метод сохранения alias в BD
func (s *Storage) SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error {
	const op = "dbstorage.SaveStorageURL"

	//готовим запрос на вставку
	query := `INSERT INTO urlshort (alias,url,userid) VALUES($1,$2,$3);`

	//пишем слайс urlов в базу данных
	for _, ss := range saveURL {

		_, err := s.DB.Exec(ctx, query, ss.ShortURL, ss.OriginalURL, ss.UserID)
		//обработка ошибки вставки url
		if err != nil {
			//если url неуникальный
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				var shorturl string
				err = s.DB.QueryRow(ctx, "SELECT alias FROM urlshort WHERE url=$1", ss.OriginalURL).Scan(&shorturl)
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

// Close закрывает соединение с базой данных
func (s *Storage) Close() {
	s.DB.Close()
}

// GetAllUserURL осуществляет выборку всех записей, сделанных пользователем ID
// Принимает ID пользователя, возвращает слайс сокращенных и оригинальных URL
func (s *Storage) GetAllUserURL(ctx context.Context, userID int) ([]getuserallurl.AllURLUserID, error) {
	const op = "dbstorage.GetAllUserURL"
	log = log.With(slog.String("method from", op))

	var getUserURLs []getuserallurl.AllURLUserID

	query := `SELECT alias,url FROM urlshort WHERE userid=$1`

	result, err := s.DB.Query(ctx, query, userID)
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
	if err := result.Err(); err != nil {
		return nil, err
	}
	return getUserURLs, nil
}

// DeleteBatch метод удаления записей по списку сокращенных URl сделанных определенным пользователем
func (s *Storage) DeleteBatch(ctx context.Context, delmsges ...config.DeletedRequest) error {
	const op = "dbstorage.DeleteBatch"
	log = log.With(slog.String("method from", op))

	query := `UPDATE urlshort SET deletedflag=true WHERE alias=$1 AND userid=$2;`

	batch := &pgx.Batch{} //формируем пакет запросов

	for _, delmsg := range delmsges {
		batch.Queue(query, delmsg.ShortURL, delmsg.UserID)
	}
	br := s.DB.SendBatch(ctx, batch)

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

// GetStats метод получения статистики по количеству пользователей и сокращенных URL
func (s *Storage) GetStats(ctx context.Context) (config.Stats, error) {
	const op = "dbstorage.GetStats"
	log = log.With(slog.String("method from", op))

	query := `SELECT COUNT(DISTINCT userid) AS users, COUNT(*) AS urls FROM urlshort;`

	var stats config.Stats
	err := s.DB.QueryRow(ctx, query).Scan(&stats.Users, &stats.URLs)
	if err != nil {
		log.Error("unable to execute query")
		return config.Stats{}, fmt.Errorf("unable to execute query: %w", err)
	}

	return stats, nil
}
