package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/lithammer/shortuuid/v4"
)

type Mapper struct {
	Mapping map[string]string
	Lock sync.Mutex
}

var urlMapper Mapper

func init() {
	urlMapper = Mapper{
		Mapping: make(map[string]string),
	}
}

func main() {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is running..."))
	})

	r.Post("/short-it", createShortURLHandler)
	r.Get("/short/{key}", redirectHandler)
	http.ListenAndServe(":3000", r)
}

func createShortURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	u := r.Form.Get("URL")

	if u == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL is required"))
	}

	//generate key
	key := shortuuid.New()

	//insert
	insertMapping(key, u)

	log.Println("url mapped successfully")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("http://localhost:3000/short/%s", key)))

	
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("key field is empty"))
	}

	//fetch mapping
	u := fetchMapping(key)

	if u == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("URL not found"))
	}

	http.Redirect(w, r, u, http.StatusFound)
}

func insertMapping(key string, u string) {
	urlMapper.Lock.Lock()
	defer urlMapper.Lock.Unlock()
	urlMapper.Mapping[key] = u
}

func fetchMapping(key string) string {
	urlMapper.Lock.Lock()
	defer urlMapper.Lock.Unlock()
	return urlMapper.Mapping[key]
}
