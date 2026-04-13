package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type server struct {
	db *sql.DB
}

type author struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Country string    `json:"country"`
	Created time.Time `json:"created_at"`
}

type track struct {
	ID          int64     `json:"id"`
	AuthorID    int64     `json:"author_id"`
	AuthorName  string    `json:"author_name,omitempty"`
	Title       string    `json:"title"`
	Genre       string    `json:"genre"`
	ReleaseYear int       `json:"release_year"`
	Created     time.Time `json:"created_at"`
}

type createAuthorRequest struct {
	Name    string `json:"name"`
	Country string `json:"country"`
}

type createTrackRequest struct {
	AuthorID    int64  `json:"author_id"`
	Title       string `json:"title"`
	Genre       string `json:"genre"`
	ReleaseYear int    `json:"release_year"`
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("cannot get current directory: %v", err)
	}

	dbPath := filepath.Join(cwd, "music.db")
	db, err := openDatabase(dbPath)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

	s := &server{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleAPIInfo)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/autorzy", s.handleAuthorsCollection)
	mux.HandleFunc("/autorzy/", s.handleAuthorResource)
	mux.HandleFunc("/utwory", s.handleTracksCollection)
	mux.HandleFunc("/utwory/", s.handleTrackResource)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("CRUD API listening on %s", addr)
	if err := http.ListenAndServe(addr, withJSONContentType(mux)); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func openDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	queries := []string{
		"PRAGMA foreign_keys = ON;",
		`CREATE TABLE IF NOT EXISTS authors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			country TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS tracks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			author_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			genre TEXT NOT NULL DEFAULT '',
			release_year INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE CASCADE
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}

func withJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"name":        "CRUD Music API",
		"description": "Relacyjna baza danych: autorzy i utwory muzyczne",
		"commands": []map[string]string{
			{"method": "GET", "path": "/", "description": "Lista komend i możliwości API"},
			{"method": "GET", "path": "/health", "description": "Status serwera"},
			{"method": "GET", "path": "/autorzy", "description": "Lista autorów"},
			{"method": "POST", "path": "/autorzy", "description": "Dodaj autora"},
			{"method": "GET", "path": "/autorzy/{id}", "description": "Pobierz autora po ID"},
			{"method": "PUT", "path": "/autorzy/{id}", "description": "Zaktualizuj autora"},
			{"method": "DELETE", "path": "/autorzy/{id}", "description": "Usuń autora"},
			{"method": "GET", "path": "/utwory", "description": "Lista utworów"},
			{"method": "POST", "path": "/utwory", "description": "Dodaj utwór"},
			{"method": "GET", "path": "/utwory/{id}", "description": "Pobierz utwór po ID"},
			{"method": "PUT", "path": "/utwory/{id}", "description": "Zaktualizuj utwór"},
			{"method": "DELETE", "path": "/utwory/{id}", "description": "Usuń utwór"},
		},
		"examples": map[string]any{
			"create_author": map[string]any{
				"method": "POST",
				"path":   "/autorzy",
				"body": map[string]any{
					"name":    "Marek Grechuta",
					"country": "Polska",
				},
			},
			"create_track": map[string]any{
				"method": "POST",
				"path":   "/utwory",
				"body": map[string]any{
					"author_id":    1,
					"title":        "Dni, ktorych nie znamy",
					"genre":        "poezja spiewana",
					"release_year": 1971,
				},
			},
		},
	})
}

func (s *server) handleAuthorsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listAuthors(w)
	case http.MethodPost:
		s.createAuthor(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) handleAuthorResource(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r.URL.Path, "/autorzy/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getAuthor(w, id)
	case http.MethodPut:
		s.updateAuthor(w, r, id)
	case http.MethodDelete:
		s.deleteAuthor(w, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) handleTracksCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTracks(w)
	case http.MethodPost:
		s.createTrack(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) handleTrackResource(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r.URL.Path, "/utwory/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getTrack(w, id)
	case http.MethodPut:
		s.updateTrack(w, r, id)
	case http.MethodDelete:
		s.deleteTrack(w, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) listAuthors(w http.ResponseWriter) {
	rows, err := s.db.Query(`SELECT id, name, country, created_at FROM authors ORDER BY id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	results := make([]author, 0)
	for rows.Next() {
		var a author
		if err := rows.Scan(&a.ID, &a.Name, &a.Country, &a.Created); err != nil {
			writeError(w, http.StatusInternalServerError, "database error")
			return
		}
		results = append(results, a)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (s *server) getAuthor(w http.ResponseWriter, id int64) {
	var a author
	err := s.db.QueryRow(`SELECT id, name, country, created_at FROM authors WHERE id = ?`, id).Scan(&a.ID, &a.Name, &a.Country, &a.Created)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "author not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusOK, a)
}

func (s *server) createAuthor(w http.ResponseWriter, r *http.Request) {
	var req createAuthorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Country = strings.TrimSpace(req.Country)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	res, err := s.db.Exec(`INSERT INTO authors(name, country) VALUES(?, ?)`, req.Name, req.Country)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	s.getAuthor(w, id)
}

func (s *server) updateAuthor(w http.ResponseWriter, r *http.Request, id int64) {
	var req createAuthorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Country = strings.TrimSpace(req.Country)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	res, err := s.db.Exec(`UPDATE authors SET name = ?, country = ? WHERE id = ?`, req.Name, req.Country, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if affected == 0 {
		writeError(w, http.StatusNotFound, "author not found")
		return
	}

	s.getAuthor(w, id)
}

func (s *server) deleteAuthor(w http.ResponseWriter, id int64) {
	res, err := s.db.Exec(`DELETE FROM authors WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if affected == 0 {
		writeError(w, http.StatusNotFound, "author not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "author deleted"})
}

func (s *server) listTracks(w http.ResponseWriter) {
	rows, err := s.db.Query(`
		SELECT t.id, t.author_id, a.name, t.title, t.genre, t.release_year, t.created_at
		FROM tracks t
		JOIN authors a ON a.id = t.author_id
		ORDER BY t.id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	results := make([]track, 0)
	for rows.Next() {
		var t track
		if err := rows.Scan(&t.ID, &t.AuthorID, &t.AuthorName, &t.Title, &t.Genre, &t.ReleaseYear, &t.Created); err != nil {
			writeError(w, http.StatusInternalServerError, "database error")
			return
		}
		results = append(results, t)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (s *server) getTrack(w http.ResponseWriter, id int64) {
	var t track
	err := s.db.QueryRow(`
		SELECT t.id, t.author_id, a.name, t.title, t.genre, t.release_year, t.created_at
		FROM tracks t
		JOIN authors a ON a.id = t.author_id
		WHERE t.id = ?`, id).Scan(&t.ID, &t.AuthorID, &t.AuthorName, &t.Title, &t.Genre, &t.ReleaseYear, &t.Created)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "track not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (s *server) createTrack(w http.ResponseWriter, r *http.Request) {
	var req createTrackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Genre = strings.TrimSpace(req.Genre)
	if req.AuthorID <= 0 {
		writeError(w, http.StatusBadRequest, "author_id must be greater than 0")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	res, err := s.db.Exec(`INSERT INTO tracks(author_id, title, genre, release_year) VALUES(?, ?, ?, ?)`, req.AuthorID, req.Title, req.Genre, req.ReleaseYear)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "foreign key") {
			writeError(w, http.StatusBadRequest, "author does not exist")
			return
		}
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	s.getTrack(w, id)
}

func (s *server) updateTrack(w http.ResponseWriter, r *http.Request, id int64) {
	var req createTrackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Genre = strings.TrimSpace(req.Genre)
	if req.AuthorID <= 0 {
		writeError(w, http.StatusBadRequest, "author_id must be greater than 0")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	res, err := s.db.Exec(`UPDATE tracks SET author_id = ?, title = ?, genre = ?, release_year = ? WHERE id = ?`, req.AuthorID, req.Title, req.Genre, req.ReleaseYear, id)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "foreign key") {
			writeError(w, http.StatusBadRequest, "author does not exist")
			return
		}
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if affected == 0 {
		writeError(w, http.StatusNotFound, "track not found")
		return
	}

	s.getTrack(w, id)
}

func (s *server) deleteTrack(w http.ResponseWriter, id int64) {
	res, err := s.db.Exec(`DELETE FROM tracks WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if affected == 0 {
		writeError(w, http.StatusNotFound, "track not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "track deleted"})
}

func parseIDFromPath(path string, prefix string) (int64, error) {
	value := strings.TrimPrefix(path, prefix)
	if value == "" || strings.Contains(value, "/") {
		return 0, fmt.Errorf("invalid resource id")
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid resource id")
	}
	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON payload")
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("only one JSON object is allowed")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"encoding error"}`, http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
