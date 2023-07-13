package entity

type Order struct {
	Weight        float32
	Regions       int
	DeliveryHours []string
	Cost          int
	CompletedTime string
}

type CreateOrderDTO struct {
	Weight        float32  `json:"weight" validate:"required,gt=0,numeric"`
	Regions       int      `json:"regions" validate:"required,gt=0,numeric"`
	DeliveryHours []string `json:"delivery_hours" validate:"required,gt=0,dive,time_interval_format"`
	Cost          int      `json:"cost" validate:"required,gte=0,numeric"`
}

type OrderDto struct {
	OrderId       int64    `db:"order_id" json:"order_id"`
	Weight        float32  `db:"weight" json:"weight"`
	Regions       int      `db:"region" json:"regions"`
	DeliveryHours []string `db:"delivery_hours" json:"delivery_hours"`
	Cost          int      `db:"cost" json:"cost"`
	CompletedTime string   `db:"completed_time" json:"completed_time"`
}

type CompleteOrder struct {
	CourierId    int64  `db:"courier_id" json:"courier_id" validate:"required,gt=0,numeric"`
	OrderId      int64  `db:"order_id" json:"order_id" validate:"required,gt=0,numeric"`
	CompleteTime string `db:"completed_time" json:"complete_time" validate:"required,datetime=2006-01-02T15:04:05.999Z"`
}

type CreateOrderRequest struct {
	Orders []CreateOrderDTO `json:"orders" validate:"required,dive"`
}

type CompleteOrderRequestDTO struct {
	CompleteInfo []CompleteOrder `json:"complete_info" validate:"required,dive"`
}
