package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/owncloud/cs3-wopi-server/pkg/internal/middleware"
)

func (app *demoApp) HTTPServer(ctx context.Context) error {
	// start a simple web server that will get requests from
	// app provider client, eg. ownCloud Web
	r := chi.NewRouter()

	r.Use(middleware.AccessLog(app.Logger))

	r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
	})

	r.Route("/wopi", func(r chi.Router) {

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			WopiInfoHandler(app, w, r)
		})

		r.Route("/files/{fileid}", func(r chi.Router) {

			r.Use(func(h http.Handler) http.Handler {
				// authentication and wopi context
				return WopiContextAuthMiddleware(app, h)
			},
			)

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				CheckFileInfo(app, w, r)
			})

			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				action := r.Header.Get("X-WOPI-Override")
				switch action {

				case "LOCK":
					Lock(app, w, r)
				case "GET_LOCK":
					GetLock(app, w, r)
				case "REFRESH_LOCK":
					RefreshLock(app, w, r)
				case "UNLOCK":
					UnLock(app, w, r)

				case "PUT_USER_INFO":
					// https://docs.microsoft.com/en-us/microsoft-365/cloud-storage-partner-program/rest/files/putuserinfo
					http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
				case "PUT_RELATIVE":
					// https://docs.microsoft.com/en-us/microsoft-365/cloud-storage-partner-program/rest/files/putrelativefile
					http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
				case "RENAME_FILE":
					// https://docs.microsoft.com/en-us/microsoft-365/cloud-storage-partner-program/rest/files/renamefile
					http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
				case "DELETE":
					// https://docs.microsoft.com/en-us/microsoft-365/cloud-storage-partner-program/rest/files/deletefile
					http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)

				default:
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			})

			r.Route("/contents", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					GetFile(app, w, r)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					action := r.Header.Get("X-WOPI-Override")
					switch action {

					case "PUT":
						PutFile(app, w, r)

					default:
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					}
				})
			})
		})
	})

	go func() {
		if err := http.ListenAndServe(app.Config.HTTP.BindAddr, r); err != nil {
			app.Logger.Error().Err(err).Msg("HTTP server error")
		}
	}()

	return nil
}
