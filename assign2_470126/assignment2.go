package Assignment2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	//GET const to avoid errors
	GET = "GET"
	//POST const to avoid errors
	POST = "POST"
	//URL const fpr DB to make it easier
	URL = "mongodb://admin:admin@ds133465.mlab.com:33465/assignment2wh"
)

//WebHookS struct
type WebHookS struct {
	ID              bson.ObjectId `json:"-" bson:"_id"`
	WebhookURL      string        `json:"webhookURL"`
	BaseCurrency    string        `json:"baseCurrency"`
	TargetCurrency  string        `json:"targetCurrency"`
	MinTriggerValue float64       `json:"minTriggerValue"`
	MaxTriggerValue float64       `json:"maxTriggerValue"`
}

//FixerData struct
type FixerData struct {
	Base  string                 `json:"base"`
	Date  string                 `json:"date"`
	Rates map[string]interface{} `json:"rates"`
}

//HandlerPost func ...
func HandlerPost(w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")

	if r.Method == POST {

		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		temp := &WebHookS{}
		json.Unmarshal(payload, &temp)

		session, err := mgo.Dial(URL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		c := session.DB("assignment2wh").C("collectionWH")
		temp.ID = bson.NewObjectId()
		defer session.Close()
		err = c.Insert(&temp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}

		fmt.Print(temp.ID.Hex())
		w.Write([]byte(temp.ID.Hex()))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//HandlerDel func ...
func HandlerDel(w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")

	if r.Method == GET || r.Method == "DELETE" {

		args := r.URL.Path
		part := strings.Split(args, "/")
		id := part[2]

		session, err := mgo.Dial(URL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		c := session.DB("assignment2wh").C("collectionWH")
		temp := &WebHookS{}
		defer session.Close()

		if r.Method == "GET" {
			fmt.Print(id)
			err = c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(temp)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			}

			json.NewEncoder(w).Encode(temp)
		} else {
			err := c.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
			fmt.Printf("Deleted the thing")
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			}
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

//HandlerLate func ...
func HandlerLate(w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")

	if r.Method == POST {

		session, err := mgo.Dial(URL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		c := session.DB("assignment2wh").C("collectionRates")
		defer session.Close()
		temp := &FixerData{}

		err = c.Find(nil).Sort("$natural: 1").One(&temp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if temp.Date != strings.Split(time.Now().Local().String(), " ")[0] {

			fixerURL := "http://api.fixer.io/latest"

			re, erro := http.Get(fixerURL)
			if erro != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			defer re.Body.Close()
			json.NewDecoder(re.Body).Decode(&temp)

			temp.Date = strings.Split(time.Now().Local().String(), " ")[0]
			err = c.Insert(&temp)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		} else {
			fmt.Print("Matching dates")
		}

		err = c.Find(nil).Sort("$natural: 1").One(temp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		re, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		temp2 := &WebHookS{}
		err = json.Unmarshal(re, &temp2)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		//If BaseCurrency and TargetCurrency is the same it returns 1
		if temp2.BaseCurrency == temp2.TargetCurrency {
			fmt.Fprint(w, "1")

			//If BaseCurrency is the same as the BaseCurrency requested
			//	it simply returns the TargetCurrency requested
		} else if temp.Base == temp2.BaseCurrency {
			fmt.Fprint(w, temp.Rates[temp2.TargetCurrency])
		} else {
			//If the BaseCurrency is the same as the TargetCurrency requested,
			//	a new base needs to be used to calculate the new rate
			if temp.Base == temp2.TargetCurrency {
				newBase := temp.Rates[temp2.BaseCurrency]
				newRate := 1 / newBase.(float64)
				fmt.Fprint(w, newRate)
			} else {
				//If the BaseCurrency is not the BaseCurrency requested
				//	a new base must be used and a new TargetCurrency must be used
				newBase := temp.Rates[temp2.TargetCurrency]
				newTarget := temp.Rates[temp2.BaseCurrency]
				newRate := newBase.(float64) / newTarget.(float64)
				fmt.Fprint(w, newRate)
			}
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//HandlerAvg func ...
func HandlerAvg(w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")
	URL := "mongodb://admin:admin@ds133465.mlab.com:33465/assignment2wh"

	if r.Method == POST {

		var Avg []FixerData
		var n float64
		var m float64

		session, err := mgo.Dial(URL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		c := session.DB("assignment2wh").C("collectionRates")
		defer session.Close()

		err = c.Find(nil).Sort("$natural: 1").Limit(3).All(&Avg)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		re, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		temp2 := &WebHookS{}
		err = json.Unmarshal(re, &temp2)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		//If BaseCurrency and TargetCurrency is the same it returns 1
		if temp2.BaseCurrency == temp2.TargetCurrency {
			fmt.Fprint(w, "1")

		} else if temp2.TargetCurrency == Avg[0].Base {
			for i := range Avg {
				n += Avg[i].Rates[temp2.BaseCurrency].(float64)
			}
			avgBase := n / 3
			fmt.Fprint(w, 1/avgBase)
		} else if temp2.TargetCurrency == Avg[0].Base {
			for i := range Avg {
				n += Avg[i].Rates[temp2.BaseCurrency].(float64)
			}
			avgBase := n / 3
			fmt.Fprint(w, 1/avgBase)

		} else if Avg[0].Base == temp2.BaseCurrency {
			for i := range Avg {
				m += Avg[i].Rates[temp2.TargetCurrency].(float64)
			}
			avgTarget := m / 3
			fmt.Fprint(w, avgTarget)

		} else {
			for i := range Avg {
				n += Avg[i].Rates[temp2.BaseCurrency].(float64)
				m += Avg[i].Rates[temp2.TargetCurrency].(float64)
			}
			avgBase := m / 3
			avgTarget := n / 3
			answer := avgBase / avgTarget
			fmt.Fprint(w, answer)
		}

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//HandlerEva func ...
func HandlerEva(w http.ResponseWriter, r *http.Request) {

	if r.Method == GET {

		var hooks []WebHookS
		temp := &FixerData{}

		session, err := mgo.Dial(URL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		session2, err := mgo.Dial(URL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		b := session.DB("assignment2wh").C("collectionWH")
		defer session.Close()
		c := session2.DB("assignment2wh").C("collectionRates")
		defer session2.Close()

		err = b.Find(nil).Sort("$natural: 1").All(&hooks)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err = c.Find(nil).Sort("$natural: 1").One(&temp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		for i := range hooks {
			Invoker(hooks[i], *temp)
		}
	}
}

//Invoker func ...
func Invoker(hook WebHookS, temp FixerData) {

	res := make(map[string]interface{})

	res["baseCurrency"] = hook.BaseCurrency
	res["targetCurrency"] = hook.TargetCurrency
	res["currentRate"] = temp.Rates[hook.TargetCurrency]
	res["minTriggerValue"] = hook.MinTriggerValue
	res["maxTriggerValue"] = hook.MaxTriggerValue

	post, err := json.Marshal(res)
	if err != nil {
		fmt.Println(err)
		return
	}

	r, err := http.Post(hook.WebhookURL, "application/json", bytes.NewBuffer(post))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r.StatusCode)
}
