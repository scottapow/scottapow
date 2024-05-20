package services

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type signupSuccessRepsponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func HandleSignup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signing up.")
	resp := &signupSuccessRepsponse{
		Success: true,
		Message: "you did it",
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respJson)
}
