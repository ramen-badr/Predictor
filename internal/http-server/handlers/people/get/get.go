package get

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"predictor/internal/domain/models"
	"predictor/internal/lib/api/response"
	"predictor/internal/lib/logger/sLogger"
	"predictor/internal/storage"
	"strconv"
)

type Response struct {
	response.Response
	Data  []models.People `json:"data,omitempty"`
	Total int64           `json:"total,omitempty"`
	Limit int64           `json:"limit"`
	Page  int64           `json:"page"`
}

const (
	DefaultPage  = 1
	DefaultLimit = 10
)

//go:generate go run github.com/vektra/mockery/v2 --name=PeopleGetter
type PeopleGetter interface {
	GetPeople(limit, offset int64, name, surname, patronym, gender, nationality string, age int) ([]models.People, int64, error)
}

func responseOK(w http.ResponseWriter, r *http.Request, data []models.People, total, limit, page int64) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		Data:     data,
		Total:    total,
		Limit:    limit,
		Page:     page,
	})
}

// New @Summary Get people
// @Description Get people by filters
// @Tags People
// @Accept json
// @Produce json
// @Param name query string false "Name"
// @Param surname query string false "Surname"
// @Param patronym query string false "Patronym"
// @Param age query int false "Age"
// @Param gender query string false "Gender"
// @Param nationality query string false "Nationality"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router / [get]
func New(log *slog.Logger, peopleGetter PeopleGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.people.get.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		q := r.URL.Query()

		name := q.Get("name")
		surname := q.Get("surname")
		patronym := q.Get("patronym")
		gender := q.Get("gender")
		nationality := q.Get("nationality")
		age, err := strconv.Atoi(q.Get("age"))
		if err != nil {
			log.Error("failed to get age", sLogger.Error(err))

			render.JSON(w, r, response.Error("internal server error"))

			return
		}

		page, err := strconv.ParseInt(q.Get("page"), 10, 64)
		if err != nil || page < 1 {
			page = DefaultPage
		}

		limit, err := strconv.ParseInt(q.Get("limit"), 10, 64)
		if err != nil || limit < 1 {
			limit = DefaultLimit
		}

		offset := (page - 1) * limit

		data, total, err := peopleGetter.GetPeople(limit, offset, name, surname, patronym, gender, nationality, age)
		if err != nil {
			if !errors.Is(err, storage.ErrPeopleNotFound) {
				log.Error("failed to get people", sLogger.Error(err))

				render.JSON(w, r, response.Error("internal server error"))

				return
			}
		}

		log.Info("people got")

		responseOK(w, r, data, total, limit, page)
	}
}
