package boji

import (
	"fmt"
	"os"
	"strings"
	"errors"
	"net/http"
	"golang.org/x/net/webdav"
)

type ServerSettings struct {
	TLSCertPath string
	TLSKeyPath string
	Address string
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

	path := fmt.Sprintf("%s:%d", this.Settings.Address, this.Settings.Port)

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
		username, password, key, err := parseAuth(r)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		// informational header so that clients can be assured encryption is actually working.
		if key != "" {
			// TODO: hardcoded to 256, but would benefit from actually knowing what the file was.
			w.Header().Set("X-Transparent-Encryption", "aes-256")
		}

		if username != this.Settings.AdminUsername || password != this.Settings.AdminPassword {
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

		ereq, err := this.attemptEncryptionRequest(r, key)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		if !areq && !ereq {
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

		path, err := this.checkDir(r.URL.Path)
		if err != nil {
			return true, err
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

func (this Server) attemptEncryptionRequest(r *http.Request, key string) (bool, error) {

	query := r.URL.Query()
	encryptQuery, ok := query["encrypt"]
	if r.Method == "POST" && ok && len(encryptQuery) > 0 {

		if key == "" {
			return true, errors.New("Cannot perform encryption without a key specified - provide basic auth in the format `Basic base64(user:password:key)`")
		}

		path, err := this.checkDir(r.URL.Path)
		if err != nil {
			return true, err
		}

		encrypted := encryptQuery[0] == "true"
		if encrypted {
			return true, encryptDir(path, []byte(key))
		} else {
			return true, decryptDir(path, []byte(key))
		}
	}

	return false, nil
}

/*
	Resolves the on-disk path to the given [urlPath], and returns whether or not it's an accessible directory.
*/
func (this Server) checkDir(urlPath string) (string, error) {

	path := resolve(this.Settings.Root, urlPath)
	stat, err := os.Stat(path)
	if err != nil {
		return path, errors.New("Unable to access directory")
	}
	if !stat.IsDir() {
		return path, errors.New("Not a directory")
	}

	return path, nil
}

func parseAuth(r *http.Request) (user string, password string, key string, _ error) {

	username, password, ok := r.BasicAuth()

	if !ok {
		return "", "", "", errors.New("Basic auth must be provided")
	}

	// check for symmetric encryption key
	splits := strings.Split(password, ":")
	if len(splits) > 2 {
		return "", "", "", errors.New("Neither password nor encryption key can contain colons")
	}

	password = splits[0]

	if len(splits) == 2 {
		key = splits[1]
	}

	return username, password, key, nil
}

func logStderr(request *http.Request, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to %s '%s': %v\n", request.Method, request.URL.Path, err)
	}
}