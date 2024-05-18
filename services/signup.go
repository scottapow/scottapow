package services

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type signupSuccessRepsponse struct {
	success bool
	message string
}

func HandleSignup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signing up.")
	resp := &signupSuccessRepsponse{
		success: true,
		message: "you did it",
	}
	respJson, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(respJson)
	}

}
