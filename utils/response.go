package utils


import (
	"net/http"
	"encoding/json"
)

type ApiResponse struct{
	Sucess bool `json:"success"`
	ERROR string `json:"error"`
	Message string `json:"message"`
	Data interface{} `json:"data,omitempty"`
}


func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	

	response := ApiResponse{
		Sucess: true,
		Data: payload,
		//Data: details,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func RespondWithError(w http.ResponseWriter, code int, message string, details interface{}){
	w.Header().Set("Content-Type", "application/json")
w.WriteHeader(code)

json.NewEncoder(w).Encode(ApiResponse{
	Sucess: false,
	ERROR: message,
	Data: details,
})
}