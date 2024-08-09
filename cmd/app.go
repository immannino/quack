package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/marcboeker/go-duckdb"
)

var db *sql.DB

func init() {
	conn, err := sql.Open("duckdb", "")
	if err != nil {
		panic(err)
	}
	db = conn
}

func Start() {
	router := mux.NewRouter()

	router.HandleFunc("/exec", ExecHandler).Methods("POST")
	fs := http.FileServer(http.Dir(filepath.Join("./static")))
	router.PathPrefix("/").Handler(http.StripPrefix("/", fs))

	srv := &http.Server{
		Addr: "0.0.0.0:1234",

		// Slowloris is no friend of mine
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,

		Handler: router,
	}

	fmt.Println("Server started on 0.0.0.0:1234")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

func Dir(w http.ResponseWriter, r *http.Request) {

}

type ExecBody struct {
	File  string `json:"File"`
	Query string `json:"Query"`
}

func ExecHandler(w http.ResponseWriter, r *http.Request) {
	var body *ExecBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("tee hee"))
		return
	}

	rows, err := db.Query(body.Query)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	data := []interface{}{}
	types, _ := rows.ColumnTypes()
	header, _ := rows.Columns()
	data = append(data, header)

	for rows.Next() {
		row := make([]interface{}, len(types))
		for i := range types {
			row[i] = new(interface{})
		}
		rows.Scan(row...)
		data = append(data, row)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
