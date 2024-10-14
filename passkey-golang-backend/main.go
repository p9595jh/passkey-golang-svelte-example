package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"passkey-golang-backend/cache"
	"passkey-golang-backend/dto"
	"passkey-golang-backend/entity"
	"passkey-golang-backend/logger"
	"passkey-golang-backend/repository"
	"passkey-golang-backend/resp"
	"passkey-golang-backend/utils"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/rs/cors"
)

type sessionData struct {
	Session *webauthn.SessionData `json:"session"`
	User    *entity.User          `json:"user"`
}

var (
	webAuthn *webauthn.WebAuthn
	repo     *repository.Repository // representing database repository
	scache   *cache.Cache           // representing cache server for session
)

func main() {
	// configuration for RP
	var (
		rpProtocol = "http"
		rpHost     = "localhost"
		rpPort     = ":5173"
		err        error
	)

	wconfig := &webauthn.Config{
		RPDisplayName: "golang passkey example",
		RPID:          rpHost,
		RPOrigins:     []string{fmt.Sprintf("%s://%s%s", rpProtocol, rpHost, rpPort)},
	}

	webAuthn, err = webauthn.New(wconfig)
	if err != nil {
		panic(err)
	}

	repo = repository.New()
	scache = cache.New()

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/api/passkey/register/start", BeginRegistration)
	serveMux.HandleFunc("/api/passkey/register/finish", FinishRegistration)
	serveMux.HandleFunc("/api/passkey/login/start", BeginLogin)
	serveMux.HandleFunc("/api/passkey/login/finish", FinishLogin)
	serveMux.HandleFunc("/api/forbidden", ForbiddenPage)

	logger.Info().Msg("server start")
	handler := cors.Default().Handler(serveMux)
	corHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Content-Type", "X-Session-Id"},
	})
	handler = corHandler.Handler(handler)
	if err := http.ListenAndServe(":4000", handler); err != nil {
		logger.Error().Err(err).Send()
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// register
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func BeginRegistration(w http.ResponseWriter, r *http.Request) {
	// parse user request body
	var userDTO dto.UserRegistrationDTO
	if err := json.NewDecoder(r.Body).Decode(&userDTO); err != nil {
		resp.ErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	// create random user id
	userID, err := utils.RandID(64)
	if err != nil {
		resp.ErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// create new user with the given name and generated id
	// user := entity.NewUser(userID, userDTO.Name)
	user := &entity.User{
		ID:        userID,
		Name:      userDTO.Name,
		Email:     userDTO.Email,
		BirthYear: userDTO.BirthYear,
	}
	// begin registration with the created user
	options, session, err := webAuthn.BeginRegistration(user)
	if err != nil {
		resp.ErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// marshal registration session and user to store into the cache
	sessionJSON, err := json.Marshal(&sessionData{
		Session: session,
		User:    user,
	})
	if err != nil {
		resp.ErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// store into the cache server with new uuid
	// need to finish registration until 1 hour
	sessionID := uuid.NewString()
	scache.Put(sessionID, sessionJSON, time.Hour)

	resp.JSONResponse(w, http.StatusOK, &dto.BeginRegistrationDTO{
		Options: options,
		SID:     sessionID,
	})
}

func FinishRegistration(w http.ResponseWriter, r *http.Request) {
	// get session id from header
	sid := r.Header.Get("X-Session-Id")
	if sid == "" {
		resp.ErrorResponse(w, r, http.StatusBadRequest, errors.New("session id required"))
		return
	}

	// get session json from the cache by session id
	sessionJSON, exists := scache.Get(sid)
	if !exists {
		resp.ErrorResponse(w, r, http.StatusNotFound, errors.New("session not found"))
		return
	}

	// unmarshal session json
	var sd sessionData
	if err := json.Unmarshal(sessionJSON, &sd); err != nil {
		resp.ErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unmarshal: %w", err))
		return
	}

	// finish registration
	credential, err := webAuthn.FinishRegistration(sd.User, *sd.Session, r)
	if err != nil {
		// type assertion to protocol.Error and print detail
		if perr, ok := err.(*protocol.Error); ok {
			logger.Error().Any("error", perr).Send()
		}
		resp.ErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("finishRegistration: %w", err))
		return
	}

	// store in the database and remove from the cache
	sd.User.AddCredential(credential)
	repo.Save(sd.User)
	scache.Delete(sid)

	resp.JSONResponse(w, http.StatusOK, &dto.FinishRegistrationDTO{
		Message: "registration success",
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// login
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func BeginLogin(w http.ResponseWriter, r *http.Request) {
	var userDTO dto.UserLoginDTO
	if err := json.NewDecoder(r.Body).Decode(&userDTO); err != nil {
		resp.ErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	user := repo.Find(userDTO.Name)
	options, session, err := webAuthn.BeginLogin(user)
	if err != nil {
		resp.JSONResponse(w, http.StatusBadRequest, err)
		return
	}

	sessionJSON, err := json.Marshal(&sessionData{
		Session: session,
		User:    user,
	})
	if err != nil {
		resp.ErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// store into the cache server with new uuid
	// need to finish login until 1 hour
	sessionID := uuid.NewString()
	scache.Put(sessionID, sessionJSON, time.Hour)

	resp.JSONResponse(w, http.StatusOK, &dto.BeginLoginDTO{
		Options: options,
		SID:     sessionID,
	})
}

func FinishLogin(w http.ResponseWriter, r *http.Request) {
	sid := r.Header.Get("X-Session-Id")
	if sid == "" {
		resp.ErrorResponse(w, r, http.StatusBadRequest, errors.New("session id required"))
		return
	}

	sessionJSON, exists := scache.Get(sid)
	if !exists {
		resp.ErrorResponse(w, r, http.StatusNotFound, errors.New("session not found"))
		return
	}

	var sd sessionData
	if err := json.Unmarshal(sessionJSON, &sd); err != nil {
		resp.ErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unmarshal: %w", err))
		return
	}

	credential, err := webAuthn.FinishLogin(sd.User, *sd.Session, r)
	if err := json.Unmarshal(sessionJSON, &sd); err != nil {
		resp.ErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unmarshal: %w", err))
		return
	}

	// Handle credential.Authenticator.CloneWarning
	if credential.Authenticator.CloneWarning {
		logger.Warn().Msgf("[WARN] can't finish login: %s", "CloneWarning")
	}

	sd.User.AddCredential(credential)
	repo.Save(sd.User)
	scache.Delete(sid)

	loginSessionJSON, err := json.Marshal(&sessionData{
		User: sd.User,
	})
	if err != nil {
		resp.ErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// store into the cache server with new uuid
	// login session sustained until 1 hour
	sessionID := uuid.NewString()
	scache.Put(sessionID, loginSessionJSON, time.Hour)

	resp.JSONResponse(w, http.StatusOK, &dto.FinishLoginDTO{
		SID: sessionID,
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// forbidden page
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func ForbiddenPage(w http.ResponseWriter, r *http.Request) {
	sid := r.Header.Get("X-Session-Id")
	if sid == "" {
		resp.ErrorResponse(w, r, http.StatusBadRequest, errors.New("session id required"))
		return
	}

	sessionJSON, exists := scache.Get(sid)
	if !exists {
		resp.ErrorResponseWithText(w, r, http.StatusForbidden)
		return
	}

	var sd sessionData
	if err := json.Unmarshal(sessionJSON, &sd); err != nil {
		resp.ErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unmarshal: %w", err))
		return
	}

	resp.JSONResponse(w, http.StatusOK, map[string]any{
		"name":      sd.User.Name,
		"email":     sd.User.Email,
		"birthYear": sd.User.BirthYear,
	})
}
