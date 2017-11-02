package Assignment2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_HandlerPost(t *testing.T) {

	test := httptest.NewServer(http.HandlerFunc(HandlerPost))
	defer test.Close()

	temp := WebHookS{"", "www.example.com", "EUR", "NOK", float64(1.5), float64(2.55)}
	res, err := json.Marshal(&temp)
	if err != nil {
		t.Errorf("Could not marshal", err)
	}

	post, err := http.Post(test.URL, "application/json", bytes.NewBuffer(res))
	if err != nil {
		t.Errorf("Could not post", err)
	}
	fmt.Println(post)
}

func Test_HandlerLate(t *testing.T) {
	test := httptest.NewServer(http.HandlerFunc(HandlerLate))
	defer test.Close()

	temp := make(map[string]interface{})

	temp["baseCurrency"] = "EUR"
	temp["targetCurrency"] = "NOK"

	res, err := json.Marshal(&temp)
	if err != nil {
		t.Errorf("Could not marshal", err)
		return
	}

	post, err := http.Post(test.URL, "application/json", bytes.NewBuffer(res))
	if err != nil {
		t.Errorf("Could not post", err)
		return
	}
	fmt.Println(post)

}

func Test_HandlerAverage(t *testing.T) {
	test := httptest.NewServer(http.HandlerFunc(HandlerAvg))
	defer test.Close()

	temp := make(map[string]interface{})

	temp["baseCurrency"] = "EUR"
	temp["targetCurrency"] = "NOK"

	res, err := json.Marshal(&temp)
	if err != nil {
		t.Errorf("Could not marshal", err)
		return
	}

	post, err := http.Post(test.URL, "application/json", bytes.NewBuffer(res))
	if err != nil {
		t.Errorf("Could not post", err)
		return
	}
	fmt.Println(post)
}

func Test_HandlerEva(t *testing.T) {
	test := httptest.NewServer(http.HandlerFunc(HandlerEva))
	defer test.Close()

	get, err := http.Get(test.URL)
	if err != nil {
		t.Errorf("Could not get", err)
		return
	}
	fmt.Println(get)
}
