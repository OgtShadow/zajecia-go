package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var icons = map[string]string{
	"sunny":               "slonecznie.png",
	"slonecznie":          "slonecznie.png",
	"cloudy":              "zachmurzenie.png",
	"zachmurzenie":        "zachmurzenie.png",
	"partly_cloudy":       "lekkie-zachmurzenie.png",
	"lekkie_zachmurzenie": "lekkie-zachmurzenie.png",
	"rain":                "deszcz.png",
	"deszcz":              "deszcz.png",
	"heavy_rain":          "ulewa.png",
	"ulewa":               "ulewa.png",
	"thunder":             "burza.png",
	"burza":               "burza.png",
	"snow":                "snieg.png",
	"snieg":               "snieg.png",
	"sleet":               "deszcz-ze-sniegiem.png",
	"deszcz-ze-sniegiem":  "deszcz-ze-sniegiem.png",
	"fog":                 "mgla.png",
	"mgla":                "mgla.png",
	"drizzle":             "mzawka.png",
	"mzawka":              "mzawka.png",
}

type IconResponse struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type IconsResponse struct {
	Icons []string `json:"icons"`
}

func iconsHandler(w http.ResponseWriter, r *http.Request) {
	unique := make([]string, 0, len(icons))
	seen := map[string]bool{}
	for key := range icons {
		if !seen[key] {
			seen[key] = true
			unique = append(unique, key)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IconsResponse{Icons: unique})
}

func iconHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := strings.TrimSpace(q.Get("name"))
	if name == "" {
		http.Error(w, `missing query parameter "name"`, http.StatusBadRequest)
		return
	}

	format := strings.TrimSpace(q.Get("format"))
	if format == "" {
		format = "raw"
	}

	key := strings.ToLower(name)
	fileName, ok := icons[key]
	if !ok {
		http.Error(w, fmt.Sprintf("icon '%s' not found", name), http.StatusNotFound)
		return
	}

	iconPath := filepath.Join("weather icons", fileName)
	imgBytes, err := os.ReadFile(iconPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read icon file: %v", err), http.StatusInternalServerError)
		return
	}

	if format == "base64" {
		encoded := base64.StdEncoding.EncodeToString(imgBytes)
		resp := IconResponse{Name: name, Data: encoded}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// raw PNG response
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Write(imgBytes)
}

func main() {
	http.HandleFunc("/icon", iconHandler)
	http.HandleFunc("/icons", iconsHandler)
	fmt.Println("Icon service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
