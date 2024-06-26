package geturl

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
)

type URLGetter interface{  //имплементирует интерфейс с методом поиска по хранилищу
	GetStorageURL(alias string) (string, bool)
}


//GetURL хендлер запроса (GET{ID}) полного URL по его алиасу
func GetURL(log *slog.Logger, storage URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: GetURL")

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")
			http.Error(w, "alias is empty", http.StatusBadRequest)
			return
		}

		url, ok := storage.GetStorageURL(alias)
		if !ok {
			http.Error(w, "alias not found", http.StatusBadRequest)
			log.Error("alias not found", slog.String("alias:", alias))
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)

		log.Info("Request alias  successful:", slog.String("alias:", alias), slog.String("url:", url))
		//http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	}
}
