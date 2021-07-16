package main

import (
	"fmt"
	"net/http"
	"time"

	bolt "github.com/boltdb/bolt"
	urlshort "github.com/ei09010/urlshort"
)

func main() {
	mux := defaultMux()

	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		panic(err)
	}

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	// Build the YAMLHandler using the mapHandler as the
	// fallback
	// 	yaml := `
	// - path: /urlshort
	//   url: https://github.com/gophercises/urlshort
	// - path: /urlshort-final
	//   url: https://github.com/gophercises/urlshort/tree/solution
	// `

	//yamlHandler, err := urlshort.YAMLHandler([]byte(yaml), "../conf.yaml", mapHandler)

	// 	json := `{
	//     "PathUrl": [
	//         {
	//             "path" : "/urlshort",
	//             "url": "https://github.com/gophercises/urlshort"
	//         },
	//         {
	//             "path" : "/urlshort-final",
	//             "url": "https://github.com/gophercises/urlshort/tree/solution"
	//         }
	//     ]
	// }`

	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	defer db.Close()

	// 	jsonHandler, err := urlshort.JSONHandler([]byte(json), "", mapHandler)

	// 	if err != nil {
	// 		panic(err)
	// 	}

	dbHandler := urlshort.DBHandler(db, mapHandler)

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", dbHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
