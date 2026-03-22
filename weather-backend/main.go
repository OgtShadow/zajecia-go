package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type WeatherResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Current   struct {
		Temperature float64 `json:"temperature"`
	} `json:"current_weather"`
	Hourly struct {
		Time        []string  `json:"time"`
		Temperature []float64 `json:"temperature_2m"`
	} `json:"hourly"`
}

type HourlyTemperature struct {
	Time        string
	Temperature float64
}

func getWeather(lat, lon string) (float64, []HourlyTemperature, error) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current_weather=true&hourly=temperature_2m", lat, lon)
	resp, err := http.Get(url)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return 0, nil, err
	}

	hourly := []HourlyTemperature{}
	for i := range weather.Hourly.Time {
		hourly = append(hourly, HourlyTemperature{
			Time:        weather.Hourly.Time[i],
			Temperature: weather.Hourly.Temperature[i],
		})
	}
	return weather.Current.Temperature, hourly, nil
}

func DataSort(hourly []HourlyTemperature) []HourlyTemperature {
	day := time.Now().Format("2006-01-02")
	filtered := make([]HourlyTemperature, 0, len(hourly))
	for _, h := range hourly {
		if strings.HasPrefix(h.Time, day) {
			filtered = append(filtered, h)
		}
	}
	return filtered
}

func coordinatesSetter() (string, string) {
	var lat, lon string
	fmt.Print("Enter latitude: ")
	_, err := fmt.Scanln(&lat)
	if err != nil {
		log.Fatalf("Invalid latitude input: %v", err)
	}
	fmt.Print("Enter longitude: ")
	_, err = fmt.Scanln(&lon)
	if err != nil {
		log.Fatalf("Invalid longitude input: %v", err)
	}

	return lat, lon

}
func main() {
	lat, lon := coordinatesSetter()
	temp, hourly, err := getWeather(lat, lon)
	if err != nil {
		log.Fatalf("Failed to get weather: %v", err)
	}

	fmt.Printf("\nCurrent temperature at %s,%s: %.2f°C\n", lat, lon, temp)

	hourly = DataSort(hourly)

	if len(hourly) == 0 {
		fmt.Println("Brak godzinowych danych dla dzisiejszej daty.")
		return
	}

	fmt.Println("Hourly forecast for today:")
	fmt.Printf("%-20s %s\n", "Time", "Temp (°C)")
	fmt.Println("-----------------------------------")
	for _, h := range hourly {
		fmt.Printf("%-20s %.2f\n", h.Time, h.Temperature)
	}
}
