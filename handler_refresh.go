package main

import (
	"net/http"
	"time"

	"github.com/C4triplezero/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting token", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting refresh token", err)
		return
	}
	if refreshToken.RevokedAt.Valid || refreshToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is revoked or expired", nil)
		return
	}

	token, err := auth.MakeJWT(refreshToken.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating new token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{Token: token})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting token", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking refresh token", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
