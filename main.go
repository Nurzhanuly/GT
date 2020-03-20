package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// var tpl *template.Template

// func init() {
// 	tpl = template.Must(template.ParseGlob("static/templates/*"))
// }

// URls is a struct all API
type URls struct {
	Artists   string `json:"artists"`
	Locations string `json:"locations"`
	Dates     string `json:"dates"`
	Relation  string `json:"relation"`
}

var all *URls

// ArtistJSON is API Artists
type ArtistJSON struct {
	ID            int      `json:"id"`
	Image         string   `json:"image"`
	Name          string   `json:"name"`
	Members       []string `json:"members"`
	CreationDate  int      `json:"creationDate"`
	FirstAlbum    string   `json:"firstAlbum"`
	Relations     string   `json:"relations"`
	RelationsData RelationJSON
}

var artists []ArtistJSON

// RelationJSON is API RelationData
type RelationJSON struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

var rJSON []RelationJSON

func main() {
	PORT := "8888"

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handler)
	http.HandleFunc("/artist", artist)
	http.HandleFunc("/search", search)
	log.Println("Listening to server on", PORT)
	http.ListenAndServe(":"+PORT, nil)
}

func search(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("search")
	var searchResult []ArtistJSON
	input := r.FormValue("search-bar")

	for _, v := range artists {
		if strings.ToLower(v.Name) == strings.ToLower(input) {
			searchResult = append(searchResult, v)
		} else {
			for _, ch := range v.Members {
				if strings.ToLower(ch) == strings.ToLower(input) {
					searchResult = append(searchResult, v)
				}
			}
		}

		if strings.ToLower(strconv.Itoa(v.CreationDate)) == strings.ToLower(input) {
			searchResult = append(searchResult, v)
		}

		if strings.ToLower(v.FirstAlbum) == strings.ToLower(input) {
			searchResult = append(searchResult, v)
		}

		for key, value := range v.RelationsData.DatesLocations {
			if key == input {
				searchResult = append(searchResult, v)
			}
			for _, i := range value {
				if i == input {
					searchResult = append(searchResult, v)
				}
			}
		}
	}
	if len(searchResult) == 0 {
		errHandler(w, http.StatusBadRequest)
	} else {
		handleSearch(w, &searchResult)
	}

}

func handleSearch(w http.ResponseWriter, r *[]ArtistJSON) {
	tpl, err := template.ParseFiles("static/templates/search.html")
	if err != nil {
		fmt.Println(err)
		errHandler(w, http.StatusInternalServerError)
	}
	tpl.Execute(w, r)
}

func artist(w http.ResponseWriter, r *http.Request) {
	btnArtist := r.FormValue("artist-btn")
	btnSearch := r.FormValue("search-btn")
	found := false

	if btnArtist == "" && btnSearch == "" {
		rand.Seed(time.Now().UnixNano())
		min := 2
		max := 53
		tmp := rand.Intn(max-min+1) + min
		btnArtist = artists[tmp-1].Name
	}

	if btnSearch == "" && btnArtist != "" {
		for _, v := range artists {
			if v.Name == btnArtist {
				handleArtists(w, &v)
				found = true
			}
		}
	}
	if !found {
		errHandler(w, http.StatusBadRequest)
	}
}
func handleArtists(w http.ResponseWriter, v *ArtistJSON) {
	tpl, err := template.ParseFiles("static/templates/artist.html")
	if err != nil {
		errHandler(w, http.StatusInternalServerError)
	}
	tpl.Execute(w, v)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		errHandler(w, http.StatusNotFound)
		// fmt.Println("0")
		// return
	} else {
		url := "https://groupietrackers.herokuapp.com/api"
		data, err := getAPI(url)
		if err != nil {
			fmt.Println("1")
			errHandler(w, http.StatusInternalServerError)
			return
		}
		json.Unmarshal(data, &all)
		parseHandler(w)
	}
}

func parseHandler(w http.ResponseWriter) {
	data, err := getAPI(all.Artists)
	if err != nil {
		errHandler(w, http.StatusInternalServerError)
	}
	json.Unmarshal(data, &artists)

	relData, err := getAPI(all.Relation)
	if err != nil {
		errHandler(w, http.StatusInternalServerError)
	} else {
		relData = relData[9 : len(relData)-2]
		json.Unmarshal(relData, &rJSON)
		// tpl.ExecuteTemplate(w, "assets/templates/index.html", nil)
		for i := range artists {
			artists[i].RelationsData = rJSON[i]
		}

		tpl, errParse := template.ParseFiles("static/templates/index.html")
		if errParse != nil {
			errHandler(w, http.StatusInternalServerError)
		} else {
			tpl.Execute(w, &artists)
		}
	}
}

func getAPI(url string) ([]byte, error) {

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		// log.Fatal(err)
		return nil, err
	}
	defer res.Body.Close()
	return body, nil
}

func errHandler(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	if status == http.StatusBadRequest {
		statusBadRequest(w)
	} else if status == http.StatusNotFound {
		statusNotFound(w)
	} else if status == http.StatusInternalServerError {
		statusInternalServerError(w)
	}
}

func statusInternalServerError(w http.ResponseWriter) {
	// tpl.ExecuteTemplate(w, "static/templates,error/500.html", nil)
	tpl, err := template.ParseFiles("static/templates/error/500.html")
	if err != nil {
		log.Fatal(err)
	}
	tpl.Execute(w, nil)
}

func statusNotFound(w http.ResponseWriter) {
	// tpl.ExecuteTemplate(w, "static/templates,error/404.html", nil)
	tpl, err := template.ParseFiles("static/templates/error/404.html")
	if err != nil {
		log.Fatal(err)
	}
	tpl.Execute(w, nil)
}

func statusBadRequest(w http.ResponseWriter) {
	// tpl, err := template.ParseFiles("")
	// tpl.ExecuteTemplate(w, "static/templates,error/400.html", nil)
	tpl, err := template.ParseFiles("static/templates/error/400.html")
	if err != nil {
		log.Fatal(err)
	}
	tpl.Execute(w, nil)
}
