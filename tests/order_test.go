package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"yandex-lavka-service/internal/controller/httpapi/handlers"
	"yandex-lavka-service/internal/controller/httpapi/responses"
	"yandex-lavka-service/internal/domain/entity"
	"yandex-lavka-service/internal/domain/service"
	"yandex-lavka-service/internal/store/postgressql/adapters"
	"yandex-lavka-service/internal/store/postgressql/adapters/mocks"
)

func TestOrder_AddOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderStore := mock_service.NewMockOrderStore(ctrl)

	orderService := service.NewOrder(mockOrderStore)

	var validate = validator.New()
	err := validate.RegisterValidation("time_interval_format", handlers.ValidateTimeInterval)
	require.NoError(t, err)

	h := handlers.NewOrder(orderService, validate)

	e := echo.New()

	// Test the success scenario
	{
		orders := []entity.CreateOrderDTO{
			{Weight: 10, Regions: 1, DeliveryHours: []string{"10:00-12:00", "14:00-15:00"}, Cost: 20},
			{Weight: 20, Regions: 2, DeliveryHours: []string{"11:00-12:30", "14:00-16:00"}, Cost: 10}}

		createOrderRequest := entity.CreateOrderRequest{Orders: orders}

		ordersDTO := []entity.OrderDto{
			{OrderId: 1, Weight: orders[0].Weight, Regions: orders[0].Regions, DeliveryHours: orders[0].DeliveryHours, Cost: orders[0].Cost, CompletedTime: ""},
			{OrderId: 2, Weight: orders[1].Weight, Regions: orders[1].Regions, DeliveryHours: orders[1].DeliveryHours, Cost: orders[1].Cost, CompletedTime: ""}}

		reqBody, _ := json.Marshal(createOrderRequest)
		mockOrderStore.EXPECT().Add([]entity.Order{
			{Weight: orders[0].Weight, Regions: orders[0].Regions, DeliveryHours: orders[0].DeliveryHours, Cost: orders[0].Cost, CompletedTime: ""},
			{Weight: orders[1].Weight, Regions: orders[1].Regions, DeliveryHours: orders[1].DeliveryHours, Cost: orders[1].Cost, CompletedTime: ""}}).Return(&ordersDTO, nil)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err = h.AddOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.CreateOrdersResponse{
			Orders: &ordersDTO,
		}
		var actualResponse responses.CreateOrdersResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}

	// Test the bad request scenario (invalid delivery hours)
	{
		orders := []entity.CreateOrderDTO{
			{Weight: 10, Regions: 1, DeliveryHours: []string{"10:00-12:90", "14:00-15:00"}, Cost: 20},
			{Weight: 20, Regions: 2, DeliveryHours: []string{"11:00-12:30", "14:00-16:00"}, Cost: 10}}

		createOrderRequest := entity.CreateOrderRequest{Orders: orders}

		reqBody, _ := json.Marshal(createOrderRequest)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err = h.AddOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// Test the bad request scenario (invalid weight)
	{
		orders := []entity.CreateOrderDTO{
			{Weight: 0, Regions: 1, DeliveryHours: []string{"10:00-12:00", "14:00-15:00"}, Cost: 20},
			{Weight: 20, Regions: 2, DeliveryHours: []string{"11:00-12:30", "14:00-16:00"}, Cost: 10}}

		createOrderRequest := entity.CreateOrderRequest{Orders: orders}

		reqBody, _ := json.Marshal(createOrderRequest)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err = h.AddOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// Test the bad request scenario (invalid region)
	{
		orders := []entity.CreateOrderDTO{
			{Weight: 10, Regions: 0, DeliveryHours: []string{"10:00-12:00", "14:00-15:00"}, Cost: 20},
			{Weight: 20, Regions: 2, DeliveryHours: []string{"11:00-12:30", "14:00-16:00"}, Cost: 10}}

		createOrderRequest := entity.CreateOrderRequest{Orders: orders}

		reqBody, _ := json.Marshal(createOrderRequest)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err = h.AddOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestOrder_GetOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderStore := mock_service.NewMockOrderStore(ctrl)

	orderService := service.NewOrder(mockOrderStore)

	var validate = validator.New()
	err := validate.RegisterValidation("time_interval_format", handlers.ValidateTimeInterval)
	require.NoError(t, err)

	h := handlers.NewOrder(orderService, validate)

	e := echo.New()

	// Test the success scenario
	{
		ordersDTO := entity.OrderDto{OrderId: 1, Weight: 10, Regions: 1, DeliveryHours: []string{"10:00-12:00", "14:00-15:00"}, Cost: 10, CompletedTime: "2023-05-02 11:25:43.022000 +00:00"}

		mockOrderStore.EXPECT().Get(ordersDTO.OrderId).Return(&ordersDTO, nil)

		idString := strconv.FormatInt(ordersDTO.OrderId, 10)
		req := httptest.NewRequest(http.MethodGet, "/orders/"+idString, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/orders/:id")
		c.SetParamNames("id")
		c.SetParamValues(idString)

		err = h.GetOrder(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetOrderResponse{OrderDto: ordersDTO}
		var actualResponse responses.GetOrderResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)

	}

	// 	// Test the validation error scenario
	{
		idString := strconv.FormatInt(0, 10)
		req := httptest.NewRequest(http.MethodGet, "/orders/"+idString, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/orders/:id")
		c.SetParamNames("id")
		c.SetParamValues(idString)

		err = h.GetOrder(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// Test the error not found scenario
	{
		orderId := int64(1000)
		idString := strconv.FormatInt(orderId, 10)
		mockOrderStore.EXPECT().Get(orderId).Return(nil, sql.ErrNoRows)
		req := httptest.NewRequest(http.MethodGet, "/orders/"+idString, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/orders/:id")
		c.SetParamNames("id")
		c.SetParamValues(idString)

		err = h.GetOrder(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestOrder_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderStore := mock_service.NewMockOrderStore(ctrl)

	orderService := service.NewOrder(mockOrderStore)

	var validate = validator.New()
	err := validate.RegisterValidation("time_interval_format", handlers.ValidateTimeInterval)
	require.NoError(t, err)

	h := handlers.NewOrder(orderService, validate)

	e := echo.New()

	// Test the success scenario
	{
		orders := []entity.CreateOrderDTO{
			{Weight: 10, Regions: 1, DeliveryHours: []string{"10:00-12:00", "14:00-15:00"}, Cost: 20},
			{Weight: 20, Regions: 2, DeliveryHours: []string{"11:00-12:30", "14:00-16:00"}, Cost: 10}}

		ordersDTO := []entity.OrderDto{
			{OrderId: 1, Weight: orders[0].Weight, Regions: orders[0].Regions, DeliveryHours: orders[0].DeliveryHours, Cost: orders[0].Cost, CompletedTime: ""},
			{OrderId: 2, Weight: orders[1].Weight, Regions: orders[1].Regions, DeliveryHours: orders[1].DeliveryHours, Cost: orders[1].Cost, CompletedTime: ""}}

		mockOrderStore.EXPECT().GetAll(2, 1).Return(&ordersDTO, nil)
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Request().URL.RawQuery = "limit=2&offset=1"
		err = h.GetOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetOrdersResponse{
			Orders: &ordersDTO,
			Limit:  2,
			Offset: 1,
		}
		var actualResponse responses.GetOrdersResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}
	// Test the success scenario without limit and offset query params
	{
		orders := []entity.CreateOrderDTO{
			{Weight: 10, Regions: 1, DeliveryHours: []string{"10:00-12:00", "14:00-15:00"}, Cost: 20},
		}

		ordersDTO := []entity.OrderDto{
			{OrderId: 1, Weight: orders[0].Weight, Regions: orders[0].Regions, DeliveryHours: orders[0].DeliveryHours, Cost: orders[0].Cost, CompletedTime: ""},
		}

		mockOrderStore.EXPECT().GetAll(1, 0).Return(&ordersDTO, nil)
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err = h.GetOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetOrdersResponse{
			Orders: &ordersDTO,
			Limit:  1,
			Offset: 0,
		}
		var actualResponse responses.GetOrdersResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}

	// Test the not found scenario
	{
		orders := &[]entity.OrderDto{}

		mockOrderStore.EXPECT().GetAll(1, 1000).Return(nil, sql.ErrNoRows)
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Request().URL.RawQuery = "limit=1&offset=1000"
		err = h.GetOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetOrdersResponse{
			Orders: orders,
			Limit:  1,
			Offset: 1000,
		}
		var actualResponse responses.GetOrdersResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}

	// Test invalid query param scenario
	{
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Request().URL.RawQuery = "limit=-1&offset=1000"
		err = h.GetOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

}

func TestOrder_SetCompleteOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderStore := mock_service.NewMockOrderStore(ctrl)

	orderService := service.NewOrder(mockOrderStore)

	var validate = validator.New()
	err := validate.RegisterValidation("time_interval_format", handlers.ValidateTimeInterval)
	require.NoError(t, err)

	h := handlers.NewOrder(orderService, validate)

	e := echo.New()

	// Test the success scenario
	{
		orders := []entity.CreateOrderDTO{
			{Weight: 10, Regions: 1, DeliveryHours: []string{"10:00-12:00", "14:00-15:00"}, Cost: 20}}

		ordersDTO := []entity.OrderDto{
			{OrderId: 1, Weight: orders[0].Weight, Regions: orders[0].Regions, DeliveryHours: orders[0].DeliveryHours, Cost: orders[0].Cost, CompletedTime: "2023-05-05T12:02:19.959Z"}}

		completeOrder := []entity.CompleteOrder{
			{CourierId: 1, OrderId: 1, CompleteTime: "2023-05-05T12:02:19.959Z"}}

		completeOrderRequest := entity.CompleteOrderRequestDTO{CompleteInfo: completeOrder}

		reqBody, _ := json.Marshal(completeOrderRequest)
		mockOrderStore.EXPECT().SetCompleted(completeOrder).Return(&ordersDTO, nil)
		req := httptest.NewRequest(http.MethodPost, "/orders/complete", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err = h.SetCompleteOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.SetCompleteOrdersResponse{Orders: &ordersDTO}
		var actualResponse responses.SetCompleteOrdersResponse
		err := json.Unmarshal(rec.Body.Bytes(), &actualResponse.Orders)
		if err != nil {
			return
		}
		assert.Equal(t, expectedResponse.Orders, actualResponse.Orders)
	}

	// Test the bad request scenario (order not found)
	{
		completeOrder := []entity.CompleteOrder{
			{CourierId: 1, OrderId: 1, CompleteTime: "2023-05-05T12:02:19.959Z"}}

		completeOrderRequest := entity.CompleteOrderRequestDTO{CompleteInfo: completeOrder}

		reqBody, _ := json.Marshal(completeOrderRequest)
		mockOrderStore.EXPECT().SetCompleted(completeOrder).Return(nil, sql.ErrNoRows)
		req := httptest.NewRequest(http.MethodPost, "/orders/complete", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err = h.SetCompleteOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// Test the bad request scenario (order set to another courier or courier id not set)
	{
		completeOrder := []entity.CompleteOrder{
			{CourierId: 1, OrderId: 1, CompleteTime: "2023-05-05T12:02:19.959Z"}}

		completeOrderRequest := entity.CompleteOrderRequestDTO{CompleteInfo: completeOrder}

		reqBody, _ := json.Marshal(completeOrderRequest)
		mockOrderStore.EXPECT().SetCompleted(completeOrder).Return(nil, adapters.ErrorCourierIdNotEq)
		req := httptest.NewRequest(http.MethodPost, "/orders/complete", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err = h.SetCompleteOrders(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}
