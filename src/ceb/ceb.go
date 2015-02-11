package ceb

import "strconv"

//Données du Jeu
var nbPlaqueUtilise int = 6
var plaquesPossibles = [14]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 25, 50, 75, 100}
var minTotal int = 100
var maxTotal int = 999

type Jeu struct {
	Plaques []int
	Total   int
}

const (
	MAX_PLAQUE_JEUX     = 6
	MAX_PLAQUE_RESULTAT = MAX_PLAQUE_JEUX
	MAX_PLAQUE          = MAX_PLAQUE_JEUX + MAX_PLAQUE_RESULTAT
)

/**
 * Type Operators
 */
type operators int

var str_operators = [...]string{
	"+",
	"-",
	"x",
	"/",
}

const (
	PLUS operators = iota
	MINUS
	TIMES
	DIVIDED
)

//end Type Operators

func calcul_de_2_plaques(i int, j int, op operators) int {
	if i > 0 && j > 0 {
		//		if i < j {
		//			i = i + j
		//			j = i - j
		//			i = i - j
		//		}
		switch op {
		case PLUS:
			return i + j
		case MINUS:
			return i - j
		case TIMES:
			return i * j
		case DIVIDED:
			if i%j == 0 && j > 1 {
				return i / j
			}
		}
	}
	return 0
}

/**
 * Tableau des résultats trouvé + son compteur
 */
var nb_resultats int
var tab_resultats = []string{}

//TODO: resultat aproximatif
//var result_finded int
var nb_operations int

//MAX_PLAQUE_RESULTAT
var tmp_resultats = [MAX_PLAQUE_RESULTAT]string{}

func add_operation(i int, j int, op operators) int {
	if i < j {
		i = i + j
		j = i - j
		i = i - j
	}
	r := calcul_de_2_plaques(i, j, op)
	tmp_resultats[nb_operations] = (strconv.Itoa(i) + " " + str_operators[op] + " " + strconv.Itoa(j) + " = " + strconv.Itoa(r))
	//	fmt.Println(tmp_resultats[nb_operations])
	nb_operations++
	return r
}

func del_operation() {
	tmp_resultats[nb_operations] = ""
	nb_operations--
}

func save_resultat() {
	var tmp string
	for i := 0; i < nb_operations; i++ {
		tmp += (tmp_resultats[i] + " \n")
	}
	tab_resultats = append(tab_resultats, tmp)
	nb_resultats++
}

//End tableau ds resultats

//Tableau des plaquettes + des résultats des operations
var total int

//MAX_PLAQUE
var tab_plaques = [MAX_PLAQUE]int{}

/**
 * tab_lock : tableau verrou des plaquettes déjà utilisé
 */
var tab_lock = [MAX_PLAQUE]bool{}

func lock_init() {
	for i := 0; i < MAX_PLAQUE; i++ {
		tab_lock[i] = false
	}
}

func check_unlock(j int) bool {
	return tab_lock[j] == false
}

func lock(j int) {
	tab_lock[j] = true
}

func unlock(j int) {
	tab_lock[j] = false
}

//End tab_lock

func resolver(max_tab int) {
	for i := 0; i < max_tab-1; i++ {
		for j := i + 1; j < max_tab; j++ {
			if check_unlock(i) && check_unlock(j) {
				lock(i)
				lock(j)
				for op := PLUS; op <= DIVIDED; op++ {
					tab_plaques[max_tab] = add_operation(tab_plaques[i], tab_plaques[j], op)
					if tab_plaques[max_tab] == total {
						save_resultat()
					} else if tab_plaques[max_tab] != 0 {
						//						fmt.Println(max_tab, " =>", tab_plaques[max_tab])
						resolver(max_tab + 1)
					}
					del_operation()
				}
				unlock(i)
				unlock(j)
			}
		}
	}
}

func (j *Jeu) Resolv() string {
	tab_plaques[0] = j.Plaques[0]
	tab_plaques[1] = j.Plaques[1]
	tab_plaques[2] = j.Plaques[2]
	tab_plaques[3] = j.Plaques[3]
	tab_plaques[4] = j.Plaques[4]
	tab_plaques[5] = j.Plaques[5]
	total = j.Total
	lock_init()
	nb_operations = 0
	nb_resultats = 0
	resolver(MAX_PLAQUE_JEUX)
	return tab_resultats[0]
}

func (j *Jeu) CheckJeu() bool {
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

/*
func main() {
	var j Jeu
	j.Plaques = append(j.Plaques, 100)
        j.Plaques = append(j.Plaques, 2)
        j.Plaques = append(j.Plaques, 75)
        j.Plaques = append(j.Plaques, 3)
        j.Plaques = append(j.Plaques, 1)
        j.Plaques = append(j.Plaques, 10)
	j.Total = 888
	t := j.Resolv()
	fmt.Println(t)
	fmt.Println(j)
}
*/
