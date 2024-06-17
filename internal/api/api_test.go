package api_test

import (
	"github.com/goto/stencil/internal/api"
	"github.com/goto/stencil/internal/api/mocks"
	mocks2 "github.com/goto/stencil/pkg/newrelic/mocks"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func setup() (*mocks.NamespaceService, *mocks.SchemaService, *mocks.SearchService, *runtime.ServeMux, *mocks.ChangeDetectorService, *api.API, *mocks2.NewRelic) {
	nsService := &mocks.NamespaceService{}
	schemaService := &mocks.SchemaService{}
	searchService := &mocks.SearchService{}
	changeDetectorService := &mocks.ChangeDetectorService{}
	newRelic := &mocks2.NewRelic{}
	mux := runtime.NewServeMux()
	v1beta1 := api.NewAPI(nsService, schemaService, searchService, changeDetectorService, newRelic)
	v1beta1.RegisterSchemaHandlers(mux, nil)
	return nsService, schemaService, searchService, mux, changeDetectorService, v1beta1, newRelic
}
