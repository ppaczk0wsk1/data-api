package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var db *sql.DB
var err error

func main() {

	config.WithOptions(config.ParseEnv)

	// add driver for support yaml content
	config.AddDriver(yaml.Driver)

	err := config.LoadFiles("config.yml")
	if err != nil {
		panic(err)
	}

	// Open Mysql Connection
	db, err = sql.Open("mysql", config.String("name")+":"+config.String("password")+"!@tcp("+config.String("Host")+":"+config.String("port")+")/data")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	//Set router
	router := mux.NewRouter()

	router.HandleFunc("/data/worldwide", getDataWorldwide).Methods("GET")
	router.HandleFunc("/data/countries/last", getDataCountriesLast).Methods("GET")
	router.HandleFunc("/data/countries/delta/last", getDataCountriesDeltaLast).Methods("GET")
	router.HandleFunc("/data/countries/confirmed", getDataConfirmed).Methods("GET")
	router.HandleFunc("/data/countries/confirmed/delta", getDataDeltaConfirmed).Methods("GET")
	router.HandleFunc("/data/countries/active/delta", getDataDeltaActive).Methods("GET")
	router.HandleFunc("/data/countries/list", getDataCountriesList).Methods("GET")
	router.HandleFunc("/data/countries/{country}", getDataCountries).Methods("GET")
	router.HandleFunc("/data/countries/{country}/delta", getDataCountriesDelta).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5000"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Fatal(http.ListenAndServe(":3001", handler))
}
