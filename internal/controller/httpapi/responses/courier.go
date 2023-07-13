package responses

import (
	"yandex-lavka-service/internal/domain/entity"
)

type CreateCouriersResponse struct {
	Couriers *[]entity.CourierDto `json:"couriers"`
}

type GetCourierMetaInfoResponse struct {
	entity.CourierMetaInfo
}
type GetCourierResponse struct {
	entity.CourierDto
}

type GetCouriersResponse struct {
	Couriers *[]entity.CourierDto `json:"couriers"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
}
