package update

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"predictor/internal/lib/api/response"
	"predictor/internal/lib/logger/sLogger"
	"predictor/internal/storage"
	"strconv"
)

type Request struct {
	Id          int64  `json:"id" validate:"required,id"`
	Name        string `json:"name,omitempty"`
	Surname     string `json:"surname,omitempty"`
	Patronym    string `json:"patronym,omitempty"`
	Age         int    `json:"age,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Nationality string `json:"nationality,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2 --name=PeopleUpdater
type PeopleUpdater interface {
	UpdatePeople(name, surname, patronym, gender, nationality string, age int, id int64) error
}

// New @Summary Update person
// @Description Update person by ID with any of their information (partial or full)
// @Tags People
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Param req body Request true "Updated person info"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /people/{id} [put]
// @Router /people/{id} [patch]
func New(log *slog.Logger, peopleUpdater PeopleUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.people.update.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		strId := chi.URLParam(r, "id")

		id, err := strconv.ParseInt(strId, 10, 64)
		if err != nil || id == 0 {
			log.Info("id is invalid")

			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		log.Info("URL params read")

		var req Request

		if err = render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request", sLogger.Error(err))

			render.JSON(w, r, response.Error("internal server error"))

			return
		}

		log.Info("request decoded", slog.Any("request", req))

		err = peopleUpdater.UpdatePeople(req.Name, req.Surname, req.Patronym, req.Gender, req.Nationality, req.Age, id)
		if errors.Is(err, storage.ErrPeopleNotFound) {
			log.Info("people not found", "id", id)

			render.JSON(w, r, response.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to update people", sLogger.Error(err))

			render.JSON(w, r, response.Error("internal server error"))

			return
		}

		log.Info("people updated")

		render.JSON(w, r, response.OK())
	}
}
