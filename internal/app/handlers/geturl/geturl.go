package geturl

import (
	"log/slog"
	"net/http"

	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurl"
	"github.com/go-chi/chi"
)

func GetURL(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: GetURL")

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")
			http.Error(w, "alias is empty", http.StatusBadRequest)
			return
		}

		url, ok := saveurl.URLStorage[alias]
		if !ok {
			http.Error(w, "alias not found", http.StatusBadRequest)
			log.Error("alias not found", slog.String("alias: ", alias))
			return
		}

		/*
		   resp, err := json.Marshal(task)
		   if err != nil {
		       http.Error(w, err.Error(), http.StatusBadRequest)
		       return
		   }
		*/

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)

		// w.Write([]byte(url))
		log.Info("Request alias  successful: ", slog.String("alias:", alias), slog.String("url:", url))
		//http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	}
}
