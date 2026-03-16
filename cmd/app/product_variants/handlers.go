package product_variants

import (
	repo "ecom/internal/adapters/postgresql/sqlc"
	"ecom/internal/json"

	"log"
	"net/http"
)

type handler struct {
	service Service
}

func (h *handler) CreateProductVariant(w http.ResponseWriter, r *http.Request) {
	var variant repo.CreateProductVariantParams
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusBadRequest)
	}

	if err := json.Read(r, &variant); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	createdVariant, err := h.service.CreateProductVariant(r.Context(), variant)
	if err != nil {
		log.Println(err.Error())
	}

	resp := &VariantResp{
		Weight:      createdVariant.Weight,
		Unit:        createdVariant.Unit,
		PriceInKobo: createdVariant.PriceInKobo,
		Stock:       createdVariant.Stock,
	}

	json.Write(w, http.StatusAccepted, resp)

}
