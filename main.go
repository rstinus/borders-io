package main

import (
	"Jeu/engine"
	"Jeu/network"
	"log"
	"net/http"
)

func main() {
	hub := network.NewHub()
	engine.LoadSymmetricMap()
	go hub.Run()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		network.ServeWs(hub, w, r)
	})

	log.Println("Serveur Go sur :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
