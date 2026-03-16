package categories

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

func (h *handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category repo.Category
	name := category.Name
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusBadRequest)
	}
	if err := json.Read(r, &category); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	createdCategory, err := h.service.CreateCategory(r.Context(), name)
	if err != nil {
		log.Println(err.Error())
	}
	resp := CreatedCategory{
		Name: createdCategory.Name,
	}

	json.Write(w, http.StatusAccepted, resp)
}
