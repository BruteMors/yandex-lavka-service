package adapters

import (
	"database/sql"
	"errors"
	"github.com/labstack/gommon/log"
	"yandex-lavka-service/internal/domain/entity"
	"yandex-lavka-service/internal/store/postgressql/client"
)

const (
	insertOrderToOrders = `INSERT INTO orders (region, weight, cost)
						VALUES ($1,$2,$3)
						RETURNING order_id, region, weight, cost`

	insertDeliveryHoursToDeliveryHours = `INSERT INTO delivery_hours (order_id, delivery_interval)
										VALUES ($1, $2)
										RETURNING delivery_interval`

	getCompletedTimeFromOrders = `SELECT completed_time FROM orders
									WHERE order_id = $1`

	getOrderFromOrders = `	SELECT order_id, weight, region, cost, completed_time
							FROM orders
							WHERE order_id = $1`

	getDeliveryHoursFromDeliveryHours = `SELECT delivery_interval
											FROM delivery_hours
											WHERE order_id = $1`

	getOrdersIdFromOrders = `	SELECT order_id
											FROM orders
											OFFSET $1
											LIMIT $2`

	setCompletedTimeOrderToOrders = `	UPDATE orders
										SET courier_id = $1, completed_time = $2
										WHERE order_id = $3
										RETURNING order_id, courier_id,  region, weight, cost, completed_time`

	getCourierIdFromOrders = `	SELECT courier_id
								FROM orders
								WHERE order_id = $1`
)

var ErrorCourierIdNotEq = errors.New("courier id in orders table not eq courier id in request")

type OrderRepository struct {
	client *client.PostgresSql
}

func NewOrderRepository(client *client.PostgresSql) *OrderRepository {
	return &OrderRepository{client: client}
}

func (c *OrderRepository) Add(orders ...entity.Order) (*[]entity.OrderDto, error) {
	log.Debug("Add orders to db")
	addedOrders := make([]entity.OrderDto, 0, len(orders))
	tx, err := c.client.Db.Beginx()
	if err != nil {
		return nil, err
	}
	for _, order := range orders {
		var addedOrder entity.OrderDto
		var orderId int64
		var weight float32
		var regions int
		var cost int
		var completedTime sql.NullString
		addedOrder.DeliveryHours = make([]string, 0, len(order.DeliveryHours))

		errQuery := tx.QueryRowx(insertOrderToOrders, order.Regions, order.Weight, order.Cost).Scan(&orderId, &regions, &weight, &cost)
		if errQuery != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				return nil, errTx
			}
			return nil, errQuery
		}

		for _, deliveryInterval := range order.DeliveryHours {
			var addedDeliveryInterval string
			errQuery := tx.QueryRowx(insertDeliveryHoursToDeliveryHours, orderId, deliveryInterval).Scan(&addedDeliveryInterval)
			if errQuery != nil {
				errTx := tx.Rollback()
				if errTx != nil {
					return nil, errTx
				}
				return nil, errQuery
			}
			addedOrder.DeliveryHours = append(addedOrder.DeliveryHours, addedDeliveryInterval)
		}

		errQuery = tx.QueryRowx(getCompletedTimeFromOrders, orderId).Scan(&completedTime)
		if errQuery != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				return nil, errTx
			}
			return nil, errQuery
		}

		addedOrder.OrderId = orderId
		addedOrder.Regions = regions
		addedOrder.Cost = cost
		addedOrder.Weight = weight
		addedOrder.CompletedTime = completedTime.String
		addedOrders = append(addedOrders, addedOrder)
	}
	errTx := tx.Commit()
	if errTx != nil {
		return nil, errTx
	}
	return &addedOrders, nil
}

func (c *OrderRepository) Get(id int64) (*entity.OrderDto, error) {
	log.Debug("Get order from db")
	var orderId int64
	var weight float32
	var regions int
	var cost int
	var completedTime sql.NullString

	err := c.client.Db.QueryRowx(getOrderFromOrders, id).Scan(&orderId, &weight, &regions, &cost, &completedTime)
	if err != nil {
		return nil, err
	}

	var deliveryHours []string
	err = c.client.Db.Select(&deliveryHours, getDeliveryHoursFromDeliveryHours, orderId)
	if err != nil {
		return nil, err
	}

	order := &entity.OrderDto{
		OrderId:       orderId,
		Weight:        weight,
		Regions:       regions,
		DeliveryHours: deliveryHours,
		Cost:          cost,
		CompletedTime: completedTime.String,
	}
	return order, nil

}

func (c *OrderRepository) GetAll(limit int, offset int) (*[]entity.OrderDto, error) {
	log.Debug("Get orders from db")
	var orderIds []int64
	err := c.client.Db.Select(&orderIds, getOrdersIdFromOrders, offset, limit)
	if err != nil {
		return nil, err
	}

	if len(orderIds) == 0 {
		return nil, sql.ErrNoRows
	}

	orders := make([]entity.OrderDto, 0, len(orderIds))
	for _, orderId := range orderIds {
		order, err := c.Get(orderId)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *order)
	}

	return &orders, nil
}

func (c *OrderRepository) SetCompleted(completedOrders ...entity.CompleteOrder) (*[]entity.OrderDto, error) {
	log.Debug("Mark orders as completed")

	updatedOrders := make([]entity.OrderDto, 0, len(completedOrders))
	tx, err := c.client.Db.Beginx()
	if err != nil {
		return nil, err
	}
	for _, orderCompleteInfo := range completedOrders {
		var courierId sql.NullInt64
		errQuery := c.client.Db.QueryRowx(getCourierIdFromOrders, orderCompleteInfo.OrderId).Scan(&courierId)
		if errQuery != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				return nil, errTx
			}
			return nil, errQuery
		}

		if !courierId.Valid {
			errTx := tx.Rollback()
			if errTx != nil {
				return nil, errTx
			}
			return nil, ErrorCourierIdNotEq
		}

		value, err := courierId.Value()
		if err != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				return nil, errTx
			}
			return nil, ErrorCourierIdNotEq
		}

		if value != orderCompleteInfo.CourierId {
			errTx := tx.Rollback()
			if errTx != nil {
				return nil, errTx
			}
			return nil, ErrorCourierIdNotEq
		}

		var updatedOrder entity.OrderDto
		var orderId int64
		var weight float32
		var regions int
		var cost int
		var completedTime string

		errQuery = tx.QueryRowx(setCompletedTimeOrderToOrders, orderCompleteInfo.CourierId, orderCompleteInfo.CompleteTime, orderCompleteInfo.OrderId).Scan(&orderId, &courierId, &regions, &weight, &cost, &completedTime)
		if errQuery != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				return nil, errTx
			}
			return nil, errQuery
		}

		var deliveryHours []string
		err = c.client.Db.Select(&deliveryHours, getDeliveryHoursFromDeliveryHours, orderId)
		if err != nil {
			return nil, err
		}

		updatedOrder.OrderId = orderId
		updatedOrder.Weight = weight
		updatedOrder.Regions = regions
		updatedOrder.DeliveryHours = deliveryHours
		updatedOrder.Cost = cost
		updatedOrder.CompletedTime = completedTime

		updatedOrders = append(updatedOrders, updatedOrder)
	}
	errTx := tx.Commit()
	if errTx != nil {
		return nil, errTx
	}
	return &updatedOrders, nil
}
