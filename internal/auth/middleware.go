package auth

import (
	"net/http"
	"strings"
)

func JWTMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			//1. Берем Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			//2. Проверяем формат: Bearer <token>
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]

			//3. Проверяем JWT
			userID, err := jwtManager.Verify(tokenStr)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			//4. Кладет user_id в context
			ctx := WithUserID(r.Context(), userID)

			//5. Передаем управление дальше
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
