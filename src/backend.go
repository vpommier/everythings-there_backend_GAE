package backend

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sort"
	"strconv"

	"appengine"
	"appengine/datastore"
)

var resultsTypes = [3]string{"pending", "ongoing", "finished"}

//Données du Jeu
var nbPlaqueUtilise int = 6
var plaquesPossibles = [14]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 25, 50, 75, 100}
var minTotal int = 100
var maxTotal int = 999

type jeu struct {
	Plaques []int
	Total   int
}

type inputs struct {
	Jeu  jeu
	Type string
	Min  int
	Max  int
}

type solution struct {
	Jeu            jeu
	TempsExecution int
	NbOperations   int
	Resultats      []string
}

func contains(str string, ts []string) bool {
	for _, s := range ts {
		if str == s {
			return true
		}
	}
	return false
}

func genKey(j jeu) string {
	sort.Ints(j.Plaques)
	var key string
	for i, p := range j.Plaques {
		key += strconv.Itoa(int(p))
		if i < len(j.Plaques)-1 {
			key += ","
		}
	}
	key += ","
	key += strconv.Itoa(int(j.Total))

	return key
}

func (j *jeu) checkJeu() bool {
	if len(j.Plaques) != nbPlaqueUtilise {
		return false
	}

	for _, p := range j.Plaques {
		b := false
		for _, pp := range plaquesPossibles {
			if p == pp {
				b = true
				break
			}
		}
		if b == false {
			return false
		}
	}

	if minTotal > j.Total || j.Total > maxTotal {
		return false
	}

	return true
}

func results(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)

	//Reception de la requete
	var i inputs
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		context.Errorf("Requête invalide: %s", err)
		return
	}

	//Vérification du type de la requête
	if contains(i.Type, resultsTypes[:]) == false {
		w.WriteHeader(http.StatusBadRequest)
		context.Errorf("Incorrect type: %s", i.Type)
		return
	}

	//Composition et execution de la requête
	var res interface{}

	if !reflect.DeepEqual(i.Jeu, jeu{}) {
		if i.Jeu.checkJeu() == false {
			context.Errorf("Données du jeu invalide: Plaques: %d, Total: %d", i.Jeu.Plaques, i.Jeu.Total)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		key := datastore.NewKey(context, i.Type, genKey(i.Jeu), 0, nil)

		if i.Type == "pending" || i.Type == "ongoing" {
			res = new(jeu)
		} else {
			res = new(solution)
		}

		if err := datastore.Get(context, key, res); err != nil {
			context.Errorf("%s", err)
		}
	} else {
		q := datastore.NewQuery(i.Type)
		if i.Min > 0 {
			q.Offset(i.Min)
		}

		if i.Max > 0 && i.Max > i.Min {
			q.Limit(i.Max)
		}

		if i.Type == "pending" || i.Type == "ongoing" {
			res = new([]jeu)
		} else {
			res = new([]solution)
		}

		if _, err := q.GetAll(context, res); err != nil {
			context.Errorf("%s", err)
		}
	}

	//Création du JSON en contenu dans la réponse
	js, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		context.Errorf("%s", err)
		return
	}

	//Renvoi de la réponse
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func init() {
	http.HandleFunc("/results", results)
}
