package products

import (
	"ecom/cmd/helpers"
	repo "ecom/internal/adapters/postgresql/sqlc"
	"ecom/internal/json"
	"log"
	"net/http"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product repo.CreateProductParams
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}
	if err := json.Read(r, &product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdProduct, err := h.service.CreateProduct(r.Context(), product)

	if err != nil {
		log.Println(err.Error())
		return
	}

	resp := &CreatedProduct{
		Name:        createdProduct.Name,
		Description: helpers.TextToString(createdProduct.Description),
	}

	json.Write(w, http.StatusAccepted, resp)

}
