package rest

import (
	"apigateway/core"
	"apigateway/utils"
	"encoding/json"
	"log"
	"net/http"
)

// ContactInfoApisHandler handles the contact information rest APIs implementation
type ContactInfoApisHandler struct {
	app *core.Application
}

// NewContactInfoApisHandler creates new rest Handler instance for contact info functions
func NewContactInfoApisHandler(app *core.Application) ContactInfoApisHandler {
	return ContactInfoApisHandler{app: app}
}

// GetContactInfo returns the contact information of a person
// @Summary Returns the name, permanent and mailing addresses, phone number and emergency contact information for a person
// @Tags Client
// @ID ConatctInfo
// @Param id query string true "User ID"
// @Accept  json
// @Produce json
// @Success 200 {object} model.Person
// @Security RokwireAuth ExternakAuth
// @Router /person/contactinfo [get]
func (h ContactInfoApisHandler) GetContactInfo(w http.ResponseWriter, r *http.Request) {

	log.Printf("Beginning call for %s", r.URL)
	externalToken := r.Header.Get("External-Authorization")

	id := ""
	mode := "0"

	reqParams := utils.ConstructFilter(r)
	if reqParams != nil {
		for _, v := range reqParams.Items {
			switch v.Field {
			case "id":
				id = v.Value[0]
			case "mode":
				mode = v.Value[0]
			}
		}
	}

	if id == "123456789" {
		mode = "1"
	}

	if id == "" || id == "null" {
		log.Printf("Error: missing id parameter")
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	if externalToken == "" {
		log.Printf("Error: External access token not includeed for %s", id)
		http.Error(w, "Missing external access token", http.StatusBadRequest)
		return
	}

	person, statusCode, err := h.app.Services.GetContactInfo(id, externalToken, mode)
	if err != nil {
		log.Printf("Error getting contact information for %s: Server returned %d %s \n", id, statusCode, err.Error())
		switch statusCode {
		case 401:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case 403:
			http.Error(w, err.Error(), http.StatusForbidden)
		case 404:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	resAsJSON, err := json.Marshal(person)
	if err != nil {
		log.Printf("Error on marshalling contact information for %s: %s\n", id, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//w.WriteHeader(http.StatusOK)
	w.Write(resAsJSON)
}
