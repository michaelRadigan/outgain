package controller

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"

	"github.com/egnwd/outgain/server/config"
	"github.com/egnwd/outgain/server/lobby"
	"github.com/gorilla/mux"
)

func LobbiesView(staticDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/lobbies.html")
	})
}

// LobbiesPeek returns all the lobbies serialized as JSON
func LobbiesPeek(w http.ResponseWriter, r *http.Request) {
	data := lobby.Serialize()
	log.Printf("%v", data)
	bs, err := json.Marshal(data)
	if err != nil {
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)
}

// LobbiesGetUsers gets all the user names from the lobby specified at the end
// of the caught URL, and returns them as a JSON
func LobbiesGetUsers(w http.ResponseWriter, r *http.Request) {
	// Get lobby ID from URL
	vars := mux.Vars(r)
	id, _ := strconv.ParseUint(vars["id"], 10, 64)

	l, ok := lobby.GetLobby(uint64(id))
	if !ok {
		// TODO: lobby no longer exists, perhaps refresh page and error popup
		return
	}
	// Get all usernames from lobby
	guestCount := len(l.Guests.List)
	usernames := make([]string, 0, guestCount)
	firstUser := guestCount - l.Guests.UserSize
	for _, g := range l.Guests.List[firstUser:] {
		usernames = append(usernames, g.GetName())
	}
	// Convert to JSON and return it
	js, err := json.Marshal(usernames)
	if err != nil {
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func LobbiesJoin(w http.ResponseWriter, r *http.Request) {
	// Get the id of the requested lobby
	id, err := strconv.ParseUint(r.PostFormValue("id"), 10, 64)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Get the lobby from the global list
	l, ok := lobby.GetLobby(id)
	if !ok {
		log.Printf("Join: No Lobby (%d)\n", id)
		http.Error(w, "Lobby doesn't exist", http.StatusBadRequest)
		return
	}

	// Get the username of the authenicated user
	username, err := GetUserName(r)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Add the user to the lobby
	l.AddUser(username)

	l.Start()

	// Redirect user to the lobby
	log.Printf("User: %s Joined Lobby: %d", username, id)
	rawurl := fmt.Sprintf("http://%s/lobbies/%d", r.Host, id)
	http.Redirect(w, r, rawurl, http.StatusFound)
}

func LobbiesCreate(config *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := html.EscapeString(r.PostFormValue("name"))
		l := lobby.NewLobby(name, config)

		log.Printf("Created Lobby: %s", l.Name)
		http.Redirect(w, r, "/lobbies", http.StatusFound)
	})
}

func TrainingJoin(config *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := lobby.NewTrainingLobby(config)
		log.Printf("Created Training lobby")

		// Get the lobby from the global list
		//if !ok {
		//	log.Printf("Issue with joining the training lobby")
		//	http.Error(w, "Lobby doesn't exist", http.StatusBadRequest)
		//	return
		//}

		// Get the username of the authenicated user
		username, err := GetUserName(r)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Something went wrong fam", http.StatusBadRequest)
			return
		}

		// Add the trainee to the lobby
		l.AddTrainee(username)

		l.StartTrainingLobby()

		// Redirect user to the lobby
		log.Printf("User: %s Joined Training lobby", username)
		rawurl := fmt.Sprintf("http://%s/lobbies/%d", r.Host, 0)
		http.Redirect(w, r, rawurl, http.StatusFound)
	})
}

func LobbiesGame(staticDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, _ := strconv.ParseUint(vars["id"], 10, 64)
		username, _ := GetUserName(r)
		var (
			l  *lobby.Lobby
			ok bool
		)

		if id == 0 {
			l, ok = lobby.GetTrainingLobby(username)
		} else {
			l, ok = lobby.GetLobby(id)

		}
		if !ok || !l.ContainsUser(username) {
			u := fmt.Sprintf("http://%s/lobbies", r.Host)
			http.Redirect(w, r, u, http.StatusFound)
			return
		}

		http.ServeFile(w, r, staticDir+"/game-view.html")
	})
}

// LobbiesLeave temporarily logs the user out - this will change in the future
func LobbiesLeave(w http.ResponseWriter, r *http.Request) {
	// Get the id of the requested lobby
	id, err := strconv.ParseUint(r.PostFormValue("id"), 10, 64)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Get the lobby from the global list
	l, ok := lobby.GetLobby(id)
	if !ok {
		log.Printf("Join: No Lobby (%d)\n", id)
		http.Error(w, "Lobby doesn't exist", http.StatusBadRequest)
		return
	}

	// Get the username of the authenicated user
	username, err := GetUserName(r)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Remove the user to the lobby
	l.RemoveUser(username)

	// Redirect user to the lobby
	log.Printf("User: %s Left Lobby: %d", username, id)
	http.Redirect(w, r, "/", http.StatusFound)
}
