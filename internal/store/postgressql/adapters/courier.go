package adapters

import (
	"database/sql"
	"github.com/labstack/gommon/log"
	"yandex-lavka-service/internal/domain/entity"
	"yandex-lavka-service/internal/store/postgressql/client"
)

const (
	getCourierTypeId = `SELECT courier_type_id, courier_type FROM courier_types WHERE courier_type = $1`

	insertCourierToCouriers = `INSERT INTO couriers (courier_type_id)
												VALUES ($1)
												RETURNING courier_id`

	insertCourierRegionToRegion = `INSERT INTO couriers_to_regions (courier_id, region)
												VALUES ($1, $2)
												RETURNING region`

	insertCourierWorkingHoursToWorkingHours = `INSERT INTO working_hours (courier_id, working_interval)
												VALUES ($1, $2)
												RETURNING working_interval`

	getCourierFromCouriers = `SELECT courier_id, courier_type_id
								FROM couriers
								WHERE courier_id = $1`

	getCourierTypeFromCourierTypes = `SELECT courier_type
										FROM courier_types
										WHERE courier_type_id = $1`

	getRegionsFromCourierRegion = `SELECT region
									FROM couriers_to_regions
									WHERE courier_id = $1`

	getWorkingHoursFromWorkingHours = `SELECT working_interval
										FROM working_hours
										WHERE courier_id = $1`

	getAllCouriersFromCouriers = `SELECT courier_id
									FROM couriers
									OFFSET $1
									LIMIT $2`

	getCostByCourierIdFromOrders = `SELECT cost
									FROM orders
									WHERE courier_id = $1 AND completed_time BETWEEN $2 AND $3`
)

type CourierRepository struct {
	client *client.PostgresSql
}

func NewCourierRepository(client *client.PostgresSql) *CourierRepository {
	return &CourierRepository{client: client}
}

func (c *CourierRepository) Add(couriers ...entity.Courier) (*[]entity.CourierDto, error) {
	log.Debug("Add couriers to db")
	addedCouriers := make([]entity.CourierDto, 0, len(couriers))
	tx, err := c.client.Db.Beginx()
	if err != nil {
		return nil, err
	}
	for _, courier := range couriers {
		var addedCourier entity.CourierDto
		addedCourier.Regions = make([]int, 0, len(courier.Regions))
		addedCourier.WorkingHours = make([]string, 0, len(courier.WorkingHours))
		var courierTypeId int
		var courierType string
		err = c.client.Db.QueryRowx(getCourierTypeId, courier.CourierType).Scan(&courierTypeId, &courierType)
		if err != nil {
			return nil, err
		}
		var courierId int64
		errQuery := tx.QueryRowx(insertCourierToCouriers, courierTypeId).Scan(&courierId)
		if errQuery != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				return nil, errTx
			}
			return nil, errQuery
		}
		for _, region := range courier.Regions {
			var addedRegion int
			errQuery := tx.QueryRowx(insertCourierRegionToRegion, courierId, region).Scan(&addedRegion)
			if errQuery != nil {
				errTx := tx.Rollback()
				if errTx != nil {
					return nil, errTx
				}
				return nil, errQuery
			}
			addedCourier.Regions = append(addedCourier.Regions, addedRegion)
		}

		for _, workHours := range courier.WorkingHours {
			var addedWorkHours string
			errQuery := tx.QueryRowx(insertCourierWorkingHoursToWorkingHours, courierId, workHours).Scan(&addedWorkHours)
			if errQuery != nil {
				errTx := tx.Rollback()
				if errTx != nil {
					return nil, errTx
				}
				return nil, errQuery
			}
			addedCourier.WorkingHours = append(addedCourier.WorkingHours, addedWorkHours)
		}
		addedCourier.CourierId = courierId
		addedCourier.CourierType = courierType
		addedCouriers = append(addedCouriers, addedCourier)
	}
	errTx := tx.Commit()
	if errTx != nil {
		return nil, errTx
	}
	return &addedCouriers, nil
}

func (c *CourierRepository) Get(id int64) (*entity.CourierDto, error) {
	log.Debug("Get courier from db")
	var courierId int64
	var courierTypeId int
	err := c.client.Db.QueryRowx(getCourierFromCouriers, id).Scan(&courierId, &courierTypeId)
	if err != nil {
		return nil, err
	}
	var courierType string
	err = c.client.Db.QueryRowx(getCourierTypeFromCourierTypes, courierTypeId).Scan(&courierType)
	if err != nil {
		return nil, err
	}
	var courierRegions []int
	err = c.client.Db.Select(&courierRegions, getRegionsFromCourierRegion, courierId)
	if err != nil {
		return nil, err
	}

	var courierWorkingHours []string
	err = c.client.Db.Select(&courierWorkingHours, getWorkingHoursFromWorkingHours, courierId)
	if err != nil {
		return nil, err
	}

	courier := &entity.CourierDto{
		CourierId:    courierId,
		CourierType:  courierType,
		Regions:      courierRegions,
		WorkingHours: courierWorkingHours,
	}
	return courier, nil

}

func (c *CourierRepository) GetAll(limit int, offset int) (*[]entity.CourierDto, error) {
	log.Debug("Get couriers from db")
	var courierIds []int64
	err := c.client.Db.Select(&courierIds, getAllCouriersFromCouriers, offset, limit)
	if err != nil {
		return nil, err
	}

	if len(courierIds) == 0 {
		return nil, sql.ErrNoRows
	}

	couriers := make([]entity.CourierDto, 0, len(courierIds))
	for _, courierId := range courierIds {
		courier, err := c.Get(courierId)
		if err != nil {
			return nil, err
		}
		couriers = append(couriers, *courier)
	}

	return &couriers, nil
}

func (c *CourierRepository) GetCosts(id int64, startDate string, endDate string) (costs []int, err error) {
	log.Debug("Get costs from db")
	err = c.client.Db.Select(&costs, getCostByCourierIdFromOrders, id, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return costs, nil
}
