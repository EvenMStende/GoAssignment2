package main

import (
		"encoding/json"
		"fmt"
		"net/http"
		"strings"
		"time"
		"gopkg.in/mgo.v2"
		"Assignment2"
)

//DailyClock updates the database with the new exchange rates every 24 hours
func DailyClock() {

		session, err := mgo.Dial(Assignment2.URL)
		if err != nil {
			fmt.Print(err)
			return
		}

		c := session.DB("assignment2wh").C("collectionRates")
		defer session.Close()

		temp := &Assignment2.FixerData{}
		err = c.Find(nil).Sort("$natural: 1").One(&temp)
		if err != nil {
			fmt.Print(err)
			return
		}

		fixerURL := "http://api.fixer.io/latest"

		re, err := http.Get(fixerURL)
		if err != nil {
			fmt.Print(err)
			return
		}

		defer re.Body.Close()
		json.NewDecoder(re.Body).Decode(&temp)

		temp.Date = strings.Split(time.Now().Local().String(), " ")[0]
		err = c.Insert(&temp)
		if err != nil {
			fmt.Print(err)
			return
		}

		var hooks []Assignment2.WebHookS

		session2, err := mgo.Dial(Assignment2.URL)
		if err != nil {
			fmt.Print(err)
			return
		}

		b := session.DB("assignment2wh").C("collectionWH")
		defer session.Close()
		c = session2.DB("assignment2wh").C("collectionRates")
		defer session2.Close()

		err = b.Find(nil).Sort("$natural: 1").All(&hooks)
		if err != nil {
			fmt.Print(err)
			return
		}

		err = c.Find(nil).Sort("$natural: 1").One(&temp)
		if err != nil {
			fmt.Print(err)
			return
		}

		for i := range hooks {
			if hooks[i].BaseCurrency == temp.Base {
				if hooks[i].MinTriggerValue > temp.Rates[hooks[i].TargetCurrency].(float64) {
					Assignment2.Invoker(hooks[i], *temp)
				} else if hooks[i].MaxTriggerValue < temp.Rates[hooks[i].TargetCurrency].(float64) {
					Assignment2.Invoker(hooks[i], *temp)
				}
			} else {
				fmt.Print(hooks[i].BaseCurrency)
				fmt.Print(" is not a valid base currency/not implemented\n")
			}
		}
}

func main() {
	for {
		DailyClock()
		time.Sleep(time.Hour * 24)
	}
}
