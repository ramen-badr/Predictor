package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"predictor/internal/lib/api"
	"predictor/internal/lib/api/response"
	"predictor/internal/lib/logger/sLogger"
)

type Request struct {
	Name     string `json:"name" validate:"required"`
	Surname  string `json:"surname" validate:"required"`
	Patronym string `json:"patronym,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2 --name=PeopleSaver
type PeopleSaver interface {
	SavePeople(name, surname, patronym, gender, nationality string, age int) (int64, error)
}

// New @Summary Save person
// @Description Save person by name, surname and optional patronym
// @Tags People
// @Accept json
// @Produce json
// @Param req body Request true "Name, surname and optional patronym"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /people [post]
func New(log *slog.Logger, peopleSaver PeopleSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.people.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", sLogger.Error(err))

			render.JSON(w, r, response.Error("internal server error"))

			return
		}

		log.Info("request decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors

			errors.As(err, &validateErr)

			log.Error("invalid request", sLogger.Error(err))

			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		age := api.GetAge(req.Name)
		if age == 0 {
			log.Error("failed to get age", sLogger.Error(err))

			render.JSON(w, r, response.Error("internal server error"))

			return
		}

		gender := api.GetGender(req.Name)
		if gender == "" {
			log.Error("failed to get gender", sLogger.Error(err))

			render.JSON(w, r, response.Error("internal server error"))

			return
		}

		nationality := api.GetNationality(req.Name)
		if nationality == "" {
			log.Error("failed to get nationality", sLogger.Error(err))

			render.JSON(w, r, response.Error("internal server error"))

			return
		}

		id, err := peopleSaver.SavePeople(req.Name, req.Surname, req.Patronym, gender, nationality, age)
		if err != nil {
			log.Error("failed to save people", sLogger.Error(err))

			render.JSON(w, r, response.Error("internal server error"))

			return
		}

		log.Info("people saved", slog.Int64("id", id))

		render.JSON(w, r, response.OK())
	}
}
