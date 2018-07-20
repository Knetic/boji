package boji

import (
	"fmt"
	"os"
	"errors"
	"net/http"
	"golang.org/x/net/webdav"
)

type ServerSettings struct {
	TLSCertPath string
	TLSKeyPath string
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
			FileSystem: archivableFS(settings.Root),
			LockSystem: webdav.NewMemLS(),
			Logger: logStderr,
		},
	}
}

func (this *Server) Listen() error {

	path := fmt.Sprintf(":%d", this.Settings.Port)

	// if we're set up for TLS, serve https
	_, certErr := os.Stat(this.Settings.TLSCertPath)
	_, keyErr := os.Stat(this.Settings.TLSKeyPath)

	if certErr == nil && keyErr == nil {
		fmt.Printf("Listening on TLS %s\n", path)
		return http.ListenAndServeTLS(path, this.Settings.TLSCertPath, this.Settings.TLSKeyPath, this.authenticatedHandler())
	}

	// otherwise just plain http
	fmt.Printf("Listening on unencrypted http %s\n", path)
	return http.ListenAndServe(path, this.authenticatedHandler())
}

func (this *Server) authenticatedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// auth
		username, password, ok := r.BasicAuth()
		if !ok || username != this.Settings.AdminUsername || password != this.Settings.AdminPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Not authorized", 401)
			return
		}

		// check to see if this is a request to compress a directory
		areq, err := this.attemptArchiveRequest(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		if !areq {
			this.wdav.ServeHTTP(w, r)
		}
	})
}

/*
	Checks to see if this a request to archive/unarchive a directory.
	Returns true if this was a compression request, false otherwise.
	An error will only be returned if 
*/
func (this Server) attemptArchiveRequest(r *http.Request) (bool, error) {

	query := r.URL.Query()
	compressQuery, ok := query["compress"]
	if r.Method == "POST" && ok && len(compressQuery) > 0 {

		path := resolve(this.Settings.Root, r.URL.Path)
		stat, err := os.Stat(path)
		if err != nil {
			return true, errors.New("Unable to access directory")
		}
		if !stat.IsDir() {
			return true, errors.New("Not a directory")
		}

		compressed := compressQuery[0] == "true"
		if compressed {
			return true, archiveDir(path)
		} else {
			return true, unarchiveDir(path)
		}
	}

	return false, nil
}

func logStderr(request *http.Request, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to %s '%s': %v\n", request.Method, request.URL.Path, err)
	}
}