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

// var t *template.Template
// var debug string

var db *sql.DB

func init() {
	conn, err := sql.Open("duckdb", "")
	if err != nil {
		panic(err)
	}
	db = conn
}

// go:embed *.html
// var fs embed.FS

//	func init() {
//		t = template.Must(template.ParseFS(fs, "./*.html"))
//		debug = os.Getenv("DEBUG")
//	}
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

func PageHandler(w http.ResponseWriter, r *http.Request) {

	// html(w, 200, "index", nil)
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

// func html(w http.ResponseWriter, httpStatus int, templateName string, data interface{}) {
// 	w.Header().Add("Content-Type", "text/html")
// 	w.WriteHeader(http.StatusOK)

// 	if len(debug) > 0 {
// 		d := template.Must(template.ParseGlob(filepath.Join("./cmd", "*.html")))
// 		tmp := d.Lookup(templateName)
// 		err := tmp.Execute(w, data)

// 		if err != nil {
// 			log.Println("Error executing template :", err)
// 		}

// 		return
// 	}

// 	tmp := t.Lookup(templateName)
// 	err := tmp.Execute(w, data)

// 	if err != nil {
// 		log.Println("Error executing template :", err)
// 	}
// }
