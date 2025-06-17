package router

import (
	"context"
	"github.com/carinfinin/keeper/internal/jwtr"
	"github.com/carinfinin/keeper/internal/logger"
	"net/http"
	"strings"
)

type keyUserID string

const UserData keyUserID = "userData"

func (r *Router) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		ah := request.Header.Get("Authorization")
		token := strings.Replace(ah, "Bearer ", "", 1)
		if token == "" {
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			return
		}

		data, err := jwtr.Decode(token, r.Config)
		if err != nil {
			logger.Log.Error(err)
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logger.Log.Debug("AuthMiddleware UserId: ", data)
		ctx := context.WithValue(request.Context(), UserData, data)
		newReq := request.WithContext(ctx)
		next.ServeHTTP(writer, newReq)
	})
}
