package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

const (
	DBName = "sample_bank"
)

type jsonTime time.Time

func (t jsonTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(time.Time(t).Format(time.RFC3339))), nil
}

func (t *jsonTime) UnmarshalJSON(s []byte) (err error) {
	q, err := strconv.Unquote(string(s))
	if err != nil {
		return err
	}
	*(*time.Time)(t), err = time.Parse(time.RFC3339, q)
	return
}

//TimeStamp ...
type TimeStamp struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`
}

//Account  ...
type Account struct {
	ID       int64  `json:"id"`
	Type     string `json:"accountType"`
	ClientID int64  `json:"clientId"`
	TimeStamp
}

//Client  ...
type Client struct {
	ID            int64     `json:"id"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	DateOfBirth   time.Time `json:"dateOfBirth"`
	PhoneNumber   string    `json:"phoneNumber"`
	Accounts      []Account `json:"accounts"`
	AccountNumber int64     `json:"accountNumber"`
	Branch        string    `json:"branch"`
	Occupation    string    `json:"occupation"`
	SnnitNumber   int64     `json:"snnitNumber"`
	TimeStamp
}

//AccountTransactions ....
type AccountTransactions struct {
	Amount          int64     `json:"amount"`
	Moment          time.Time `json:"transactionDate"`
	AccountID       int64     `json:"accountId"`
	TransactionType string    `json:"trasanctionType"` //deposit, withdrawal, transfer
	TransactionBy   string    `json:"transactionBy"`   //person
	TransactionFrom string    `json:"transactionFrom"` //atm, bank, online
	TimeStamp
}

//AppDb ...
type AppDb struct {
	Db gorm.DB
}

func (appDb *AppDb) initDb() {
	var err error
	var dbConfig = fmt.Sprintf("dbname=%s sslmode=disable", DBName)
	appDb.Db, err = gorm.Open("postgres", dbConfig)
	if err != nil {
		panic(err)
	}
	appDb.Db.LogMode(true)
}

func (appDb *AppDb) initSchema() {
	appDb.Db.AutoMigrate(&Account{})
	appDb.Db.AutoMigrate(&Client{})
	appDb.Db.AutoMigrate(&AccountTransactions{})
}

func (appDb *AppDb) createDummyData() {
	t := time.Now()
	ts := TimeStamp{}
	ts.CreatedAt = t
	ts.UpdatedAt = t

	accounts := []Account{
		{Type: "checking", TimeStamp: ts},
		{Type: "savings", TimeStamp: ts},
		{Type: "current", TimeStamp: ts},
	}

	client := Client{
		FirstName:     "Kofi",
		LastName:      "Mensah",
		DateOfBirth:   time.Now(),
		Accounts:      accounts,
		AccountNumber: 111222333345,
		Branch:        "Dansoman",
	}

	appDb.Db.Save(&client)
}

func main() {
	fmt.Println("Welcome to Sample bank Server:::::::")
	fmt.Println(time.Now())
	///DO db stuff
	appDb := AppDb{}
	appDb.initDb()
	appDb.initSchema()
	// appDb.createDummyData()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3500"},
	})

	// c := cors.Default()
	r := mux.NewRouter().StrictSlash(false)
	accounts := r.Path("/accounts").Subrouter()
	accounts.Methods("GET").HandlerFunc(appDb.accountsHandler)
	accounts.Methods("POST").HandlerFunc(appDb.createClientsHandler)

	account := r.PathPrefix("/accounts/{id}").Subrouter()
	account.Methods("GET").HandlerFunc(accountHandler)

	http.ListenAndServe(":8050", c.Handler(r))
}

func (appDb *AppDb) accountsHandler(rw http.ResponseWriter, r *http.Request) {
	clients := []Client{}
	appDb.Db.Find(&clients)
	js, err := json.Marshal(clients)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	rw.Header().Set("Content-type", "application/json")
	fmt.Println("rw", "accounts")
	rw.Write(js)
}

func accountHandler(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	fmt.Println("rw", "accounts", id)
}

func (appDb *AppDb) createClientsHandler(rw http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	var client Client
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&client)
	if err != nil {
		panic(err)
	}
	fmt.Println("Creating")

	fmt.Println("params are .... ")
	fmt.Printf("%+v\n", client)

	appDb.Db.Save(&client)

	rw.Header().Set("Content-type", "application/json")
	js, err := json.Marshal(client)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rw.Write(js)
}
