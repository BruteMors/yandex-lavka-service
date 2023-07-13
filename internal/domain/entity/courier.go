package entity

type Courier struct {
	CourierType  string
	Regions      []int
	WorkingHours []string
}

type CreateCourierDTO struct {
	CourierType  string   `json:"courier_type" validate:"required,oneof=FOOT BIKE AUTO"`
	Regions      []int    `json:"regions" validate:"required,gt=0,dive,gt=0,numeric"`
	WorkingHours []string `json:"working_hours" validate:"required,gt=0,dive,time_interval_format"`
}

type CourierDto struct {
	CourierId    int64    `db:"courier_id" json:"courier_id"`
	CourierType  string   `db:"courier_type" json:"courier_type"`
	Regions      []int    `db:"regions" json:"regions"`
	WorkingHours []string `db:"working_hours" json:"working_hours"`
}

type CreateCourierRequest struct {
	Couriers []CreateCourierDTO `json:"couriers" validate:"required,dive"`
}

type CourierMetaInfo struct {
	CourierId    int64    `db:"courier_id" json:"courier_id"`
	CourierType  string   `db:"courier_type" json:"courier_type"`
	Regions      []int    `db:"regions" json:"regions"`
	WorkingHours []string `db:"working_hours" json:"working_hours"`
	Rating       int      `db:"rating" json:"rating"`
	Earnings     int      `db:"earnings" json:"earnings"`
}
