package shipping_rules

import (
	repo "ecom/internal/adapters/postgresql/sqlc"
	"ecom/internal/json"
	"log"
	"net/http"
)

type handler struct {
	service Service
}

func (h *handler) CreateShippingRules(w http.ResponseWriter, r *http.Request) {
	var shippingRules repo.CreateShippingRulesParams
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusBadRequest)
	}
	if err := json.Read(r, &shippingRules); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	createdRules, err := h.service.CreateShippingRules(r.Context(), shippingRules)
	if err != nil {
		log.Println(err.Error())
	}
	resp := &CreateShippingRulesParams{
		MaxPriceInKobo: createdRules.MaxPriceInKobo,
		MinPriceInKobo: createdRules.MinPriceInKobo,
		Type:           createdRules.Type,
		Value:          createdRules.Value,
	}

	json.Write(w, http.StatusAccepted, resp)
}
