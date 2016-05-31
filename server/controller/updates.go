package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/egnwd/outgain/server/lobby"
	"github.com/gorilla/mux"
	"gopkg.in/antage/eventsource.v1"
)

func UpdatesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("In")
		vars := mux.Vars(r)
		id, _ := strconv.ParseUint(vars["id"], 10, 64)
		log.Printf("ID: %d\n", id)

		l, ok := lobby.GetLobby(uint64(id))
		log.Printf("Lobby: %#v\n", l)
		if !ok {
			http.Error(w, "Lobby doesn't exist", http.StatusInternalServerError)
			return
		}

		eng := l.Engine

		fmt.Printf("\n%#v\n\n", eng)

		events := eventsource.New(nil, nil)
		go func() {
			for event := range eng.Events {
				packet, err := json.Marshal(event.Data)
				if err != nil {
					log.Printf("JSON serialization failed %v", err)
				} else {
					events.SendEventMessage(string(packet), event.Type, "")
				}
			}
		}()

		events.ServeHTTP(w, r)
	})
}
