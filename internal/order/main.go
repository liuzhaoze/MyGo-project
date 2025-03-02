package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux() // http 多路请求复用器

	// 注册路由和处理函数
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		if _, err := fmt.Fprintln(w, "<h1>Welcome to home page</h1>"); err != nil {
			log.Fatal(err)
		}
	})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, "pong"); err != nil {
			log.Fatal(err)
		}
	})

	log.Println("Listening on :8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatal(err)
	}
}
