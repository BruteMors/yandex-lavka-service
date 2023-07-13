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

type Courier struct {
	service  *service.Courier
	validate *validator.Validate
}

// NewCourier - create new handler courier instance
func NewCourier(service *service.Courier, validate *validator.Validate) *Courier {
	return &Courier{service: service, validate: validate}
}

// AddCouriers - add couriers to database
func (c *Courier) AddCouriers(e echo.Context) error {
	log.Debug("add couriers by POST /couriers")
	couriersDTO := new(entity.CreateCourierRequest)
	badRequestResponse := responses.BadRequestResponse{}
	if err := e.Bind(couriersDTO); err != nil {
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	err := c.validate.Struct(couriersDTO)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	couriers := make([]entity.Courier, len(couriersDTO.Couriers))
	for i, dto := range couriersDTO.Couriers {
		couriers[i] = c.transferCourierDTOToCourier(dto)
	}

	addedCouriers, err := c.service.AddCouriers(couriers...)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	createCouriersResponse := responses.CreateCouriersResponse{
		Couriers: addedCouriers,
	}

	return e.JSON(http.StatusOK, createCouriersResponse)
}

// GetCourier - get info about courier by courier id
func (c *Courier) GetCourier(e echo.Context) error {
	log.Debug("get courier by GET /couriers/:id")
	badRequestResponse := responses.BadRequestResponse{}
	courierId := e.Param("id")

	courierIdInt, err := strconv.ParseInt(courierId, 10, 64)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	err = c.validate.Var(courierIdInt, "numeric,gt=0")
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	courier, err := c.service.GetCourier(courierIdInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info(err)
			notFoundResponse := responses.NotFoundResponse{}
			return e.JSON(http.StatusNotFound, notFoundResponse)
		}
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}
	getCourierResponse := responses.GetCourierResponse{CourierDto: *courier}
	return e.JSON(http.StatusOK, getCourierResponse)
}

// GetCouriers - get info about all couriers
func (c *Courier) GetCouriers(e echo.Context) error {
	log.Debug("get couriers by GET /couriers")
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

	couriers, err := c.service.GetCouriers(limitInt, offsetInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			getCouriersResponse := responses.GetCouriersResponse{
				Couriers: &[]entity.CourierDto{},
				Limit:    limitInt,
				Offset:   offsetInt,
			}

			return e.JSON(http.StatusOK, getCouriersResponse)
		}
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	getCouriersResponse := responses.GetCouriersResponse{
		Couriers: couriers,
		Limit:    limitInt,
		Offset:   offsetInt,
	}

	return e.JSON(http.StatusOK, getCouriersResponse)
}

// GetCourierMetaInfo - get meta info about courier by courier id
func (c *Courier) GetCourierMetaInfo(e echo.Context) error {
	log.Debug("get courier meta info by GET /couriers/meta-info/:id")
	badRequestResponse := responses.BadRequestResponse{}
	courierId := e.Param("id")

	courierIdInt, err := strconv.ParseInt(courierId, 10, 64)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	err = c.validate.Var(courierIdInt, "numeric,gt=0")
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	startDate := e.QueryParam("start_date")
	endDate := e.QueryParam("end_date")

	err = c.validate.Var(startDate, "gt=0,datetime=2006-01-02")
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	err = c.validate.Var(endDate, "gt=0,datetime=2006-01-02")
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}

	courier, err := c.service.GetCourierMetaInfo(courierIdInt, startDate, endDate)
	if err != nil {
		log.Info(err)
		return e.JSON(http.StatusBadRequest, badRequestResponse)
	}
	getCourierMetaInfoResponse := responses.GetCourierMetaInfoResponse{CourierMetaInfo: *courier}
	return e.JSON(http.StatusOK, getCourierMetaInfoResponse)
}

// transferCourierDTOToCourier - func for transfer CreateCourierDTO struct to Courier struct
func (c *Courier) transferCourierDTOToCourier(dto entity.CreateCourierDTO) entity.Courier {
	return entity.Courier{
		CourierType:  dto.CourierType,
		Regions:      dto.Regions,
		WorkingHours: dto.WorkingHours,
	}
}
