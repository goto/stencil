package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goto/stencil/core/changedetector"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/goto/stencil/core/schema"
	stencilv1beta1 "github.com/goto/stencil/proto/v1beta1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func schemaToProto(s schema.Schema) *stencilv1beta1.Schema {
	return &stencilv1beta1.Schema{
		Name:          s.Name,
		Format:        stencilv1beta1.Schema_Format(stencilv1beta1.Schema_Format_value[s.Format]),
		Compatibility: stencilv1beta1.Schema_Compatibility(stencilv1beta1.Schema_Compatibility_value[s.Compatibility]),
		Authority:     s.Authority,
	}
}

func (a *API) CreateSchema(ctx context.Context, in *stencilv1beta1.CreateSchemaRequest) (*stencilv1beta1.CreateSchemaResponse, error) {
	metadata := &schema.Metadata{Format: in.GetFormat().String(), Compatibility: in.GetCompatibility().String()}
	sc, err := a.schema.Create(ctx, in.NamespaceId, in.SchemaId, metadata, in.GetData())
	return &stencilv1beta1.CreateSchemaResponse{
		Version:  sc.Version,
		Id:       sc.ID,
		Location: sc.Location,
	}, err
}
func (a *API) HTTPUpload(w http.ResponseWriter, req *http.Request, pathParams map[string]string) error {
	endFunc := a.newrelic.StartGenericSegment(req.Context(), "UploadSchema")
	defer endFunc()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	format := req.Header.Get("X-Format")
	compatibility := req.Header.Get("X-Compatibility")

	metadata := &schema.Metadata{Format: format, Compatibility: compatibility}
	namespaceID := pathParams["namespace"]
	schemaName := pathParams["name"]
	sc, err := a.schema.Create(req.Context(), namespaceID, schemaName, metadata, data)
	if err != nil {
		return err
	}
	respData, _ := json.Marshal(sc)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respData)
	return nil
}

func (a *API) CheckCompatibility(ctx context.Context, req *stencilv1beta1.CheckCompatibilityRequest) (*stencilv1beta1.CheckCompatibilityResponse, error) {
	resp := &stencilv1beta1.CheckCompatibilityResponse{}
	err := a.schema.CheckCompatibility(ctx, req.GetNamespaceId(), req.GetSchemaId(), req.GetCompatibility().String(), req.GetData())
	return resp, err
}

func (a *API) HTTPCheckCompatibility(w http.ResponseWriter, req *http.Request, pathParams map[string]string) error {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	compatibility := req.Header.Get("X-Compatibility")
	namespaceID := pathParams["namespace"]
	schemaName := pathParams["name"]
	return a.schema.CheckCompatibility(req.Context(), namespaceID, schemaName, compatibility, data)
}

func (a *API) ListSchemas(ctx context.Context, in *stencilv1beta1.ListSchemasRequest) (*stencilv1beta1.ListSchemasResponse, error) {
	schemas, err := a.schema.List(ctx, in.Id)

	var ss []*stencilv1beta1.Schema
	for _, s := range schemas {
		ss = append(ss, schemaToProto(s))
	}
	return &stencilv1beta1.ListSchemasResponse{Schemas: ss}, err
}

func (a *API) GetLatestSchema(ctx context.Context, in *stencilv1beta1.GetLatestSchemaRequest) (*stencilv1beta1.GetLatestSchemaResponse, error) {
	_, data, err := a.schema.GetLatest(ctx, in.NamespaceId, in.SchemaId)
	return &stencilv1beta1.GetLatestSchemaResponse{
		Data: data,
	}, err
}

func (a *API) HTTPLatestSchema(w http.ResponseWriter, req *http.Request, pathParams map[string]string) (*schema.Metadata, []byte, error) {
	endFunc := a.newrelic.StartGenericSegment(req.Context(), "GetLatestSchema")
	defer endFunc()
	namespaceID := pathParams["namespace"]
	schemaName := pathParams["name"]
	metadata, data, err := a.schema.GetLatest(req.Context(), namespaceID, schemaName)
	return metadata, data, err
}

func (a *API) GetSchema(ctx context.Context, in *stencilv1beta1.GetSchemaRequest) (*stencilv1beta1.GetSchemaResponse, error) {
	_, data, err := a.schema.Get(ctx, in.NamespaceId, in.SchemaId, in.GetVersionId())
	return &stencilv1beta1.GetSchemaResponse{
		Data: data,
	}, err
}

func (a *API) HTTPGetSchema(w http.ResponseWriter, req *http.Request, pathParams map[string]string) (*schema.Metadata, []byte, error) {
	endFunc := a.newrelic.StartGenericSegment(req.Context(), "GetSchema")
	defer endFunc()
	namespaceID := pathParams["namespace"]
	schemaName := pathParams["name"]
	versionString := pathParams["version"]
	v, err := strconv.ParseInt(versionString, 10, 32)
	if err != nil {
		return nil, nil, &runtime.HTTPStatusError{HTTPStatus: http.StatusBadRequest, Err: errors.New("invalid version number")}
	}
	return a.schema.Get(req.Context(), namespaceID, schemaName, int32(v))
}

func (a *API) ListVersions(ctx context.Context, in *stencilv1beta1.ListVersionsRequest) (*stencilv1beta1.ListVersionsResponse, error) {
	versions, err := a.schema.ListVersions(ctx, in.NamespaceId, in.SchemaId)
	return &stencilv1beta1.ListVersionsResponse{Versions: versions}, err
}

func (a *API) GetSchemaMetadata(ctx context.Context, in *stencilv1beta1.GetSchemaMetadataRequest) (*stencilv1beta1.GetSchemaMetadataResponse, error) {
	meta, err := a.schema.GetMetadata(ctx, in.NamespaceId, in.SchemaId)
	return &stencilv1beta1.GetSchemaMetadataResponse{
		Format:        stencilv1beta1.Schema_Format(stencilv1beta1.Schema_Format_value[meta.Format]),
		Compatibility: stencilv1beta1.Schema_Compatibility(stencilv1beta1.Schema_Compatibility_value[meta.Compatibility]),
		Authority:     meta.Authority,
	}, err
}

func (a *API) UpdateSchemaMetadata(ctx context.Context, in *stencilv1beta1.UpdateSchemaMetadataRequest) (*stencilv1beta1.UpdateSchemaMetadataResponse, error) {
	meta, err := a.schema.UpdateMetadata(ctx, in.NamespaceId, in.SchemaId, &schema.Metadata{
		Compatibility: in.Compatibility.String(),
	})
	return &stencilv1beta1.UpdateSchemaMetadataResponse{
		Format:        stencilv1beta1.Schema_Format(stencilv1beta1.Schema_Format_value[meta.Format]),
		Compatibility: stencilv1beta1.Schema_Compatibility(stencilv1beta1.Schema_Compatibility_value[meta.Compatibility]),
		Authority:     meta.Authority,
	}, err
}

func (a *API) DeleteSchema(ctx context.Context, in *stencilv1beta1.DeleteSchemaRequest) (*stencilv1beta1.DeleteSchemaResponse, error) {
	err := a.schema.Delete(ctx, in.NamespaceId, in.SchemaId)
	message := "success"
	if err != nil {
		message = "failed"
	}
	return &stencilv1beta1.DeleteSchemaResponse{
		Message: message,
	}, err
}

func (a *API) DeleteVersion(ctx context.Context, in *stencilv1beta1.DeleteVersionRequest) (*stencilv1beta1.DeleteVersionResponse, error) {
	err := a.schema.DeleteVersion(ctx, in.NamespaceId, in.SchemaId, in.GetVersionId())
	message := "success"
	if err != nil {
		message = "failed"
	}
	return &stencilv1beta1.DeleteVersionResponse{
		Message: message,
	}, err
}

func (a *API) DetectSchemaChange(writer http.ResponseWriter, request *http.Request, pathParams map[string]string) error {
	namespaceID := pathParams["namespaceId"]
	schemaID := pathParams["schemaId"]

	fromVersion := request.URL.Query().Get("from")
	toVersion := request.URL.Query().Get("to")

	versionList, err := a.schema.ListVersions(context.Background(), namespaceID, schemaID)
	if err != nil {
		return fmt.Errorf("error getting version list - %s", err.Error())
	}
	if len(versionList) == 0 {
		return fmt.Errorf("got empty version list")
	}
	if len(versionList) == 1 {
		return fmt.Errorf("only one version exists for schema %s", schemaID)
	}
	latestVersion := versionList[len(versionList)-1]

	if fromVersion == "" {
		fromVersion = strconv.Itoa(int(latestVersion - 1))
	}
	if toVersion == "" {
		toVersion = strconv.Itoa(int(latestVersion))
	}

	fromVer, errFrom := strconv.ParseInt(fromVersion, 10, 32)
	toVer, errTo := strconv.ParseInt(toVersion, 10, 32)

	if errFrom != nil || errTo != nil {
		return fmt.Errorf("invalid version format")
	}
	if fromVer >= toVer {
		return fmt.Errorf("'from' should be less than 'to'")
	}

	ctx := context.Background()
	_, fromVerData, fromVerDataError := a.schema.Get(ctx, namespaceID, schemaID, int32(fromVer))
	if fromVerDataError != nil {
		return fmt.Errorf("error getting data for version %v - %s", fromVer, fromVerDataError.Error())
	}

	_, toVerData, toVerDataError := a.schema.Get(ctx, namespaceID, schemaID, int32(toVer))
	if toVerDataError != nil {
		return fmt.Errorf("error getting data for version %v - %s", toVer, toVerDataError.Error())
	}

	req := &changedetector.ChangeRequest{
		NamespaceID: namespaceID,
		SchemaName:  schemaID,
		OldData:     fromVerData,
		NewData:     toVerData,
		Version:     int32(toVer),
		Depth:       -1,
	}

	sce, err := a.changeDetector.IdentifySchemaChange(ctx, req)
	if err != nil {
		return fmt.Errorf("got error while identifying schema change for namespace : %s, schema: %s, version: %d, %s", req.NamespaceID, req.SchemaName, req.Version, err.Error())
	}
	log.Printf("schema change result %s", sce.String())
	if err := a.schema.SendNotification(ctx, sce, req); err != nil {
		return err
	}
	writer.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(writer).Encode(sce)
}
