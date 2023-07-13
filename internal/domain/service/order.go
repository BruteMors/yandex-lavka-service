package service

import (
	"github.com/labstack/gommon/log"
	"yandex-lavka-service/internal/domain/entity"
)

// OrderStore - interface type for database
type OrderStore interface {
	Add(orders ...entity.Order) (*[]entity.OrderDto, error)
	Get(id int64) (*entity.OrderDto, error)
	GetAll(limit int, offset int) (*[]entity.OrderDto, error)
	SetCompleted(completedOrders ...entity.CompleteOrder) (*[]entity.OrderDto, error)
}

type Order struct {
	store OrderStore
}

// NewOrder - create Order service instance
func NewOrder(store OrderStore) *Order {
	return &Order{store: store}
}

// AddOrders - add orders to database
func (c *Order) AddOrders(orders ...entity.Order) (*[]entity.OrderDto, error) {
	log.Debug("Add orders to db")
	addedOrders, err := c.store.Add(orders...)
	if err != nil {
		return nil, err
	}
	return addedOrders, nil
}

// GetOrder - get order from database
func (c *Order) GetOrder(id int64) (*entity.OrderDto, error) {
	log.Debug("get order from database")
	order, err := c.store.Get(id)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrders - get all orders from database
func (c *Order) GetOrders(limit int, offset int) (*[]entity.OrderDto, error) {
	log.Debug("get orders from database")
	orders, err := c.store.GetAll(limit, offset)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// SetCompleteOrders - mark orders in database as complete
func (c *Order) SetCompleteOrders(completedOrders ...entity.CompleteOrder) (*[]entity.OrderDto, error) {
	log.Debug("Add orders to db")
	addedOrders, err := c.store.SetCompleted(completedOrders...)
	if err != nil {
		return nil, err
	}
	return addedOrders, nil
}
