package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/fabwi987/irec-front/callback"
	"github.com/fabwi987/irec-front/session"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
)

type Recommendation struct {
	Idrecommendations uuid.UUID `json:"idrecommendations" bson:"idrecommendations"`
	User              *User     `json:"user" bson:"user"`
	Referral          *Referral `json:"referral" bson:"referral"`
	Position          *Position `json:"position" bson:"position"`
	Text              string    `json:"text" bson:"text"`
	Confirmed         bool      `json:"confirmed" bson:"confirmed"`
	Created           time.Time `json:"created" bson:"created"`
	Lastupdated       time.Time `json:"lastupdated" bson:"lastupdated"`
	HREF              string    `json:"href" bson:"href"`
	Meta              string    `json:"meta" bson:"meta"`
}

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

type User struct {
	IdUser      string    `json:"iduser" bson:"iduser"`
	UserType    int       `json:"usertype" bson:"usertype"`
	Name        string    `json:"name" bson:"name"`
	Telephone   string    `json:"telephone" bson:"telephone"`
	Mail        string    `json:"mail" bson:"mail"`
	Picture     string    `json:"picture" bson:"picture"`
	Headline    string    `json:"headline" bson:"headline"`
	ProfileURL  string    `json:"profileURL" bson:"profileURL"`
	Created     time.Time `json:"created" bson:"created"`
	Lastupdated time.Time `json:"lastupdated" bson:"lastupdated"`
	HREF        string    `json:"href" bson:"href"`
	Meta        string    `json:"meta" bson:"meta"`
}

type Referral struct {
	Idreferrals    uuid.UUID `json:"idreferral" bson:"idreferral"`
	ReferralUserID string    `json:"referraluserid" bson:"referraluserid"`
	Name           string    `json:"name" bson:"name"`
	Telephone      string    `json:"telephone" bson:"telephone"`
	Mail           string    `json:"mail" bson:"mail"`
	Picture        string    `json:"picture" bson:"picture"`
	Headline       string    `json:"headline" bson:"headline"`
	ProfileURL     string    `json:"profileURL" bson:"profileURL"`
	Created        time.Time `json:"created" bson:"created"`
	Lastupdated    time.Time `json:"lastupdated" bson:"lastupdated"`
	HREF           string    `json:"href" bson:"href"`
	Meta           string    `json:"meta" bson:"meta"`
}

type Error struct {
	Time    time.Time `json:"Time"`
	Message string    `json:"Message"`
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

	r.Handle("/user/single/{id}", negroni.New(
		negroni.HandlerFunc(IsAuthenticated),
		negroni.Wrap(http.HandlerFunc(UserUpdate)),
	))

	r.Handle("/usercontrol", negroni.New(
		negroni.HandlerFunc(IsAuthenticated),
		negroni.Wrap(http.HandlerFunc(UserControl)),
	))

	r.Handle("/positions", negroni.New(
		negroni.HandlerFunc(IsAuthenticated),
		negroni.Wrap(http.HandlerFunc(PositionHandler)),
	))

	r.Handle("/positions/single/{id}", negroni.New(
		negroni.HandlerFunc(IsAuthenticated),
		negroni.Wrap(http.HandlerFunc(SinglePositionHandler)),
	))

	r.Handle("/recommendation/{id}", negroni.New(
		negroni.HandlerFunc(IsAuthenticated),
		negroni.Wrap(http.HandlerFunc(RecommendationHandler)),
	))

	r.Handle("/recommendations/positions/{id}", negroni.New(
		negroni.HandlerFunc(IsAuthenticated),
		negroni.Wrap(http.HandlerFunc(RecommendationPositionHandler)),
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
		Name     string
		Picture  string
		Posts    []*Position
	}{
		ThreadID: 1,
		Name:     session.Values["name"].(string),
		Picture:  session.Values["picture"].(string),
		Posts:    pos,
	}

	t, _ := template.ParseFiles("views/positions_v1.html")
	t.Execute(w, layoutData)

}

func SinglePositionHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]
	session, err := session.Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("GET", os.Getenv("API_URL")+"/positions/single/"+id, nil)
	req.Header.Add("Authorization", "Bearer "+session.Values["id_token"].(string))
	resp, err := Client.Do(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	var pos *Position

	if err := json.NewDecoder(resp.Body).Decode(&pos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, _ := template.ParseFiles("views/singlepos.html")
	t.Execute(w, pos)

}

func RecommendationHandler(w http.ResponseWriter, r *http.Request) {
	session, err := session.Store.Get(r, "auth-session")
	vars := mux.Vars(r)
	posid := vars["id"]

	log.Println(session)
	log.Println(posid)

	var rec Referral
	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		decoder := schema.NewDecoder()
		err = decoder.Decode(&rec, r.PostForm)
	}

	form := url.Values{}
	form.Add("Name", rec.Name)
	form.Add("Telephone", rec.Telephone)
	form.Add("Mail", rec.Mail)

	request_url := os.Getenv("API_URL") + "/recommendations/position/" + posid
	req, err := http.NewRequest("POST", request_url, strings.NewReader(form.Encode()))
	req.Header.Add("Authorization", "Bearer "+session.Values["id_token"].(string))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := Client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Println(resp)

	http.Redirect(w, r, "/positions", http.StatusSeeOther)

}

func RecommendationPositionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	session, err := session.Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("GET", os.Getenv("API_URL")+"/recommendations/all/position/"+id, nil)
	req.Header.Add("Authorization", "Bearer "+session.Values["id_token"].(string))
	resp, err := Client.Do(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	var rec []*Recommendation

	if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	layoutData := struct {
		ThreadID int
		Posts    []*Recommendation
	}{
		ThreadID: 1,
		Posts:    rec,
	}

	t, _ := template.ParseFiles("views/recommendations.html")
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

func UserControl(w http.ResponseWriter, r *http.Request) {
	session, err := session.Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("GET", os.Getenv("API_URL")+"/users/single/"+session.Values["UserID"].(string), nil)
	req.Header.Add("Authorization", "Bearer "+session.Values["id_token"].(string))
	resp, err := Client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	var usr *User
	if err := json.NewDecoder(resp.Body).Decode(&usr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if usr.IdUser == "" {
		layoutData := struct {
			ThreadID   int
			Name       string
			Picture    string
			Headline   string
			ProfileURL string
			UserID     string
		}{
			ThreadID:   1,
			Name:       session.Values["name"].(string),
			Picture:    session.Values["picture"].(string),
			Headline:   session.Values["Headline"].(string),
			ProfileURL: session.Values["ProfileURL"].(string),
			UserID:     session.Values["UserID"].(string),
		}

		t, _ := template.ParseFiles("views/userInfo.html")
		t.Execute(w, layoutData)
	} else {
		http.Redirect(w, r, "/positions", http.StatusSeeOther)
	}
}

func UserUpdate(w http.ResponseWriter, r *http.Request) {
	session, err := session.Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var usr User
	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		decoder := schema.NewDecoder()
		err = decoder.Decode(&usr, r.PostForm)
	}

	form := url.Values{}
	form.Add("IdUser", session.Values["UserID"].(string))
	form.Add("UserType", "1")
	form.Add("Name", usr.Name)
	form.Add("Telephone", usr.Telephone)
	form.Add("Mail", usr.Mail)
	form.Add("Headline", usr.Headline)
	form.Add("ProfileURL", usr.ProfileURL)
	form.Add("Picture", session.Values["picture"].(string))

	request_url := os.Getenv("API_URL") + "/users"
	req, err := http.NewRequest("POST", request_url, strings.NewReader(form.Encode()))
	req.Header.Add("Authorization", "Bearer "+session.Values["id_token"].(string))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := Client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Println(resp)

	http.Redirect(w, r, "/positions", http.StatusSeeOther)

}
