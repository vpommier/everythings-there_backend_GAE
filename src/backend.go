package backend

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"api_compte_est_bon/src/ceb"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
)

// Etats du traitement d'un jeu.
const (
	PENDING  string = "pending"
	ONGOING  string = "ongoing"
	FINISHED string = "finished"
)

// Namespace par défaut.
const (
	DEFAULT_CONTEXT string = "epsi"
)

var resultsTypes = [3]string{PENDING, ONGOING, FINISHED}

type inputs struct {
	Jeu  ceb.Jeu
	Type string
	Min  int
	Max  int
}

type solution struct {
	Jeu            ceb.Jeu
	TempsExecution int
	NbOperations   int
	Resultats      string
}

func contains(str string, ts []string) bool {
	for _, s := range ts {
		if str == s {
			return true
		}
	}
	return false
}

func genStringID(j ceb.Jeu) string {
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

func setNamespace(context appengine.Context, r *http.Request) (c appengine.Context, err error) {
	if n := r.Header.Get("User"); n != "" {
		if c, err = appengine.Namespace(context, n); err != nil {
			return nil, err
		}
	} else {
		if c, err = appengine.Namespace(context, DEFAULT_CONTEXT); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func getParamsJeu(r *http.Request) (j ceb.Jeu, err error) {
	t := strings.Split(r.FormValue("Plaques"), ",")
	for _, p := range t {
		i, err := strconv.Atoi(p)
		if err != nil {
			return ceb.Jeu{}, err
		}
		j.Plaques = append(j.Plaques, i)
	}
	j.Total, err = strconv.Atoi(r.FormValue("Total"))
	if err != nil {
		return ceb.Jeu{}, err
	}
	return j, nil
}

func demand(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)

	context, err := setNamespace(context, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		context.Errorf("%s", err)
		return
	}

	//Reception de la requete
	var j ceb.Jeu
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		context.Errorf("%s", err)
		return
	}

	if j.CheckJeu() == false {
		context.Errorf("Données du jeu invalide: Plaques: %d, Total: %d", j.Plaques, j.Total)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exist := false
	stringID := genStringID(j)
	var res interface{}

	for _, t := range resultsTypes {
		if t == PENDING || t == ONGOING {
			res = new(ceb.Jeu)
		} else {
			res = new(solution)
		}

		key := datastore.NewKey(context, t, stringID, 0, nil)
		if err := datastore.Get(context, key, res); err != nil {
			if err == datastore.ErrNoSuchEntity {
				continue
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				context.Errorf("%s", err)
				return
			}
		} else {
			exist = true
		}
	}

	// Insertion
	if exist == false {
		err := datastore.RunInTransaction(context, func(c appengine.Context) error {
			var err error

			key := datastore.NewKey(c, PENDING, stringID, 0, nil)
			if _, err := datastore.Put(c, key, &j); err != nil {
				return err
			}

			params := url.Values{}

			var t []string
			for _, p := range j.Plaques {
				t = append(t, strconv.Itoa(p))
			}

			params.Set("Plaques", strings.Join(t, ","))
			params.Set("Total", strconv.Itoa(j.Total))

			task := taskqueue.NewPOSTTask("/solve", params)
			if _, err = taskqueue.Add(c, task, ""); err != nil {
				return err
			}

			return err
		}, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			context.Errorf("%s", err)
			return
		}
	}
}

func solve(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)

	j, err := getParamsJeu(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		context.Errorf("%s", err)
		return
	}

	if j.CheckJeu() == false {
		context.Errorf("Données du jeu invalide: Plaques: %d, Total: %d", j.Plaques, j.Total)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Generation de la Key
	k := genStringID(j)
	//Netoyage de pending
	err = datastore.RunInTransaction(context, func(c appengine.Context) error {
		var err error
		key := datastore.NewKey(c, PENDING, k, 0, nil)
		if err = datastore.Delete(c, key); err != nil {
			return err
		}
		//Ajout dans ongoing
		key = datastore.NewKey(c, ONGOING, k, 0, nil)
		if _, err = datastore.Put(c, key, &j); err != nil {
			return err
		}

		//Résolution du compte est bon
		var sol solution
		sol.Resultats = j.Resolv()
		sol.Jeu = j

		//Netoyage de ongoing
		key = datastore.NewKey(c, ONGOING, k, 0, nil)
		if err = datastore.Delete(c, key); err != nil {
			return err
		}

		//Ajout dans finised
		key = datastore.NewKey(c, FINISHED, k, 0, nil)
		if _, err = datastore.Put(c, key, &sol); err != nil {
			return err
		}

		return err
	}, &datastore.TransactionOptions{XG: true})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		context.Errorf("%s", err)
		return
	}
	return
}

func results(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)

	context, err := setNamespace(context, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		context.Errorf("%s", err)
		return
	}

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

	if !reflect.DeepEqual(i.Jeu, ceb.Jeu{}) {
		if i.Jeu.CheckJeu() == false {
			context.Errorf("Données du jeu invalide: Plaques: %d, Total: %d", i.Jeu.Plaques, i.Jeu.Total)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		key := datastore.NewKey(context, i.Type, genStringID(i.Jeu), 0, nil)

		if i.Type == PENDING || i.Type == ONGOING {
			res = new(ceb.Jeu)
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

		if i.Type == PENDING || i.Type == ONGOING {
			res = new([]ceb.Jeu)
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
	http.HandleFunc("/demand", demand)
	http.HandleFunc("/results", results)
	http.HandleFunc("/solve", solve)
}
