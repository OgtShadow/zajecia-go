package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type ColorInfo struct {
	Color string `json:"color"`
	Lng   string `json:"lng"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

var colors = []struct {
	key    string
	namePL string
	nameEN string
	nameIT string
	value  string
}{
	{"zielony", "zielony", "green", "verde", "#00ff73"},
	{"czerwony", "czerwony", "red", "rosso", "#ff0000"},
	{"niebieski", "niebieski", "blue", "blu", "#0000ff"},
	{"żółty", "żółty", "yellow", "giallo", "#ffff00"},
	{"czarny", "czarny", "black", "nero", "#000000"},
}

func findColor(input string) (string, string, string, string, bool) {
	low := strings.ToLower(input)
	for _, c := range colors {
		if strings.ToLower(c.key) == low || strings.ToLower(c.namePL) == low || strings.ToLower(c.nameEN) == low || strings.ToLower(c.nameIT) == low {
			return c.key, c.namePL, c.nameEN, c.nameIT, true
		}
	}
	return "", "", "", "", false
}

func getNameByLang(pl, en, it, lang string) (string, bool) {
	switch strings.ToLower(lang) {
	case "pl":
		return pl, true
	case "en":
		return en, true
	case "it":
		return it, true
	default:
		return "", false
	}
}

func colorHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := strings.TrimSpace(q.Get("name"))
	lng := strings.TrimSpace(q.Get("lng"))
	if name == "" {
		http.Error(w, `missing query param "name"`, http.StatusBadRequest)
		return
	}
	if lng == "" {
		http.Error(w, `missing query param "lng"`, http.StatusBadRequest)
		return
	}

	key, pl, en, it, found := findColor(name)
	if !found {
		http.Error(w, fmt.Sprintf("color '%s' not found", name), http.StatusNotFound)
		return
	}

	targetName, ok := getNameByLang(pl, en, it, lng)
	if !ok {
		http.Error(w, fmt.Sprintf("unsupported language '%s'", lng), http.StatusBadRequest)
		return
	}

	result := ColorInfo{
		Color: name,
		Lng:   lng,
		Name:  targetName,
		Value: "",
	}

	for _, c := range colors {
		if strings.ToLower(c.key) == strings.ToLower(key) {
			result.Value = c.value
			break
		}
	}

	if result.Value == "" {
		http.Error(w, "RGB value missing for color", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/color", colorHandler)
	fmt.Println("Color service listening at :8080")
	fmt.Println("request struct: http://localhost:8080/color?name=[name]&lng=[language]")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
