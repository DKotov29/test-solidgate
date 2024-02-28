package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	creditcard "github.com/durango/go-credit-card"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/validate", validate).Methods("POST")
	http.ListenAndServe(":8080", router)
}

const jsonValid = "{\"valid\":true}"
const layout = "2006-01"

type request struct {
	Card  string `json:"card"`
	Year  string `json:"year"`
	Month string `json:"month"`
}

func validate(w http.ResponseWriter, r *http.Request) {
	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		jss, err := toJsonError(2, "Bad json")
		if err != nil {
			log.Println("Problem with json creating for bad json reques")
			return
		}
		http.Error(w, string(jss[:]), 400)
		return
	}

	card := req.Card
	emonth := req.Month
	eyear := req.Year

	t := time.Now()
	toparse := eyear + "-" + emonth
	parsed, err := time.Parse(layout, toparse)

	if err != nil {
		http.Error(w, "You provided bad month or year. We expect date in such layout: \"2020-01\"", 400)
		return
	}

	if parsed.Before(t) {
		jss, err := toJsonError(1, "Provided year and month expired")
		if err != nil {
			log.Println("Problem with json creating for expired card")
			return
		}
		http.Error(w, string(jss[:]), 400)
		return
	}

	ccard := creditcard.Card{Number: card, Cvv: "1111", Month: emonth, Year: eyear}

	err = ccard.Validate(true)
	if err != nil {
		jss, err := toJsonError(3, "Card number is wrong")
		if err != nil {
			log.Println("Problem with json creating for wrong number cards")
			return
		}
		http.Error(w, string(jss[:]), 400)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte(jsonValid))

}

func toJsonError(id int, msg string) ([]byte, error) {
	data := map[string]interface{}{
		"valid":   false,
		"code":    id,
		"message": msg,
	}

	jsonData, err := json.Marshal(data)
	return jsonData, err
}
