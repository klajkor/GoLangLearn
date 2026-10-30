// Command server runs the GoLearn app on localhost.
package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	course "golearn/content"
	"golearn/internal/api"
	"golearn/internal/content"
	"golearn/internal/playground"
	"golearn/internal/store"
	"golearn/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	contentDir := flag.String("content", "", "read the curriculum from this directory instead of the copy built into the binary (for editing content)")
	openFlag := flag.Bool("open", false, "open the app in your default browser once it is running")
	dbPath := flag.String("db", "data/progress.db", "SQLite progress database")
	flag.Parse()

	var courseFS fs.FS = course.FS
	if *contentDir != "" {
		courseFS = os.DirFS(*contentDir)
	}
	cur, err := content.Load(courseFS, content.WithShuffledOptions())
	if err != nil {
		log.Fatalf("load content: %v", err)
	}
	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	if _, err := fs.Stat(web.Dist(), "index.html"); err == nil {
		if err := verifyFrontend(web.Dist()); err != nil {
			log.Fatalf("%v\nRebuild it with ./run.ps1 build (or: cd web && npm run build, then go build)", err)
		}
	}

	mux := http.NewServeMux()
	api.New(cur, st, playground.New()).Register(mux)
	mux.Handle("/", spaHandler(web.Dist()))

	// Listen before announcing anything so a busy port gives a clear message.
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("cannot listen on %s: %v\nIs GoLearn already running? Stop it, or pick another port with -addr 127.0.0.1:8090", *addr, err)
	}

	srv := &http.Server{Addr: *addr, Handler: logRequests(mux), ReadHeaderTimeout: 5 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	url := "http://" + *addr
	log.Printf("GoLearn: %d chapters loaded, open %s", len(cur.Chapters), url)
	if *openFlag {
		if err := openBrowser(url); err != nil {
			log.Printf("could not open a browser: %v", err)
		}
	}
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// spaHandler serves static files and falls back to index.html so that
// client-side routes such as /chapters/intro work on reload.
func spaHandler(fsys fs.FS) http.Handler {
	files := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(fsys, name); err != nil {
			if _, err := fs.Stat(fsys, "index.html"); err != nil {
				http.Error(w, "Frontend not built. Run: cd web && npm install && npm run build", http.StatusServiceUnavailable)
				return
			}
			http.ServeFileFS(w, r, fsys, "index.html")
			return
		}
		// Vite fingerprints everything in assets/, so those files never change.
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Microsecond))
		}
	})
}
