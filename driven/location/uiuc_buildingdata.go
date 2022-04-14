package buildinglocation

type entrance struct {
	UUID         *string `json:"uuid"`
	Name         *string `json:"descriptive_name"`
	ADACompliant bool    `json:"is_ada_compliant"`
	Available    bool    `json:"is_available_for_use"`
	ImageURL     *string `json:"image"`
	Latitude     float32 `json:"latitude"`
	Longitude    float32 `json:"longitude"`
}

type building struct {
	UUID        *string `json:"uuid"`
	Name        *string `json:"name"`
	Number      *string `json:"number"`
	FullAddress *string `json:"location"`
	Address1    *string `json:"address_1"`
	Address2    *string `json:"address_2"`
	City        *string `json:"city"`
	State       *string `json:"state"`
	ZipCode     *string `json:"zipcode"`
	ImageURL    *string `json:"image"`
	MailCode    *string `json:"mailcode"`
}

type serverResponse struct {
	Status         *string `json:"status"`
	HttpStatusCode int     `json:"http_return"`
	CollectionType *string `json:"collection"`
	Count          int     `json:"count"`
	ErrorList      *string `json:"errors"`
	ErrorMessage   *string `json:"error_text"`
}
