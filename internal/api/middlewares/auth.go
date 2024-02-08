package middlewares

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/shulganew/gophermart/internal/entities"
	"github.com/shulganew/gophermart/internal/services"
	"go.uber.org/zap"
)

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		passVal := req.Context().Value(entities.CtxPassKey{})
		pass, ok := passVal.(string)
		if !ok {
			zap.S().Errorln("Can't git pass key from context.")
			h.ServeHTTP(res, req)
			return
		}
		jwt, isSet := services.GetHeaderJWT(req.Header)

		var userID uuid.UUID
		var err error
		if isSet {
			userID, err = services.GetUserIDJWT(jwt, pass)
			if err != nil {
				zap.S().Infoln("Can't get user UUID form JWT.", err)
				isSet = false
			}
		}
		ctx := context.WithValue(req.Context(), entities.MiddlwDTO{}, entities.NewMiddlwDTO(userID, isSet))
		h.ServeHTTP(res, req.WithContext(ctx))
	})
}
