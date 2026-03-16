package brand

import (
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

func (h *handler) CreateBrand(w http.ResponseWriter, r *http.Request) {
	var brand repo.Brand
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
	if err := json.Read(r, &brand); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	createdBrand, err := h.service.CreateBrand(r.Context(), brand.Name)
	if err != nil {
		log.Println(err)
	}
	resp := BrandResponse{
		Name: createdBrand.Name,
	}
	json.Write(w, http.StatusAccepted, resp.Name)
}
