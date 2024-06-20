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

func TestHTTPGetSchema(t *testing.T) {
	nsName := "namespace1"
	schemaName := "scName"
	t.Run("should validate version number", func(t *testing.T) {
		_, _, _, mux, _, newrelic := setup()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/v1beta1/namespaces/%s/schemas/%s/versions/invalidNumber", nsName, schemaName), nil)
		var called bool
		newrelic.On("StartGenericSegment", mock.Anything, "GetSchema").Return(func() { called = true })
		mux.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
		assert.JSONEq(t, `{"code":2,"message":"invalid version number","details":[]}`, w.Body.String())
		newrelic.AssertExpectations(t)
		assert.True(t, called)
	})
	t.Run("should return http error if getSchema fails", func(t *testing.T) {
		version := int32(2)
		_, schemaSvc, _, mux, _, newrelic := setup()
		var called bool
		newrelic.On("StartGenericSegment", mock.Anything, "GetSchema").Return(func() { called = true })
		schemaSvc.On("Get", mock.Anything, nsName, schemaName, version).Return(nil, nil, errors.New("get error"))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/v1beta1/namespaces/%s/schemas/%s/versions/%d", nsName, schemaName, version), nil)
		mux.ServeHTTP(w, req)
		assert.Equal(t, 500, w.Code)
		assert.JSONEq(t, `{"code":2,"message":"get error","details":[]}`, w.Body.String())
		newrelic.AssertExpectations(t)
		assert.True(t, called)
	})
	t.Run("should return octet-stream content type for protobuf schema", func(t *testing.T) {
		version := int32(2)
		data := []byte("test data")
		_, schemaSvc, _, mux, _, newrelic := setup()
		var called bool
		newrelic.On("StartGenericSegment", mock.Anything, "GetSchema").Return(func() { called = true })
		schemaSvc.On("Get", mock.Anything, nsName, schemaName, version).Return(&schema.Metadata{Format: "FORMAT_PROTOBUF"}, data, nil)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/v1beta1/namespaces/%s/schemas/%s/versions/%d", nsName, schemaName, version), nil)
		mux.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, data, w.Body.Bytes())
		assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
		newrelic.AssertExpectations(t)
		assert.True(t, called)
	})
}

func TestHTTPSchemaCreate(t *testing.T) {
	nsName := "namespace"
	scName := "schemaName"
	format := "PROTOBUF"
	compatibility := "FULL"
	sourceUrl := "https://source.golabs.io/asgard/data-change-management-worker"
	commitSHA := "commit-sha"
	body := []byte("protobuf contents")
	metadata := &schema.Metadata{Format: format, Compatibility: compatibility, SourceURL: sourceUrl}
	t.Run("should return error if schema create fails", func(t *testing.T) {
		_, schemaSvc, _, mux, _, newrelic := setup()
		var called bool
		newrelic.On("StartGenericSegment", mock.Anything, "UploadSchema").Return(func() { called = true })
		schemaSvc.On("Create", mock.Anything, nsName, scName, metadata, body, commitSHA).Return(schema.SchemaInfo{}, errors.New("create error"))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", fmt.Sprintf("/v1beta1/namespaces/%s/schemas/%s", nsName, scName), bytes.NewBuffer(body))
		req.Header.Add("X-Format", format)
		req.Header.Add("X-Compatibility", compatibility)
		req.Header.Add("X-SourceURL", sourceUrl)
		req.Header.Add("X-CommitSHA", commitSHA)
		mux.ServeHTTP(w, req)
		assert.Equal(t, 500, w.Code)
		schemaSvc.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
	})
	t.Run("should return schemaInfo in JSON after create", func(t *testing.T) {
		_, schemaSvc, _, mux, _, newrelic := setup()
		scInfo := schema.SchemaInfo{ID: "someID", Version: int32(2)}
		var called bool
		newrelic.On("StartGenericSegment", mock.Anything, "UploadSchema").Return(func() { called = true })
		schemaSvc.On("Create", mock.Anything, nsName, scName, metadata, body, commitSHA).Return(scInfo, nil)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", fmt.Sprintf("/v1beta1/namespaces/%s/schemas/%s", nsName, scName), bytes.NewBuffer(body))
		req.Header.Add("X-Format", format)
		req.Header.Add("X-Compatibility", compatibility)
		req.Header.Add("X-SourceURL", sourceUrl)
		req.Header.Add("X-CommitSHA", commitSHA)
		mux.ServeHTTP(w, req)
		assert.Equal(t, 201, w.Code)
		assert.JSONEq(t, `{"id": "someID", "location": "", "version": 2}`, w.Body.String())
		schemaSvc.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
	})
}
