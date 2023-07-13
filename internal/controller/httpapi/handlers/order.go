package handlers

import (
	"database/sql"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"strconv"
	"yandex-lavka-service/internal/controller/httpapi/responses"
	"yandex-lavka-service/internal/domain/entity"
	"yandex-lavka-service/internal/domain/service"
)

type Order struct {
	service  *service.Order
	validate *validator.Validate
}

// NewOrder - create new handler order instance
func NewOrder(service *service.Order, validate *validator.Validate) *Order {
	return &Order{service: service, validate: validate}
}

// AddOrders - add Orders to database
func (c *Order) AddOrders(e echo.Context) error {
	log.Debug("add orders by POST /orders")
	ordersDTO := new(entity.CreateOrderRequest)
	badRequestResponse := responses.BadRequestResponse{}
	if err := e.Bind(ordersDTO); err != nil {
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	err := c.validate.Struct(ordersDTO)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	orders := make([]entity.Order, len(ordersDTO.Orders))
	for i, dto := range ordersDTO.Orders {
		orders[i] = c.transferOrderDTOToOrder(dto)
	}

	addedOrders, err := c.service.AddOrders(orders...)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	createOrdersResponse := responses.CreateOrdersResponse{
		Orders: addedOrders,
	}

	return e.JSON(http.StatusOK, createOrdersResponse)
}

// GetOrder - get info about order by order id
func (c *Order) GetOrder(e echo.Context) error {
	log.Debug("get order by GET /orders/:id")
	badRequestResponse := responses.BadRequestResponse{}
	orderId := e.Param("id")

	orderIdInt, err := strconv.ParseInt(orderId, 10, 64)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	err = c.validate.Var(orderIdInt, "numeric,gt=0")
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	order, err := c.service.GetOrder(orderIdInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info(err)
			notFoundResponse := responses.NotFoundResponse{}
			return e.JSON(http.StatusNotFound, notFoundResponse)
		}
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}
	createOrderResponse := responses.GetOrderResponse{OrderDto: *order}
	return e.JSON(http.StatusOK, createOrderResponse)
}

// GetOrders - get info about all orders
func (c *Order) GetOrders(e echo.Context) error {
	log.Debug("get orders by GET /orders")
	badRequestResponse := responses.BadRequestResponse{}
	limit := e.QueryParam("limit")
	offset := e.QueryParam("offset")

	var limitInt int
	if limit != "" {
		var err error
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			log.Info(err)
			return e.JSON(http.StatusBadRequest, badRequestResponse)
		}
		err = c.validate.Var(limitInt, "numeric,gt=0")
		if err != nil {
			log.Info(err)
			return e.JSON(http.StatusBadRequest, badRequestResponse)
		}

	} else {
		limitInt = 1
	}

	var offsetInt int
	if offset != "" {
		var err error
		offsetInt, err = strconv.Atoi(offset)
		if err != nil {
			log.Info(err)
			return e.JSON(http.StatusBadRequest, badRequestResponse)
		}

		err = c.validate.Var(offsetInt, "numeric,gte=0")
		if err != nil {
			log.Info(err)
			return e.JSON(http.StatusBadRequest, badRequestResponse)
		}

	} else {
		offsetInt = 0
	}

	orders, err := c.service.GetOrders(limitInt, offsetInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			getOrdersResponse := responses.GetOrdersResponse{
				Orders: &[]entity.OrderDto{},
				Limit:  limitInt,
				Offset: offsetInt,
			}

			return e.JSON(http.StatusOK, getOrdersResponse)
		}
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}
	getOrdersResponse := responses.GetOrdersResponse{
		Orders: orders,
		Limit:  limitInt,
		Offset: offsetInt,
	}

	return e.JSON(http.StatusOK, getOrdersResponse)
}

// SetCompleteOrders - mark orders as completed
func (c *Order) SetCompleteOrders(e echo.Context) error {
	log.Debug("set complete orders by POST /orders/complete")
	ordersDTO := new(entity.CompleteOrderRequestDTO)
	badRequestResponse := responses.BadRequestResponse{}
	if err := e.Bind(ordersDTO); err != nil {
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	err := c.validate.Struct(ordersDTO)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	setOrders, err := c.service.SetCompleteOrders(ordersDTO.CompleteInfo...)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	setCompleteOrdersResponse := responses.SetCompleteOrdersResponse{
		Orders: setOrders,
	}

	return e.JSON(http.StatusOK, setCompleteOrdersResponse.Orders)
}

// transferOrderDTOToOrder - func for transfer CreateOrderDTO struct to Order struct
func (c *Order) transferOrderDTOToOrder(dto entity.CreateOrderDTO) entity.Order {
	return entity.Order{
		Weight:        dto.Weight,
		Regions:       dto.Regions,
		DeliveryHours: dto.DeliveryHours,
		Cost:          dto.Cost,
		CompletedTime: "",
	}
}
