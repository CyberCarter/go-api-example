package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pascaldekloe/jwt"
)

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		next.ServeHTTP(w, r)
	})
}

func (app *application) checkToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			// allow annon users to access secure route or not
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			log.Println(headerParts)
			app.errorJSON(w, errors.New("invalid auth header"))
			return
		}

		if headerParts[0] != "Bearer" {
			app.errorJSON(w, errors.New("unauthorized - no bearer token"))
			return
		}

		token := headerParts[1]

		claims, err := jwt.HMACCheck([]byte(token), []byte(app.config.jwt.secret))
		if err != nil {
			app.errorJSON(w, errors.New("unauthorized - failed HMAC check"), http.StatusForbidden)
			return
		}

		// check if the token is valid at this moment in time
		if !claims.Valid(time.Now()) {
			app.errorJSON(w, errors.New("unauthorized - token expired"), http.StatusForbidden)
			return
		}

		// who issued the token
		if !claims.AcceptAudience("mydomain.com") {
			app.errorJSON(w, errors.New("unauthorized - invalid audience"), http.StatusForbidden)
			return
		}

		// check issuer
		if claims.Issuer != "mydomain.com" {
			app.errorJSON(w, errors.New("unauthorized - invalid issuer"), http.StatusForbidden)
			return
		}

		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			app.errorJSON(w, errors.New("unauthorized - no user ID"), http.StatusForbidden)
			return
		}

		log.Println("Valid user:", userID)

		next.ServeHTTP(w, r)
	})
}
