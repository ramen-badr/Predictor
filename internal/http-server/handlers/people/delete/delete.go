package delete

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

//go:generate go run github.com/vektra/mockery/v2 --name=PeopleDeleter
type PeopleDeleter interface {
	DeletePeople(id int64) error
}

// New @Summary Delete person
// @Description Delete person by ID
// @Tags People
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /people/{id} [delete]
func New(log *slog.Logger, peopleDeleter PeopleDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.people.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		strId := chi.URLParam(r, "id")

		id, err := strconv.ParseInt(strId, 10, 64)
		if err != nil {
			log.Info("id is invalid")

			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		log.Info("URL params read")

		err = peopleDeleter.DeletePeople(id)
		if errors.Is(err, storage.ErrPeopleNotFound) {
			log.Info("people not found", "id", id)

			render.JSON(w, r, response.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to delete people", sLogger.Error(err))

			render.JSON(w, r, response.Error("internal server error"))

			return
		}

		log.Info("people deleted")

		render.JSON(w, r, response.OK())
	}
}
