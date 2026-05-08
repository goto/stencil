package api_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goto/stencil/core/schema"
)

func TestGetImpactedSchemasHandler(t *testing.T) {
	nsName := "namespace1"
	schemaName := "MySchema"
	endpoint := fmt.Sprintf("/v1beta1/namespaces/%s/schemas/%s/impact", nsName, schemaName)

	t.Run("should return 500 when service returns error", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		schemaSvc.On("GetImpactedSchemas", mock.Anything, nsName, schemaName, mock.Anything, 10).
			Return(nil, errors.New("service error"))

		body := []byte(`{"fields":[{"name":"id","type":"int32","change":"REMOVED"}]}`)
		req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should return 200 with impact response on success", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		impactResp := &schema.ImpactResponse{
			RootSchema: schema.RootSchemaRef{
				NamespaceID: nsName,
				SchemaName:  schemaName,
			},
			ImpactedSchemas: []schema.ImpactedSchema{
				{
					NamespaceID: nsName,
					SchemaName:  "DependentSchema",
					ChangeType:  schema.ChangeTypeFieldRemoved,
					Depth:       1,
					ImportPath:  []string{schemaName, "DependentSchema"},
				},
			},
			Summary: schema.ImpactSummary{
				TotalImpacted:    1,
				BreakingCount:    1,
				NonBreakingCount: 0,
			},
		}

		schemaSvc.On("GetImpactedSchemas", mock.Anything, nsName, schemaName, mock.Anything, 10).
			Return(impactResp, nil)

		body := []byte(`{"fields":[{"name":"id","type":"int32","change":"REMOVED"}]}`)
		req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "DependentSchema")
		assert.Contains(t, w.Body.String(), "FIELD_REMOVED")
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should use custom max_depth from query param", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		impactResp := &schema.ImpactResponse{
			RootSchema:      schema.RootSchemaRef{NamespaceID: nsName, SchemaName: schemaName},
			ImpactedSchemas: []schema.ImpactedSchema{},
			Summary:         schema.ImpactSummary{},
		}

		schemaSvc.On("GetImpactedSchemas", mock.Anything, nsName, schemaName, mock.Anything, 3).
			Return(impactResp, nil)

		body := []byte(`{"fields":[]}`)
		req, _ := http.NewRequest("POST", endpoint+"?max_depth=3", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should fall back to default max_depth=10 when query param is invalid", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		impactResp := &schema.ImpactResponse{
			RootSchema:      schema.RootSchemaRef{NamespaceID: nsName, SchemaName: schemaName},
			ImpactedSchemas: []schema.ImpactedSchema{},
			Summary:         schema.ImpactSummary{},
		}

		schemaSvc.On("GetImpactedSchemas", mock.Anything, nsName, schemaName, mock.Anything, 10).
			Return(impactResp, nil)

		body := []byte(`{"fields":[]}`)
		req, _ := http.NewRequest("POST", endpoint+"?max_depth=invalid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should fall back to default max_depth=10 when max_depth is zero", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		impactResp := &schema.ImpactResponse{
			RootSchema:      schema.RootSchemaRef{NamespaceID: nsName, SchemaName: schemaName},
			ImpactedSchemas: []schema.ImpactedSchema{},
			Summary:         schema.ImpactSummary{},
		}

		schemaSvc.On("GetImpactedSchemas", mock.Anything, nsName, schemaName, mock.Anything, 10).
			Return(impactResp, nil)

		body := []byte(`{"fields":[]}`)
		req, _ := http.NewRequest("POST", endpoint+"?max_depth=0", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		schemaSvc.AssertExpectations(t)
	})

	t.Run("should return error when body is invalid JSON", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		body := []byte(`not-json`)
		req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		schemaSvc.AssertNotCalled(t, "GetImpactedSchemas")
	})

	t.Run("should handle multiple field changes in request", func(t *testing.T) {
		_, schemaSvc, _, mux, _, _ := setup()

		impactResp := &schema.ImpactResponse{
			RootSchema:      schema.RootSchemaRef{NamespaceID: nsName, SchemaName: schemaName},
			ImpactedSchemas: []schema.ImpactedSchema{},
			Summary:         schema.ImpactSummary{},
		}

		schemaSvc.On("GetImpactedSchemas", mock.Anything, nsName, schemaName, mock.Anything, 10).
			Return(impactResp, nil)

		body := []byte(`{"fields":[
			{"name":"field1","type":"string","change":"ADDED"},
			{"name":"field2","type":"int32","change":"REMOVED"},
			{"name":"field3","type":"bool","change":"TYPE_CHANGED"}
		]}`)
		req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		schemaSvc.AssertExpectations(t)
	})
}
