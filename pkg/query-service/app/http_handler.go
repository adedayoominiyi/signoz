package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.signoz.io/query-service/druidQuery"
	"go.signoz.io/query-service/godruid"
	"go.uber.org/zap"
)

// NewRouter creates and configures a Gorilla Router.
func NewRouter() *mux.Router {
	return mux.NewRouter().UseEncodedPath()
}

// APIHandler implements the query service public API by registering routes at httpPrefix
type APIHandler struct {
	// queryService *querysvc.QueryService
	// queryParser  queryParser
	basePath  string
	apiPrefix string
	client    *godruid.Client
	sqlClient *druidQuery.SqlClient
}

// NewAPIHandler returns an APIHandler
func NewAPIHandler(client *godruid.Client, sqlClient *druidQuery.SqlClient) *APIHandler {
	aH := &APIHandler{
		client:    client,
		sqlClient: sqlClient,
	}

	return aH
}

type structuredResponse struct {
	Data   interface{}       `json:"data"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
	Errors []structuredError `json:"errors"`
}

type structuredError struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg"`
	// TraceID ui.TraceID `json:"traceID,omitempty"`
}

// RegisterRoutes registers routes for this handler on the given router
func (aH *APIHandler) RegisterRoutes(router *mux.Router) {

	router.HandleFunc("/api/v1/get_percentiles", aH.getApplicationPercentiles).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/services", aH.getServices).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/services/list", aH.getServicesList).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/service/overview", aH.getServiceOverview).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/service/{service}/operations", aH.getOperations).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/service/top_endpoints", aH.getTopEndpoints).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/spans", aH.searchSpans).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/spans/aggregates", aH.searchSpansAggregates).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tags", aH.searchTags).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/traces/{traceId}", aH.searchTraces).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/usage", aH.getUsage).Methods(http.MethodGet)
}

func (aH *APIHandler) getOperations(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	serviceName := vars["service"]

	var err error
	if len(serviceName) == 0 {
		err = fmt.Errorf("service param not found")
	}
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	result, err := druidQuery.GetOperations(aH.sqlClient, serviceName)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)

}

func (aH *APIHandler) getServicesList(w http.ResponseWriter, r *http.Request) {

	result, err := druidQuery.GetServicesList(aH.sqlClient)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)

}

func (aH *APIHandler) searchTags(w http.ResponseWriter, r *http.Request) {

	serviceName := r.URL.Query().Get("service")

	result, err := druidQuery.GetTags(aH.sqlClient, serviceName)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)

}

func (aH *APIHandler) getTopEndpoints(w http.ResponseWriter, r *http.Request) {

	query, err := parseGetTopEndpointsRequest(r)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	result, err := druidQuery.GetTopEndpoints(aH.sqlClient, query)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)

}

func (aH *APIHandler) getUsage(w http.ResponseWriter, r *http.Request) {

	query, err := parseGetUsageRequest(r)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	result, err := druidQuery.GetUsage(aH.sqlClient, query)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)

}

func (aH *APIHandler) getServiceOverview(w http.ResponseWriter, r *http.Request) {

	query, err := parseGetServiceOverviewRequest(r)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	result, err := druidQuery.GetServiceOverview(aH.sqlClient, query)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)

}

func (aH *APIHandler) getServices(w http.ResponseWriter, r *http.Request) {

	query, err := parseGetServicesRequest(r)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	result, err := druidQuery.GetServices(aH.sqlClient, query)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)
}
func (aH *APIHandler) searchTraces(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	traceId := vars["traceId"]

	result, err := druidQuery.SearchTraces(aH.client, traceId)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)

}
func (aH *APIHandler) searchSpansAggregates(w http.ResponseWriter, r *http.Request) {

	query, err := parseSearchSpanAggregatesRequest(r)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	result, err := druidQuery.SearchSpansAggregate(aH.client, query)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)
}

func (aH *APIHandler) searchSpans(w http.ResponseWriter, r *http.Request) {

	query, err := parseSpanSearchRequest(r)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	result, err := druidQuery.SearchSpans(aH.client, query)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	aH.writeJSON(w, r, result)
}

func (aH *APIHandler) getApplicationPercentiles(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)

	query, err := parseApplicationPercentileRequest(r)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}

	result, err := druidQuery.GetApplicationPercentiles(aH.client, query)
	if aH.handleError(w, err, http.StatusBadRequest) {
		return
	}
	aH.writeJSON(w, r, result)
}

func (aH *APIHandler) handleError(w http.ResponseWriter, err error, statusCode int) bool {
	if err == nil {
		return false
	}
	if statusCode == http.StatusInternalServerError {
		zap.S().Error("HTTP handler, Internal Server Error", zap.Error(err))
	}
	structuredResp := structuredResponse{
		Errors: []structuredError{
			{
				Code: statusCode,
				Msg:  err.Error(),
			},
		},
	}
	resp, _ := json.Marshal(&structuredResp)
	http.Error(w, string(resp), statusCode)
	return true
}

func (aH *APIHandler) writeJSON(w http.ResponseWriter, r *http.Request, response interface{}) {
	marshall := json.Marshal
	if prettyPrint := r.FormValue("pretty"); prettyPrint != "" && prettyPrint != "false" {
		marshall = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		}
	}
	resp, _ := marshall(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}