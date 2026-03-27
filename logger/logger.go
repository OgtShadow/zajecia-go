package main

import (
	"encoding/csv"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type csvLogger struct {
	mu     sync.Mutex
	file   *os.File
	writer *csv.Writer
}

func newCSVLogger(path string) (*csvLogger, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	l := &csvLogger{
		file:   file,
		writer: csv.NewWriter(file),
	}

	if info.Size() == 0 {
		if err := l.writer.Write([]string{"timestamp", "ip_addr", "method", "path"}); err != nil {
			file.Close()
			return nil, err
		}
		l.writer.Flush()
		if err := l.writer.Error(); err != nil {
			file.Close()
			return nil, err
		}
	}

	return l, nil
}

func (l *csvLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer.Flush()
	if err := l.writer.Error(); err != nil {
		_ = l.file.Close()
		return err
	}
	return l.file.Close()
}

func (l *csvLogger) Log(r *http.Request) {
	record := []string{
		time.Now().Format(time.RFC3339),
		clientIP(r),
		r.Method,
		r.URL.Path,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.writer.Write(record); err != nil {
		log.Printf("failed to write log record: %v", err)
		return
	}

	l.writer.Flush()
	if err := l.writer.Error(); err != nil {
		log.Printf("failed to flush log record: %v", err)
	}
}

func clientIP(r *http.Request) string {
	forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

func loggerMiddleware(next http.Handler, reqLogger *csvLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqLogger.Log(r)
		next.ServeHTTP(w, r)
	})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Request received and logged\n"))
}

func main() {
	reqLogger, err := newCSVLogger("log.csv")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		if err := reqLogger.Close(); err != nil {
			log.Printf("failed to close logger: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: loggerMiddleware(mux, reqLogger),
	}

	log.Println("HTTP logger server listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
