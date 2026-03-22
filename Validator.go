package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ValidateResponse struct {
	IsValid   bool   `json:"is_valid"`
	Birthdate string `json:"birthdate,omitempty"`
	Gender    string `json:"gender,omitempty"`
}

func validatePesel(pesel string) (ValidateResponse, error) {
	pesel = strings.TrimSpace(pesel)
	if len(pesel) != 11 {
		return ValidateResponse{IsValid: false}, nil
	}
	for _, ch := range pesel {
		if ch < '0' || ch > '9' {
			return ValidateResponse{IsValid: false}, nil
		}
	}

	weights := []int{1, 3, 7, 9, 1, 3, 7, 9, 1, 3}
	checkSum := 0
	for i := 0; i < 10; i++ {
		digit := int(pesel[i] - '0')
		checkSum += digit * weights[i]
	}
	checkDigit := (10 - (checkSum % 10)) % 10
	if checkDigit != int(pesel[10]-'0') {
		return ValidateResponse{IsValid: false}, nil
	}

	year, _ := strconv.Atoi(pesel[0:2])
	month, _ := strconv.Atoi(pesel[2:4])
	day, _ := strconv.Atoi(pesel[4:6])
	century := 1900
	if month > 80 && month < 93 {
		century = 1800
		month -= 80
	} else if month > 60 && month < 73 {
		century = 2200
		month -= 60
	} else if month > 40 && month < 53 {
		century = 2100
		month -= 40
	} else if month > 20 && month < 33 {
		century = 2000
		month -= 20
	} else if month > 0 && month < 13 {
		century = 1900
	} else {
		return ValidateResponse{IsValid: false}, nil
	}

	year += century
	dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return ValidateResponse{IsValid: false}, nil
	}

	sexDigit := int(pesel[9] - '0')
	gender := "M"
	if sexDigit%2 == 0 {
		gender = "F"
	}

	return ValidateResponse{IsValid: true, Birthdate: dateStr, Gender: gender}, nil
}

func validateNip(nip string) (bool, error) {
	nip = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(nip), "-", ""), " ", "")
	if len(nip) != 10 {
		return false, nil
	}
	weights := []int{6, 5, 7, 2, 3, 4, 5, 6, 7}
	checksum := 0
	for i := 0; i < 9; i++ {
		if nip[i] < '0' || nip[i] > '9' {
			return false, nil
		}
		d := int(nip[i] - '0')
		checksum += d * weights[i]
	}
	control := checksum % 11
	if control == 10 {
		return false, nil
	}
	if control != int(nip[9]-'0') {
		return false, nil
	}
	return true, nil
}

func validateRegon(regon string) (bool, error) {
	regon = strings.TrimSpace(regon)
	regon = strings.ReplaceAll(regon, "-", "")
	if len(regon) != 9 && len(regon) != 14 {
		return false, nil
	}
	expected := []int{8, 9, 2, 3, 4, 5, 6, 7}
	checksum := 0
	for i := 0; i < 8; i++ {
		if regon[i] < '0' || regon[i] > '9' {
			return false, nil
		}
		checksum += int(regon[i]-'0') * expected[i]
	}
	control := checksum % 11
	if control == 10 {
		control = 0
	}
	if control != int(regon[8]-'0') {
		return false, nil
	}
	// 14-digit check if needed
	if len(regon) == 14 {
		weights14 := []int{2, 4, 8, 5, 0, 9, 7, 3, 6, 1, 2, 4, 8}
		checksum14 := 0
		for i := 0; i < 13; i++ {
			if regon[i] < '0' || regon[i] > '9' {
				return false, nil
			}
			checksum14 += int(regon[i]-'0') * weights14[i]
		}
		control14 := checksum14 % 11
		if control14 == 10 {
			control14 = 0
		}
		if control14 != int(regon[13]-'0') {
			return false, nil
		}
	}
	return true, nil
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	pesel := strings.TrimSpace(q.Get("pesel"))
	nip := strings.TrimSpace(q.Get("nip"))
	regon := strings.TrimSpace(q.Get("regon"))

	var res ValidateResponse
	var err error

	if pesel != "" {
		res, err = validatePesel(pesel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if nip != "" {
		valid, err := validateNip(nip)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res = ValidateResponse{IsValid: valid}
	} else if regon != "" {
		valid, err := validateRegon(regon)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res = ValidateResponse{IsValid: valid}
	} else {
		http.Error(w, "missing pesel/nip/regon parameter", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	http.HandleFunc("/validate", validateHandler)
	fmt.Println("Validator service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
