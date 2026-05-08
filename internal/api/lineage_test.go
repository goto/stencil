package api_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goto/stencil/core/schema"
)

func TestGetLineageHandler(t *testing.T) {
	nsName := "namespace1"
	schemaName := "MySchema"
	endpoint := fmt.Sprintf("/v1beta1/namespaces/%s/schemas/%s/lineage", nsName, schemaName)

	t.Run("should return 500 when service returns error", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		schemaSvc.On("GetLineage", mock.Anything, nsName, schemaName, 10, schema.LineageDirectionBoth).
			Return(nil, errors.New("service error"))

		req, _ := http.NewRequest("GET", endpoint, nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should return 200 with lineage response on success", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		lineageResp := &schema.LineageResponse{
			RootSchema: schema.RootSchemaRef{
				NamespaceID: nsName,
				SchemaName:  schemaName,
			},
			Direction: schema.LineageDirectionBoth,
			Downstream: []schema.LineageSchema{
				{
					NamespaceID: nsName,
					SchemaName:  "DependentSchema",
					Level:       1,
					Path:        []string{schemaName, "DependentSchema"},
				},
			},
			Summary: schema.LineageSummary{
				DownstreamCount: 1,
				TotalCount:      1,
			},
		}

		schemaSvc.On("GetLineage", mock.Anything, nsName, schemaName, 10, schema.LineageDirectionBoth).
			Return(lineageResp, nil)

		req, _ := http.NewRequest("GET", endpoint, nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "DependentSchema")
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should use custom level from query param", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		lineageResp := &schema.LineageResponse{
			RootSchema: schema.RootSchemaRef{NamespaceID: nsName, SchemaName: schemaName},
			Direction:  schema.LineageDirectionBoth,
			Summary:    schema.LineageSummary{},
		}

		schemaSvc.On("GetLineage", mock.Anything, nsName, schemaName, 3, schema.LineageDirectionBoth).
			Return(lineageResp, nil)

		req, _ := http.NewRequest("GET", endpoint+"?level=3", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should fall back to default level=10 when query param is invalid", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		lineageResp := &schema.LineageResponse{
			RootSchema: schema.RootSchemaRef{NamespaceID: nsName, SchemaName: schemaName},
			Direction:  schema.LineageDirectionBoth,
			Summary:    schema.LineageSummary{},
		}

		schemaSvc.On("GetLineage", mock.Anything, nsName, schemaName, 10, schema.LineageDirectionBoth).
			Return(lineageResp, nil)

		req, _ := http.NewRequest("GET", endpoint+"?level=invalid", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should fall back to default level=10 when level is zero", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		lineageResp := &schema.LineageResponse{
			RootSchema: schema.RootSchemaRef{NamespaceID: nsName, SchemaName: schemaName},
			Direction:  schema.LineageDirectionBoth,
			Summary:    schema.LineageSummary{},
		}

		schemaSvc.On("GetLineage", mock.Anything, nsName, schemaName, 10, schema.LineageDirectionBoth).
			Return(lineageResp, nil)

		req, _ := http.NewRequest("GET", endpoint+"?level=0", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should use downstream direction when provided", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		lineageResp := &schema.LineageResponse{
			RootSchema: schema.RootSchemaRef{NamespaceID: nsName, SchemaName: schemaName},
			Direction:  schema.LineageDirectionDownstream,
			Summary:    schema.LineageSummary{},
		}

		schemaSvc.On("GetLineage", mock.Anything, nsName, schemaName, 10, schema.LineageDirectionDownstream).
			Return(lineageResp, nil)

		req, _ := http.NewRequest("GET", endpoint+"?direction=downstream", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should use upstream direction when provided", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		lineageResp := &schema.LineageResponse{
			RootSchema: schema.RootSchemaRef{NamespaceID: nsName, SchemaName: schemaName},
			Direction:  schema.LineageDirectionUpstream,
			Summary:    schema.LineageSummary{},
		}

		schemaSvc.On("GetLineage", mock.Anything, nsName, schemaName, 10, schema.LineageDirectionUpstream).
			Return(lineageResp, nil)

		req, _ := http.NewRequest("GET", endpoint+"?direction=upstream", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should return 400 when direction is invalid", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		req, _ := http.NewRequest("GET", endpoint+"?direction=sideways", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
		schemaSvc.AssertNotCalled(t, "GetLineage")
	})
}
