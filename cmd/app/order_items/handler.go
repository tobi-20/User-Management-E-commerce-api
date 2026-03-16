package order_items

import (
	repo "ecom/internal/adapters/postgresql/sqlc"
	"ecom/internal/json"
	"log"
	"net/http"
)

type handler struct {
	service Service
}

func (h *handler) CreateOrderItem(w http.ResponseWriter, r *http.Request) {

	var orderItem repo.CreateOrderItemParams
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusBadRequest)
	}

	if err := json.Read(r, &orderItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}

	createdItem, err := h.service.CreateOrderItems(r.Context(), orderItem)
	if err != nil {
		log.Println(err.Error())
	}

	resp := &CreatedOrderItemResponse{
		Quantity:      createdItem.Quantity,
		PriceInKobo:   createdItem.PriceInKobo,
		DiscountType:  createdItem.DiscountType,
		DiscountValue: createdItem.DiscountValue,
		ItemTotal:     createdItem.ItemTotal,
	}

	json.Write(w, http.StatusAccepted, resp)
}
