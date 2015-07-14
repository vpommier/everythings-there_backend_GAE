package backend

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"appengine/aetest"
)

type testjeu struct {
	Plaques []interface{}
	Total   interface{}
}

type content struct {
	Jeu  testjeu
	Type interface{}
	Min  interface{}
	Max  interface{}
}

type cases struct {
	Description string
	Method      string
	Content     content
	Code        int
}

func TestResults(t *testing.T) {
	testCases := []cases{
		{
			Description: "Plaque interdite",
			Method:      "POST",
			Content: content{
				Jeu: testjeu{
					Plaques: []interface{}{1, 2, 3, 4, 5, 600},
					Total:   100,
				},
				Type: "pending",
			},
			Code: 400,
		},
		{
			Description: "Plaque interdite",
			Method:      "POST",
			Content: content{
				Jeu: testjeu{
					Plaques: []interface{}{"z", 2, 3, 4.55, "a", 600, 8},
					Total:   100,
				},
			},
			Code: 400,
		},
		{
			Description: "Trop de plaques",
			Method:      "POST",
			Content: content{
				Jeu: testjeu{
					Plaques: []interface{}{1, 2, 3, 4, 5, 6, 7, 8},
					Total:   100,
				},
				Type: "pending",
			},
			Code: 400,
		},
		{
			Description: "Total interdit",
			Method:      "POST",
			Content: content{
				Jeu: testjeu{
					Plaques: []interface{}{1, 2, 3, 4, 5, 6},
					Total:   1000,
				},
				Type: "pending",
			},
			Code: 400,
		},
		{
			Description: "Pas de type + Donnée de jeu OK",
			Method:      "POST",
			Content: content{
				Jeu: testjeu{
					Plaques: []interface{}{1, 2, 3, 4, 5, 6},
					Total:   100,
				},
			},
			Code: 400,
		},
		{
			Description: "Requête correcte (un seul résultat)",
			Method:      "POST",
			Content: content{
				Jeu: testjeu{
					Plaques: []interface{}{1, 2, 5, 6, 7, 100},
					Total:   100,
				},
				Type: "ongoing",
			},
			Code: 200,
		},
		{
			Description: "Juste le type",
			Method:      "POST",
			Content: content{
				Type: "pending",
			},
			Code: 200,
		},
		{
			Description: "Min > Max",
			Method:      "POST",
			Content: content{
				Type: "finished",
				Min:  10,
				Max:  0,
			},
			Code: 200,
		},
		{
			Description: "Requête correcte (résultats multiples)",
			Method:      "POST",
			Content: content{
				Type: "finished",
				Min:  0,
				Max:  10,
			},
			Code: 200,
		},
	}

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	for i, tt := range testCases {
		js, err := json.Marshal(tt.Content)
		if err != nil {
			t.Errorf("Failed to generate json body: %v", err)
		}

		req, err := inst.NewRequest(tt.Method, "/results", bytes.NewBuffer(js))
		if err != nil {
			t.Errorf("inst.NewRequest failed: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp := httptest.NewRecorder()
		results(resp, req)

		if resp.Code != tt.Code {
			t.Errorf("\nTest n°%d return %d insted %d\nBody: %v\n--", i+1, resp.Code, tt.Code, resp.Body)
			continue
		}
	}
}
