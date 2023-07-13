package responses

import (
	"yandex-lavka-service/internal/domain/entity"
)

type CreateOrdersResponse struct {
	Orders *[]entity.OrderDto `json:"orders"`
}

type GetOrderResponse struct {
	entity.OrderDto
}

type GetOrdersResponse struct {
	Orders *[]entity.OrderDto `json:"orders"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

type SetCompleteOrdersResponse struct {
	Orders *[]entity.OrderDto `json:""`
}
