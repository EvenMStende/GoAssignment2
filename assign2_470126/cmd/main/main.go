package main

import (
	"Assignment2"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/example", Assignment2.HandlerPost)
	http.HandleFunc("/example/", Assignment2.HandlerDel)
	http.HandleFunc("/example/latest", Assignment2.HandlerLate)
	http.HandleFunc("/example/average", Assignment2.HandlerAvg)
	http.HandleFunc("/example/evaluationtrigger", Assignment2.HandlerEva)
	http.ListenAndServe(":"+port, nil)
}
