package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
	"os"
)

type Welcome struct {
	Name string
	Time string
	Hostname string
}


func main() {
	hostname, errhost := os.Hostname()
	welcome := Welcome{"Anonymous", time.Now().Format(time.Stamp), "hostname"}

	templates := template.Must(template.ParseFiles("templates/welcome-template.html"))
	http.Handle("/static/", 
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static")))) 

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {


		if name := r.FormValue("name"); name != "" {
			welcome.Name = name
		}

		if err := templates.ExecuteTemplate(w, "welcome-template.html", welcome); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	if errhost == nil {
		fmt.Println("Served by: ", hostname)
		welcome.Hostname = hostname
	}
	fmt.Println("Listening on Port 8080")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
