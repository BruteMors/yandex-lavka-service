package service

import (
	"github.com/labstack/gommon/log"
	"math"
	"time"
	"yandex-lavka-service/internal/config"
	"yandex-lavka-service/internal/domain/entity"
)

// CourierStore - interface type for database
type CourierStore interface {
	Add(couriers ...entity.Courier) (*[]entity.CourierDto, error)
	Get(id int64) (*entity.CourierDto, error)
	GetAll(limit int, offset int) (*[]entity.CourierDto, error)
	GetCosts(id int64, startDate string, endDate string) (costs []int, err error)
}

type Courier struct {
	store  CourierStore
	config *config.Config
}

// NewCourier - create Courier service instance
func NewCourier(store CourierStore, config *config.Config) *Courier {
	return &Courier{store: store, config: config}
}

// AddCouriers - add couriers to database
func (c *Courier) AddCouriers(couriers ...entity.Courier) (*[]entity.CourierDto, error) {
	log.Debug("Add couriers to db")
	addedCouriers, err := c.store.Add(couriers...)
	if err != nil {
		return nil, err
	}
	return addedCouriers, nil
}

// GetCourier - get courier from database
func (c *Courier) GetCourier(id int64) (*entity.CourierDto, error) {
	log.Debug("get courier from database")
	courier, err := c.store.Get(id)
	if err != nil {
		return nil, err
	}
	return courier, nil
}

// GetCouriers - get all couriers from database
func (c *Courier) GetCouriers(limit int, offset int) (*[]entity.CourierDto, error) {
	log.Debug("get couriers from database")
	couriers, err := c.store.GetAll(limit, offset)
	if err != nil {
		return nil, err
	}
	return couriers, nil
}

// GetCourierMetaInfo - get meta info about courier by courier id from database
func (c *Courier) GetCourierMetaInfo(courierId int64, startDate string, endDate string) (*entity.CourierMetaInfo, error) {
	log.Debug("get courier meta info from database")
	costs, err := c.store.GetCosts(courierId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	courier, err := c.store.Get(courierId)
	if err != nil {
		return nil, err
	}

	if len(costs) == 0 {
		courierMetaInfo := entity.CourierMetaInfo{
			CourierId:    courier.CourierId,
			CourierType:  courier.CourierType,
			Regions:      courier.Regions,
			WorkingHours: courier.WorkingHours,
			Rating:       0,
			Earnings:     0,
		}
		return &courierMetaInfo, nil
	}

	rating, err := c.calculateRating(len(costs), startDate, endDate, courier.CourierType)
	if err != nil {
		return nil, err
	}

	earnings := c.calculateEarnings(costs, courier.CourierType)

	courierMetaInfo := entity.CourierMetaInfo{
		CourierId:    courier.CourierId,
		CourierType:  courier.CourierType,
		Regions:      courier.Regions,
		WorkingHours: courier.WorkingHours,
		Rating:       rating,
		Earnings:     earnings,
	}

	return &courierMetaInfo, nil
}

// calculateEarnings - calculate courier's earnings. Input param: costs - array of orders costs, courierType - type of courier (FOOT, BIKE or AUTO)
func (c *Courier) calculateEarnings(costs []int, courierType string) int {
	var courierCostFactor int

	switch courierType {
	case c.config.AppConfig.CourierType.FootCourierType:
		courierCostFactor = c.config.AppConfig.CourierCostFactor.FootCourierCostFactor
	case c.config.AppConfig.CourierType.BikeCourierType:
		courierCostFactor = c.config.AppConfig.CourierCostFactor.BikeCourierCostFactor
	case c.config.AppConfig.CourierType.AutoCourierType:
		courierCostFactor = c.config.AppConfig.CourierCostFactor.AutoCourierCostFactor
	}

	var sum int
	for _, cost := range costs {
		sum += cost * courierCostFactor
	}
	return sum
}

// calculateRating - calculate courier's rating. Input param: numOfCompleteOrders - number of complete orders for period from startDate to endDate, courierType - type of courier (FOOT, BIKE or AUTO)
func (c *Courier) calculateRating(numOfCompleteOrders int, startDate string, endDate string, courierType string) (int, error) {
	var courierRateFactor int

	switch courierType {
	case c.config.AppConfig.CourierType.FootCourierType:
		courierRateFactor = c.config.AppConfig.CourierRateFactor.FootCourierRateFactor
	case c.config.AppConfig.CourierType.BikeCourierType:
		courierRateFactor = c.config.AppConfig.CourierRateFactor.BikeCourierRateFactor
	case c.config.AppConfig.CourierType.AutoCourierType:
		courierRateFactor = c.config.AppConfig.CourierRateFactor.AutoCourierRateFactor
	}

	startDateTime, err := time.Parse(time.DateOnly, startDate)
	if err != nil {
		return 0, err
	}

	endDateTime, err := time.Parse(time.DateOnly, endDate)
	if err != nil {
		return 0, err
	}

	hours := endDateTime.Sub(startDateTime).Hours()

	rate := int(math.Round((float64(numOfCompleteOrders) / hours) * float64(courierRateFactor)))
	return rate, nil
}
