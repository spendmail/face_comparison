package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
)

const (
	URLComparePattern = "/compare/"
)

type Config interface {
	GetHTTPHost() string
	GetHTTPPort() string
	GetSecret() string
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Application interface {
	CompareImages(urls []string) (string, []string, []error)
}

type Server struct {
	Logger Logger
	Server *http.Server
}

type Handler struct {
	Config Config
	App    Application
	Logger Logger
}

func New(config Config, logger Logger, app Application) *Server {
	handler := &Handler{
		Config: config,
		App:    app,
		Logger: logger,
	}

	router := mux.NewRouter()
	router.HandleFunc(URLComparePattern, handler.compareHandler).Methods(http.MethodPost)

	server := &http.Server{
		Addr:    net.JoinHostPort(config.GetHTTPHost(), config.GetHTTPPort()),
		Handler: router,
	}

	return &Server{
		Logger: logger,
		Server: server,
	}
}

type ComparisonRequest struct {
	Secret string   `json:"secret"`
	URLs   []string `json:"urls"`
}

type ComparisonResponse struct {
	Target    string   `json:"target"`
	Unmatched []string `json:"unmatched"`
	Errors    []string `json:"errors"`
}

var (
	ErrWrongSecret = errors.New("wrong secret code")
)

func (h *Handler) compareHandler(w http.ResponseWriter, r *http.Request) {
	var cr ComparisonRequest
	rsp := ComparisonResponse{
		Unmatched: make([]string, 0),
		Errors:    make([]string, 0),
	}

	// request decoding
	err := json.NewDecoder(r.Body).Decode(&cr)
	if err != nil {
		rsp.Errors = []string{fmt.Sprintf("unable to decode the response: %s", err.Error())}
		SendComparisonResponse(w, h, rsp)
		return
	}

	//secret checking
	if cr.Secret != h.Config.GetSecret() {
		rsp.Errors = []string{ErrWrongSecret.Error()}
		SendComparisonResponse(w, h, rsp)
		return
	}

	// images processing
	source, unmatched, errs := h.App.CompareImages(cr.URLs)

	// converting errors to string
	strErrs := make([]string, len(errs))
	for i, err := range errs {
		strErrs[i] = err.Error()
	}

	// renaming target as a source
	rsp.Target = source
	rsp.Unmatched = unmatched
	rsp.Errors = strErrs

	SendComparisonResponse(w, h, rsp)
}

func SendComparisonResponse(w http.ResponseWriter, h *Handler, rsp ComparisonResponse) {

	// for testing purposes logging all the errors occurred
	for _, err := range rsp.Errors {
		h.Logger.Error(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(rsp)
	if err != nil {
		h.Logger.Error(err)
	}
}

func (s *Server) Start() error {
	return s.Server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}