package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/fabwi987/irec-front/callback"
	"github.com/fabwi987/irec-front/session"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
)

type Position struct {
	Idpositions uuid.UUID `json:"idpositions" bson:"idpositions"`
	Iduser      string    `json:"iduser" bson:"iduser"`
	Title       string    `json:"title" bson:"title"`
	Subtitle    string    `json:"subtitle" bson:"subtitle"`
	Text        string    `json:"text" bson:"text"`
	Enddate     time.Time `json:"enddate" bson:"enddate"`
	Reward      string    `json:"reward" bson:"reward"`
	Created     time.Time `json:"created" bson:"created"`
	Lastupdated time.Time `json:"lastupdated" bson:"lastupdated"`
	HREF        string    `json:"href" bson:"href"`
	Meta        string    `json:"meta" bson:"meta"`
}

var Client http.Client

func main() {

	err := session.Init()
	if err != nil {
		log.Fatal("Could not create session")
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := mux.NewRouter()

	r.HandleFunc("/start", LoginHandler)
	r.HandleFunc("/callback", callback.CallbackHandler)
	r.HandleFunc("/unauth", UnauthHandler)

	r.Handle("/user", negroni.New(
		negroni.HandlerFunc(IsAuthenticated),
		negroni.Wrap(http.HandlerFunc(UserHandler)),
	))

	r.Handle("/positions", negroni.New(
		negroni.HandlerFunc(IsAuthenticated),
		negroni.Wrap(http.HandlerFunc(PositionHandler)),
	))

	http.ListenAndServe(os.Getenv("APP_PORT"), handlers.LoggingHandler(os.Stdout, r))

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	t, _ := template.ParseFiles("views/login.html")
	t.Execute(w, nil)

}

func UnauthHandler(w http.ResponseWriter, r *http.Request) {

	t, _ := template.ParseFiles("views/unauth.html")
	t.Execute(w, nil)

}

func UserHandler(w http.ResponseWriter, r *http.Request) {

	session, err := session.Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, _ := template.ParseFiles("views/user.html")
	//t.Execute(w, nil)
	t.Execute(w, session.Values["profile"])

}

func PositionHandler(w http.ResponseWriter, r *http.Request) {

	session, err := session.Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("GET", os.Getenv("API_URL")+"/positions", nil)
	req.Header.Add("Authorization", "Bearer "+session.Values["id_token"].(string))
	resp, err := Client.Do(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	var pos []*Position

	if err := json.NewDecoder(resp.Body).Decode(&pos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	layoutData := struct {
		ThreadID int
		Posts    []*Position
	}{
		ThreadID: 1,
		Posts:    pos,
	}

	t, _ := template.ParseFiles("views/positions.html")
	t.Execute(w, layoutData)

}

func IsAuthenticated(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	session, err := session.Store.Get(r, "auth-session")
	if err != nil {
		log.Println("No session found")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := session.Values["id_token"]; !ok {
		http.Redirect(w, r, "/unauth", http.StatusSeeOther)
	} else {
		next(w, r)
	}
}
