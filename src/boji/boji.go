package boji

import (
	"fmt"
	"net/http"
	"golang.org/x/net/webdav"
)

type ServerSettings struct {
	Port int
	Root string
	AdminUsername string
	AdminPassword string
}

type Server struct {
	Settings ServerSettings
	wdav *webdav.Handler
}

func NewServer(settings ServerSettings) *Server {

	return &Server{
		Settings: settings,
		wdav: &webdav.Handler {
			FileSystem: webdav.Dir(settings.Root),
			LockSystem: webdav.NewMemLS(),
		},
	}
}

func (this *Server) Listen() error {

	path := fmt.Sprintf(":%d", this.Settings.Port)
	return http.ListenAndServe(path, this.authenticatedHandler())
}

func (this *Server) authenticatedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		username, password, ok := r.BasicAuth()
		if !ok || username != this.Settings.AdminUsername || password != this.Settings.AdminPassword {
			http.Error(w, "Not authorized", 401)
			return
		}

		this.wdav.ServeHTTP(w, r)
	})
}