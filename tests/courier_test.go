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
	"yandex-lavka-service/internal/config"
	"yandex-lavka-service/internal/controller/httpapi/handlers"
	"yandex-lavka-service/internal/controller/httpapi/responses"
	"yandex-lavka-service/internal/domain/entity"
	"yandex-lavka-service/internal/domain/service"
	mock_service "yandex-lavka-service/internal/store/postgressql/adapters/mocks"
)

func TestCourier_AddCouriers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCourierStore := mock_service.NewMockCourierStore(ctrl)
	cfg, err := config.New()
	require.NoError(t, err)

	cases := []struct {
		name string
		entity.CreateCourierRequest
	}{
		{name: "valid request", CreateCourierRequest: entity.CreateCourierRequest{
			Couriers: []entity.CreateCourierDTO{
				entity.CreateCourierDTO{
					CourierType:  "FOOT",
					Regions:      []int{1, 2, 3},
					WorkingHours: []string{"10:00-12:00"},
				},
			},
		}},
		{name: "validation error courier type", CreateCourierRequest: entity.CreateCourierRequest{
			Couriers: []entity.CreateCourierDTO{
				entity.CreateCourierDTO{
					CourierType:  "FOT",
					Regions:      []int{1, 2, 3},
					WorkingHours: []string{"10:00-12:00"},
				},
			},
		}},
		{name: "validation error work hours", CreateCourierRequest: entity.CreateCourierRequest{
			Couriers: []entity.CreateCourierDTO{
				entity.CreateCourierDTO{
					CourierType:  "FOOT",
					Regions:      []int{1, 2, 3},
					WorkingHours: []string{"11:70-12:00"},
				},
			},
		}},
	}

	expectedAnswers := []struct {
		name string
		responses.CreateCouriersResponse
	}{
		{
			name:                   "valid request",
			CreateCouriersResponse: responses.CreateCouriersResponse{Couriers: &[]entity.CourierDto{{CourierId: 1, CourierType: "FOOT", Regions: []int{1, 2, 3}, WorkingHours: []string{"10:00-12:00"}}}},
		},
	}

	mockCourierStore.EXPECT().Add(entity.Courier{
		CourierType:  cases[0].CreateCourierRequest.Couriers[0].CourierType,
		Regions:      cases[0].CreateCourierRequest.Couriers[0].Regions,
		WorkingHours: cases[0].CreateCourierRequest.Couriers[0].WorkingHours,
	}).Return(&[]entity.CourierDto{{CourierId: 1, CourierType: "FOOT", Regions: []int{1, 2, 3}, WorkingHours: []string{"10:00-12:00"}}}, nil)

	courierService := service.NewCourier(mockCourierStore, cfg)

	var validate = validator.New()
	err = validate.RegisterValidation("time_interval_format", handlers.ValidateTimeInterval)
	require.NoError(t, err)

	h := handlers.NewCourier(courierService, validate)

	e := echo.New()

	// Test the success scenario

	reqBody, _ := json.Marshal(cases[0])

	req := httptest.NewRequest(http.MethodPost, "/couriers", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.AddCouriers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	expectedResponse := expectedAnswers[0].CreateCouriersResponse
	var actualResponse responses.CreateCouriersResponse
	json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	assert.Equal(t, expectedResponse, actualResponse)

	// Test the bad request scenarios

	for i := 1; i < len(cases); i++ {

		reqBody, _ = json.Marshal(cases[i])

		req = httptest.NewRequest(http.MethodPost, "/couriers", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		err = h.AddCouriers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		expectedBadRequestResponse := responses.BadRequestResponse{}
		var actualBadRequestResponse responses.BadRequestResponse
		json.Unmarshal(rec.Body.Bytes(), &actualBadRequestResponse)
		assert.Equal(t, expectedBadRequestResponse, actualBadRequestResponse)
	}

}

func TestCourier_GetCourier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCourierStore := mock_service.NewMockCourierStore(ctrl)
	cfg, err := config.New()
	require.NoError(t, err)

	cases := []struct {
		name string
		id   int64
	}{
		{name: "valid request", id: 1},
		{name: "validation error id", id: -1},
		{name: "error not found", id: 1000},
	}

	expectedAnswers := []struct {
		name string
		resp responses.GetCourierResponse
	}{
		{
			name: "valid request",
			resp: struct{ entity.CourierDto }{entity.CourierDto{
				CourierId:    1,
				CourierType:  "FOOT",
				Regions:      []int{1, 2, 3},
				WorkingHours: []string{"10:00-12:00"},
			}},
		},
	}

	mockCourierStore.EXPECT().Get(cases[0].id).Return(&entity.CourierDto{CourierId: cases[0].id, CourierType: "FOOT", Regions: []int{1, 2, 3}, WorkingHours: []string{"10:00-12:00"}}, nil)
	mockCourierStore.EXPECT().Get(cases[2].id).Return(nil, sql.ErrNoRows)

	courierService := service.NewCourier(mockCourierStore, cfg)

	var validate = validator.New()
	err = validate.RegisterValidation("time_interval_format", handlers.ValidateTimeInterval)
	require.NoError(t, err)

	h := handlers.NewCourier(courierService, validate)

	e := echo.New()

	// Test the success scenario

	idString := strconv.FormatInt(cases[0].id, 10)
	req := httptest.NewRequest(http.MethodGet, "/couriers/"+idString, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/couriers/:id")
	c.SetParamNames("id")
	c.SetParamValues(idString)

	err = h.GetCourier(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	expectedResponse := expectedAnswers[0].resp
	var actualResponse responses.GetCourierResponse
	json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	assert.Equal(t, expectedResponse, actualResponse)

	// Test the validation error scenario

	idString = strconv.FormatInt(cases[1].id, 10)
	req = httptest.NewRequest(http.MethodGet, "/couriers/"+idString, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/couriers/:id")
	c.SetParamNames("id")
	c.SetParamValues(idString)

	err = h.GetCourier(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Test the error not found scenario

	idString = strconv.FormatInt(cases[2].id, 10)
	req = httptest.NewRequest(http.MethodGet, "/couriers/"+idString, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/couriers/:id")
	c.SetParamNames("id")
	c.SetParamValues(idString)

	err = h.GetCourier(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

}

func TestCourier_GetCouriers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCourierStore := mock_service.NewMockCourierStore(ctrl)
	cfg, err := config.New()
	require.NoError(t, err)

	courierService := service.NewCourier(mockCourierStore, cfg)

	var validate = validator.New()
	err = validate.RegisterValidation("time_interval_format", handlers.ValidateTimeInterval)
	require.NoError(t, err)

	h := handlers.NewCourier(courierService, validate)

	e := echo.New()

	// Test the success scenario
	{
		couriers := &[]entity.CourierDto{
			{CourierId: 1, CourierType: "FOOT", Regions: []int{1, 2, 3}, WorkingHours: []string{"10:00-12:00"}},
			{CourierId: 2, CourierType: "BIKE", Regions: []int{1, 2}, WorkingHours: []string{"10:00-12:00", "14:00-16:00"}}}

		mockCourierStore.EXPECT().GetAll(2, 1).Return(couriers, nil)
		req := httptest.NewRequest(http.MethodGet, "/couriers", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Request().URL.RawQuery = "limit=2&offset=1"
		err = h.GetCouriers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetCouriersResponse{
			Couriers: couriers,
			Limit:    2,
			Offset:   1,
		}
		var actualResponse responses.GetCouriersResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}
	// Test the success scenario without limit and offset query params
	{
		couriers := &[]entity.CourierDto{
			{CourierId: 1, CourierType: "FOOT", Regions: []int{1, 2, 3}, WorkingHours: []string{"10:00-12:00"}}}

		mockCourierStore.EXPECT().GetAll(1, 0).Return(couriers, nil)
		req := httptest.NewRequest(http.MethodGet, "/couriers", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err = h.GetCouriers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetCouriersResponse{
			Couriers: couriers,
			Limit:    1,
			Offset:   0,
		}
		var actualResponse responses.GetCouriersResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}

	// Test the not found scenario
	{
		couriers := &[]entity.CourierDto{}

		mockCourierStore.EXPECT().GetAll(1, 1000).Return(nil, sql.ErrNoRows)
		req := httptest.NewRequest(http.MethodGet, "/couriers", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Request().URL.RawQuery = "limit=1&offset=1000"
		err = h.GetCouriers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetCouriersResponse{
			Couriers: couriers,
			Limit:    1,
			Offset:   1000,
		}
		var actualResponse responses.GetCouriersResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}

	// Test invalid query param scenario
	{
		req := httptest.NewRequest(http.MethodGet, "/couriers", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Request().URL.RawQuery = "limit=-1&offset=1000"
		err = h.GetCouriers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

}

func TestCourier_GetCourierMetaInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCourierStore := mock_service.NewMockCourierStore(ctrl)
	cfg, err := config.New()
	require.NoError(t, err)

	courierService := service.NewCourier(mockCourierStore, cfg)

	var validate = validator.New()
	err = validate.RegisterValidation("time_interval_format", handlers.ValidateTimeInterval)
	require.NoError(t, err)

	h := handlers.NewCourier(courierService, validate)

	e := echo.New()

	// Test the success scenario for foot
	{
		courier := &entity.CourierMetaInfo{
			CourierId:    1,
			CourierType:  "FOOT",
			Regions:      []int{1, 2, 3},
			WorkingHours: []string{"12:00-14:00"},
			Rating:       0,
			Earnings:     60,
		}

		mockCourierStore.EXPECT().Get(courier.CourierId).Return(&entity.CourierDto{
			CourierId:    courier.CourierId,
			CourierType:  courier.CourierType,
			Regions:      courier.Regions,
			WorkingHours: courier.WorkingHours,
		}, nil)

		mockCourierStore.EXPECT().GetCosts(courier.CourierId, "2023-05-01", "2023-05-02").Return([]int{5, 10, 15}, nil)

		idString := strconv.FormatInt(courier.CourierId, 10)
		req := httptest.NewRequest(http.MethodGet, "/couriers/meta-info/"+idString, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/couriers/:id")
		c.SetParamNames("id")
		c.SetParamValues(idString)
		c.Request().URL.RawQuery = "start_date=2023-05-01&end_date=2023-05-02"
		err = h.GetCourierMetaInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetCourierMetaInfoResponse{*courier}
		var actualResponse responses.GetCourierMetaInfoResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}

	// Test the success scenario for bike
	{
		courier := &entity.CourierMetaInfo{
			CourierId:    1,
			CourierType:  "BIKE",
			Regions:      []int{1, 2, 3},
			WorkingHours: []string{"12:00-14:00"},
			Rating:       1,
			Earnings:     102,
		}

		mockCourierStore.EXPECT().Get(courier.CourierId).Return(&entity.CourierDto{
			CourierId:    courier.CourierId,
			CourierType:  courier.CourierType,
			Regions:      courier.Regions,
			WorkingHours: courier.WorkingHours,
		}, nil)

		mockCourierStore.EXPECT().GetCosts(courier.CourierId, "2023-05-01", "2023-05-02").Return([]int{5, 10, 15, 1, 1, 1, 1}, nil)

		idString := strconv.FormatInt(courier.CourierId, 10)
		req := httptest.NewRequest(http.MethodGet, "/couriers/meta-info/"+idString, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/couriers/:id")
		c.SetParamNames("id")
		c.SetParamValues(idString)
		c.Request().URL.RawQuery = "start_date=2023-05-01&end_date=2023-05-02"
		err = h.GetCourierMetaInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetCourierMetaInfoResponse{*courier}
		var actualResponse responses.GetCourierMetaInfoResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}

	// Test the success scenario for auto
	{
		courier := &entity.CourierMetaInfo{
			CourierId:    1,
			CourierType:  "AUTO",
			Regions:      []int{1, 2, 3},
			WorkingHours: []string{"12:00-14:00"},
			Rating:       0,
			Earnings:     120,
		}

		mockCourierStore.EXPECT().Get(courier.CourierId).Return(&entity.CourierDto{
			CourierId:    courier.CourierId,
			CourierType:  courier.CourierType,
			Regions:      courier.Regions,
			WorkingHours: courier.WorkingHours,
		}, nil)

		mockCourierStore.EXPECT().GetCosts(courier.CourierId, "2023-05-01", "2023-05-02").Return([]int{5, 10, 15}, nil)

		idString := strconv.FormatInt(courier.CourierId, 10)
		req := httptest.NewRequest(http.MethodGet, "/couriers/meta-info/"+idString, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/couriers/:id")
		c.SetParamNames("id")
		c.SetParamValues(idString)
		c.Request().URL.RawQuery = "start_date=2023-05-01&end_date=2023-05-02"
		err = h.GetCourierMetaInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetCourierMetaInfoResponse{*courier}
		var actualResponse responses.GetCourierMetaInfoResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}

	// Test the scenario without orders
	{
		courier := &entity.CourierMetaInfo{
			CourierId:    1,
			CourierType:  "AUTO",
			Regions:      []int{1, 2, 3},
			WorkingHours: []string{"12:00-14:00"},
			Rating:       0,
			Earnings:     0,
		}

		mockCourierStore.EXPECT().Get(courier.CourierId).Return(&entity.CourierDto{
			CourierId:    courier.CourierId,
			CourierType:  courier.CourierType,
			Regions:      courier.Regions,
			WorkingHours: courier.WorkingHours,
		}, nil)

		mockCourierStore.EXPECT().GetCosts(courier.CourierId, "2023-05-01", "2023-05-02").Return([]int{}, nil)

		idString := strconv.FormatInt(courier.CourierId, 10)
		req := httptest.NewRequest(http.MethodGet, "/couriers/meta-info/"+idString, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/couriers/:id")
		c.SetParamNames("id")
		c.SetParamValues(idString)
		c.Request().URL.RawQuery = "start_date=2023-05-01&end_date=2023-05-02"
		err = h.GetCourierMetaInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		expectedResponse := responses.GetCourierMetaInfoResponse{*courier}
		var actualResponse responses.GetCourierMetaInfoResponse
		json.Unmarshal(rec.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse, actualResponse)
	}

	// Test the invalid request scenario
	{
		idString := strconv.FormatInt(1, 10)
		req := httptest.NewRequest(http.MethodGet, "/couriers/meta-info/"+idString, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/couriers/:id")
		c.SetParamNames("id")
		c.SetParamValues(idString)
		c.Request().URL.RawQuery = "start_date=2023-05-01&end_date=2023-13-02"
		err = h.GetCourierMetaInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

}
