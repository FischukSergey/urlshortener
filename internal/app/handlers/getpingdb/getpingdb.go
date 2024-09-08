package getpingdb

import (
	"log/slog"
	"net/http"
)

type DBPinger interface { //имплементирует интерфейс с методом поиска по хранилищу
	GetPingDB() error
}

// GetPingDB хендлер проверки коннекта с базой данных
func GetPingDB(log *slog.Logger, storage DBPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: GetPingDB")

		err := storage.GetPingDB()
		if err != nil {
			http.Error(w, "database not connected", http.StatusInternalServerError)
			log.Error(err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
			log.Info("Ping database successful")
		}
	}
}
