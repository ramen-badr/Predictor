package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type AgeResponse struct {
	Age int `json:"age"`
}

type GenderResponse struct {
	Gender string `json:"gender"`
}

type NationalityResponse struct {
	Country []struct {
		CountryID string `json:"country_id"`
	} `json:"country"`
}

func GetAge(name string) int {
	resp, err := http.Get("https://api.agify.io/?name=" + name)
	if err != nil {
		log.Printf("Error getting age: %v", err)
		return 0
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data AgeResponse

	json.Unmarshal(body, &data)

	return data.Age
}

func GetGender(name string) string {
	resp, err := http.Get("https://api.genderize.io/?name=" + name)
	if err != nil {
		log.Printf("Error getting gender: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data GenderResponse

	json.Unmarshal(body, &data)

	return data.Gender
}

func GetNationality(name string) string {
	resp, err := http.Get("https://api.nationalize.io/?name=" + name)
	if err != nil {
		log.Printf("Error getting nationality: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data NationalityResponse

	json.Unmarshal(body, &data)

	if len(data.Country) > 0 {
		return data.Country[0].CountryID
	}
	return ""
}
