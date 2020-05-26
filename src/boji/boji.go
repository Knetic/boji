package boji

import (
	"fmt"
	"os"
	"strings"
	"errors"
	"context"
	"net/http"
	"time"
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

	InfluxURL string
	InfluxBucket string	
}

type Server struct {
	Settings ServerSettings
	wdav *webdav.Handler
	telemetry *telemetry

	stopTelemetry chan bool
}

func NewServer(settings ServerSettings) *Server {

	telemetry := newTelemetry(settings.InfluxURL, settings.InfluxBucket)

	return &Server{
		Settings: settings,
		wdav: &webdav.Handler {
			FileSystem: archivableFS{
				path: settings.Root,
				stats: &(telemetry.stats),
			},
			LockSystem: webdav.NewMemLS(),
			Logger: logStderr,
		},
		telemetry: telemetry,
	}
}

func (this *Server) Listen() error {

	path := fmt.Sprintf("%s:%d", this.Settings.Address, this.Settings.Port)

	if this.telemetry != nil {

		fmt.Printf("Will publish telemetry to influxdb at '%s', db '%s'\n", this.Settings.InfluxURL, this.Settings.InfluxBucket)
		this.stopTelemetry = make(chan bool)
		go this.runTelemetry()

		defer func(){
			this.stopTelemetry <- true
			close(this.stopTelemetry)
		}()
	}

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
			this.telemetry.stats.failedAuths++
			w.Header().Set("WWW-Authenticate", `Basic realm="boji"`)
			http.Error(w, err.Error(), 401)
			return
		} 

		if username != this.Settings.AdminUsername || password != this.Settings.AdminPassword {
			this.telemetry.stats.failedAuths++
			w.Header().Set("WWW-Authenticate", `Basic realm="boji"`)
			http.Error(w, "Not authorized", 401)
			return
		}

		// informational header so that clients can be assured encryption is actually working.
		if key != "" {
			// TODO: hardcoded to 256, but would benefit from actually knowing what the file was.
			w.Header().Set("X-Transparent-Encryption", encryptionProvidedHeaderValue)

			// pass the key internally, so that we can use it from archivableFS
			r = r.WithContext(context.WithValue(r.Context(), contextEncryptionKey, []byte(key)))
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
		recursiveStr, ok := query["recursive"]
		recursive := !ok || recursiveStr[0] == "true"

		if encrypted {
			return true, encryptDir(path, []byte(key), recursive)
		} else {
			return true, decryptDir(path, []byte(key), recursive)
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

func (this *Server) runTelemetry() {

	ticker := time.NewTicker(time.Minute)	
	for {
		select {
		
		case <-this.stopTelemetry:
			return

		case <-ticker.C:
			err := this.telemetry.publish()
			if err != nil {
				fmt.Printf("Telemetry publish failed: %v\n", err)
			}
		}
	}
}

func parseAuth(r *http.Request) (user string, password string, key string, _ error) {

	username, password, ok := r.BasicAuth()
	if !ok {
		return "", "", "", errors.New("Basic auth must be provided")
	}

	// check for symmetric encryption key
	idx := strings.IndexByte(password, ':')
	if idx > 0 {
		key = password[idx+1:]
		password = password[:idx]
	}
	
	return username, password, key, nil
}

func logStderr(request *http.Request, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to %s '%s': %v\n", request.Method, request.URL.Path, err)
	}
}