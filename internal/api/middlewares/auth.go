package middlewares

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/services"
	"go.uber.org/zap"
)

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		pass := req.Context().Value(config.CtxPassKey{}).(string)

		jwt, isSet := services.GetHeaderJWT(req.Header)

		var userID *uuid.UUID
		var err error
		if isSet {
			userID, err = services.GetUserIDJWT(jwt, pass)
			if err != nil {
				zap.S().Infoln("Can't get user UUID form JWT.", err)
				isSet = false
				userID = nil
			}
		}

		ctx := context.WithValue(req.Context(), config.CtxConfig{}, config.NewCtxConfig(userID, isSet))
		h.ServeHTTP(res, req.WithContext(ctx))

	})

}