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

	externalToken := r.Header.Get("External-Authorization")

	if externalToken == "" {
		log.Printf("Error: External access token not includeed")
		http.Error(w, "Missing external access token", http.StatusBadRequest)
	}

	reqParams := utils.ConstructFilter(r)
	id := ""
	for _, v := range reqParams.Items {
		if v.Field == "id" {
			id = v.Value[0]
			break
		}
	}

	if id == "" {
		log.Printf("Error: missing id parameter")
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	person, err := h.app.Services.GetContactInfo(id, externalToken)
	if err != nil {
		log.Printf("Error getting contact information: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resAsJSON, err := json.Marshal(person)
	if err != nil {
		log.Printf("Error on marshalling contact information: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resAsJSON)
}
