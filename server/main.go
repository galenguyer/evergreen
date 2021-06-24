package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/galenguyer/evergreen/controllers"
	"github.com/galenguyer/evergreen/utils"
	"github.com/gorilla/mux"
)

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func main() {
	// set up config flags
	var webRoot string
	var maxSize string
	var maxLifetime int
	flag.StringVar(&webRoot, "webroot", "../frontend/build", "the root directory of the frontend. defaults to ../frontend/build/")
	flag.StringVar(&maxSize, "size", "8MB", "the max size of files to allow")
	flag.IntVar(&maxLifetime, "lifetime", 7*24*60*60, "the max lifetime of a file in seconds")
	flag.Parse()

	// parse the max file size
	var s datasize.ByteSize
	if err := s.UnmarshalText([]byte(maxSize)); err != nil {
		log.Fatal(err)
	}

	// set up router
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	uploadControler := &controllers.UploadController{
		MaxFileSize: int(s.Bytes()),
		MaxLifetime: maxLifetime,
	}
	router.HandleFunc("/api/v1/upload", uploadControler.HandleUpload)
	router.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir("./uploads/"))))

	// set up spa handler
	spa := spaHandler{staticPath: webRoot, indexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)

	// set up server
	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// print info and start server
	log.Println("[main] listening on", srv.Addr)
	log.Println("[main] max upload size:", datasize.ByteSize(uploadControler.MaxFileSize).String())
	log.Println("[main] max lifetime:", time.Duration(uploadControler.MaxLifetime)*time.Second)

	log.Println("[main] starting cleanup loop")
	go func() {
		for {
			utils.Cleanup()
			time.Sleep(1 * time.Minute)
		}
	}()

	log.Println("[main] serving static files from", webRoot)
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
