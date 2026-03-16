package orders

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

func (h *handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order repo.CreateOrderParams

	if err := json.Read(r, &order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdOrder, err := h.service.CreateOrder(r.Context(), order)

	if err != nil {
		log.Println(err.Error())
		return
	}

	resp := &CreatedOrder{
		ShippingCostKobo:    createdOrder.ShippingCostKobo,
		RawOrderPriceInKobo: createdOrder.RawOrderPriceInKobo,
		DiscountType:        createdOrder.DiscountType,
		DiscountValue:       createdOrder.DiscountValue,
	}
	json.Write(w, http.StatusAccepted, resp)
}
