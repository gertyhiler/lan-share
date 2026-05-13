package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	fsadapter "github.com/gertyhiler/lan-share/internal/adapter/fs"
	httpadapter "github.com/gertyhiler/lan-share/internal/adapter/http"
	"github.com/gertyhiler/lan-share/internal/platform/netutil"
	chatuc "github.com/gertyhiler/lan-share/internal/usecase/chat"
	fileuc "github.com/gertyhiler/lan-share/internal/usecase/files"
	"github.com/gertyhiler/lan-share/internal/usecase/paste"
)

func main() {
	if err := run(); err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}

func run() error {
	host := flag.String("host", "0.0.0.0", "Bind host")
	port := flag.Int("port", 8000, "Bind port")
	root := flag.String("root", ".", "Directory for lan_share_* data folders (default: cwd)")
	flag.Parse()

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}
	uploads := filepath.Join(absRoot, "lan_share_uploads")
	shared := filepath.Join(absRoot, "lan_share_shared")
	pastes := filepath.Join(absRoot, "lan_share_pastes")
	chatDir := filepath.Join(absRoot, "lan_share_chat")

	pasteStore := &fsadapter.Pastes{Dir: pastes}
	uploadStore := &fsadapter.Uploads{Dir: uploads}
	sharedStore := &fsadapter.Shared{Dir: shared}
	chatStore := &fsadapter.Chat{Dir: chatDir}

	pasteSvc := paste.NewService(pasteStore)
	filesSvc := fileuc.NewService(uploadStore, sharedStore)
	chatSvc := chatuc.NewService(chatStore)

	paths := httpadapter.Paths{Uploads: uploads, Shared: shared}
	handler, err := httpadapter.NewHandler(pasteSvc, filesSvc, chatSvc, *port, paths, log.Default(), "lan-share/1.0")
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           withRequestLog(log.Default(), handler.Routes()),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       120 * time.Second,
		ErrorLog:          log.Default(),
	}

	logger := log.Default()
	logger.Printf("LAN Share started.")
	logger.Printf("- Local:   http://127.0.0.1:%d/", *port)
	if *host == "0.0.0.0" || *host == "" {
		for i, ip := range netutil.LANIPv4() {
			if i >= 5 {
				break
			}
			logger.Printf("- LAN:     http://%s:%d/", ip, *port)
		}
	} else {
		logger.Printf("- Listen:  http://%s:%d/", *host, *port)
	}
	logger.Printf("- Uploads: %s", uploads)
	logger.Printf("- Shared:  %s", shared)
	logger.Printf("- Pastes:  %s", pastes)
	logger.Printf("- Chat:    %s", chatDir)
	logger.Printf("Ctrl+C to stop.")

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sig:
		logger.Printf("Stopping...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	}
}

type logResponseWriter struct {
	http.ResponseWriter
	code int
}

func (w *logResponseWriter) WriteHeader(code int) {
	if w.code == 0 {
		w.code = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *logResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *logResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func withRequestLog(l *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &logResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)
		if lrw.code == 0 {
			lrw.code = http.StatusOK
		}
		msg := fmt.Sprintf("%s %s %d", r.Method, r.URL.RequestURI(), lrw.code)
		msg = strings.ReplaceAll(msg, "\n", "\\n")
		msg = strings.ReplaceAll(msg, "\r", "\\r")
		l.Printf("[%s] %s", r.RemoteAddr, msg)
	})
}
