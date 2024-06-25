package schema_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/goto/stencil/config"
	"github.com/goto/stencil/core/changedetector"
	mocks2 "github.com/goto/stencil/pkg/newrelic/mocks"
	stencilv1beta2 "github.com/goto/stencil/proto/gotocompany/stencil/v1beta1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goto/stencil/core/namespace"
	"github.com/goto/stencil/core/schema"
	"github.com/goto/stencil/core/schema/mocks"
	"github.com/goto/stencil/internal/store"
)

func getSvc() (*schema.Service, *mocks.NamespaceService, *mocks.SchemaProvider, *mocks.SchemaRepository, *mocks2.NewRelic, *mocks.ChangeDetectorService, *mocks.Producer, *mocks.NotificationEventRepository) {
	nsService := &mocks.NamespaceService{}
	schemaProvider := &mocks.SchemaProvider{}
	schemaRepo := &mocks.SchemaRepository{}
	cache := &mocks.SchemaCache{}
	newRelic := &mocks2.NewRelic{}
	cdService := &mocks.ChangeDetectorService{}
	neRepo := &mocks.NotificationEventRepository{}
	cache.On("Get", mock.Anything).Return("", false)
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(false)
	producer := &mocks.Producer{}
	conf := &config.Config{
		SchemaChange: config.SchemaChangeConfig{
			Enable: true,
		},
	}
	svc := schema.NewService(schemaRepo, schemaProvider, nsService, cache, newRelic, cdService, producer, conf, neRepo)
	return svc, nsService, schemaProvider, schemaRepo, newRelic, cdService, producer, neRepo
}

func TestSchemaCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error if namespace not found", func(t *testing.T) {
		svc, nsService, _, _, newrelic, _, _, _ := getSvc()
		nsName := "testNamespace"
		var called bool
		newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
		nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{}, store.NoRowsErr)
		_, err := svc.Create(ctx, nsName, "a", &schema.Metadata{}, []byte(""))
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, store.NoRowsErr)
		nsService.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
	})

	t.Run("should return error if schema validation fails", func(t *testing.T) {
		svc, nsService, schemaProvider, _, newrelic, _, _, _ := getSvc()
		nsName := "testNamespace"
		data := []byte("data")
		nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{Format: "avro"}, nil)
		var called bool
		newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
		schemaProvider.On("ParseSchema", "protobuf", data).Return(&mocks.ParsedSchema{}, errors.New("invalid schema"))
		_, err := svc.Create(ctx, nsName, "a", &schema.Metadata{Format: "protobuf"}, data)
		assert.NotNil(t, err)
		schemaProvider.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
	})

	t.Run("should get format from namespace if format at schema level not defined", func(t *testing.T) {
		svc, nsService, schemaProvider, _, newrelic, _, _, _ := getSvc()
		nsName := "testNamespace"
		data := []byte("data")
		nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{Format: "protobuf"}, nil)
		var called bool
		newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
		schemaProvider.On("ParseSchema", "protobuf", data).Return(&mocks.ParsedSchema{}, errors.New("invalid schema"))
		_, err := svc.Create(ctx, nsName, "a", &schema.Metadata{}, data)
		assert.NotNil(t, err)
		schemaProvider.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
	})

	t.Run("should skip compatibility check if previous latest schema not present", func(t *testing.T) {
		svc, nsService, schemaProvider, schemaRepo, newrelic, _, _, _ := getSvc()
		scFile := &schema.SchemaFile{}
		parsedSchema := &mocks.ParsedSchema{}
		nsName := "testNamespace"
		data := []byte("data")
		nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{Format: "protobuf"}, nil)
		schemaProvider.On("ParseSchema", "protobuf", data).Return(parsedSchema, nil)
		schemaRepo.On("GetLatestVersion", mock.Anything, nsName, "a").Return(int32(2), store.NoRowsErr)
		parsedSchema.On("GetCanonicalValue").Return(scFile)
		schemaRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), nil)
		var called bool
		var compatibility bool
		newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
		newrelic.On("StartGenericSegment", mock.Anything, "Compatibility checker").Return(func() { compatibility = true })
		scInfo, err := svc.Create(ctx, nsName, "a", &schema.Metadata{}, data)
		assert.NoError(t, err)
		assert.Equal(t, scInfo.Version, int32(1))
		schemaRepo.AssertExpectations(t)
		nsService.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
		assert.True(t, compatibility)
	})

	t.Run("should identify schema change event and push to kafka and db", func(t *testing.T) {
		svc, nsService, schemaProvider, schemaRepo, newrelic, cdService, producer, neRepo := getSvc()
		scFile := &schema.SchemaFile{}
		parsedSchema := &mocks.ParsedSchema{}
		nsName := "testNamespace"
		data := []byte("data")
		nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{Format: "protobuf"}, nil)
		schemaProvider.On("ParseSchema", "protobuf", data).Return(parsedSchema, nil)
		schemaRepo.On("GetLatestVersion", mock.Anything, nsName, "a").Return(int32(3), nil)
		schemaRepo.On("Get", mock.Anything, nsName, "a", int32(3)).Return(data, nil)
		schemaRepo.On("GetMetadata", mock.Anything, nsName, "a").Return(&schema.Metadata{Format: "protobuf"}, nil)
		parsedSchema.On("GetCanonicalValue").Return(scFile)
		schemaRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), nil)
		schemaRepo.On("GetSchemaID", mock.Anything, nsName, "a").Return(int32(1), nil)
		sce := &stencilv1beta2.SchemaChangedEvent{
			UpdatedSchemas: []string{"a,b"},
		}
		cdService.On("IdentifySchemaChange", mock.Anything, mock.Anything).Return(sce, nil)
		producer.On("Write", mock.Anything, mock.Anything).Return(nil)
		neRepo.On("GetByNameSpaceSchemaVersionAndSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(changedetector.NotificationEvent{}, pgx.ErrNoRows)
		neRepo.On("Create", mock.Anything, mock.Anything).Return(changedetector.NotificationEvent{}, nil)
		neRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(changedetector.NotificationEvent{}, nil)
		var called bool
		var compatibility bool
		var cdCalled bool
		var metadata bool
		var dataCheck bool
		newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
		newrelic.On("StartGenericSegment", mock.Anything, "Compatibility checker").Return(func() { compatibility = true })
		newrelic.On("StartGenericSegment", mock.Anything, "Identify Schema Change").Return(func() { cdCalled = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetData").Return(func() { dataCheck = true })
		scInfo, err := svc.Create(ctx, nsName, "a", &schema.Metadata{}, data)
		time.Sleep(100 * time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, scInfo.Version, int32(1))
		schemaRepo.AssertExpectations(t)
		nsService.AssertExpectations(t)
		cdService.AssertExpectations(t)
		producer.AssertExpectations(t)
		neRepo.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
		assert.True(t, compatibility)
		assert.True(t, cdCalled)
		assert.True(t, metadata)
		assert.True(t, dataCheck)
	})

	t.Run("should identify schema change event and not push to kafka and db when updated schemas is zero", func(t *testing.T) {
		svc, nsService, schemaProvider, schemaRepo, newrelic, cdService, producer, neRepo := getSvc()

		scFile := &schema.SchemaFile{}
		parsedSchema := &mocks.ParsedSchema{}
		nsName := "testNamespace"
		data := []byte("data")
		nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{Format: "protobuf"}, nil)
		schemaProvider.On("ParseSchema", "protobuf", data).Return(parsedSchema, nil)
		schemaRepo.On("GetLatestVersion", mock.Anything, nsName, "a").Return(int32(3), nil)
		schemaRepo.On("Get", mock.Anything, nsName, "a", int32(3)).Return(data, nil)
		schemaRepo.On("GetMetadata", mock.Anything, nsName, "a").Return(&schema.Metadata{Format: "protobuf"}, nil)
		parsedSchema.On("GetCanonicalValue").Return(scFile)
		schemaRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), nil)
		schemaRepo.On("GetSchemaID", mock.Anything, nsName, "a").Return(int32(1), nil)
		cdService.On("IdentifySchemaChange", mock.Anything, mock.Anything).Return(&stencilv1beta2.SchemaChangedEvent{}, nil)
		neRepo.On("GetByNameSpaceSchemaVersionAndSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(changedetector.NotificationEvent{}, pgx.ErrNoRows)
		var called bool
		var compatibility bool
		var cdCalled bool
		var metadata bool
		var dataCheck bool
		newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
		newrelic.On("StartGenericSegment", mock.Anything, "Compatibility checker").Return(func() { compatibility = true })
		newrelic.On("StartGenericSegment", mock.Anything, "Identify Schema Change").Return(func() { cdCalled = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetData").Return(func() { dataCheck = true })
		scInfo, err := svc.Create(ctx, nsName, "a", &schema.Metadata{}, data)
		time.Sleep(100 * time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, scInfo.Version, int32(1))
		schemaRepo.AssertExpectations(t)
		nsService.AssertExpectations(t)
		cdService.AssertExpectations(t)
		producer.AssertExpectations(t)
		neRepo.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
		assert.True(t, compatibility)
		assert.True(t, cdCalled)
		assert.True(t, metadata)
		assert.True(t, dataCheck)
	})

	t.Run("should not trigger identify schema change if the feature flag is OFF", func(t *testing.T) {
		nsService := &mocks.NamespaceService{}
		schemaProvider := &mocks.SchemaProvider{}
		schemaRepo := &mocks.SchemaRepository{}
		cache := &mocks.SchemaCache{}
		newrelic := &mocks2.NewRelic{}
		cdService := &mocks.ChangeDetectorService{}
		cache.On("Get", mock.Anything).Return("", false)
		cache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(false)
		neRepo := &mocks.NotificationEventRepository{}
		producer := &mocks.Producer{}
		conf := &config.Config{
			SchemaChange: config.SchemaChangeConfig{
				Enable: false,
			},
		}
		svc := schema.NewService(schemaRepo, schemaProvider, nsService, cache, newrelic, cdService, producer, conf, neRepo)
		ctx := context.Background()
		scFile := &schema.SchemaFile{}
		parsedSchema := &mocks.ParsedSchema{}
		nsName := "testNamespace"
		data := []byte("data")
		nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{Format: "protobuf"}, nil)
		schemaProvider.On("ParseSchema", "protobuf", data).Return(parsedSchema, nil)
		schemaRepo.On("GetLatestVersion", mock.Anything, nsName, "a").Return(int32(3), nil)
		schemaRepo.On("Get", mock.Anything, nsName, "a", int32(3)).Return(data, nil)
		schemaRepo.On("GetMetadata", mock.Anything, nsName, "a").Return(&schema.Metadata{Format: "protobuf"}, nil)
		parsedSchema.On("GetCanonicalValue").Return(scFile)
		schemaRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), nil)

		var called, compatibility, metadata, dataCheck bool
		newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
		newrelic.On("StartGenericSegment", mock.Anything, "Compatibility checker").Return(func() { compatibility = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetData").Return(func() { dataCheck = true })
		scInfo, err := svc.Create(ctx, nsName, "a", &schema.Metadata{}, data)
		time.Sleep(100 * time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, scInfo.Version, int32(1))

		schemaRepo.AssertExpectations(t)
		nsService.AssertExpectations(t)
		cdService.AssertNotCalled(t, "IdentifySchemaChange")
		producer.AssertNotCalled(t, "Write")
		neRepo.AssertNotCalled(t, "GetByNameSpaceSchemaVersionAndSuccess")
		neRepo.AssertNotCalled(t, "Create")
		neRepo.AssertNotCalled(t, "Update")
		newrelic.AssertExpectations(t)
		assert.True(t, called)
		assert.True(t, compatibility)
		assert.True(t, metadata)
		assert.True(t, dataCheck)
	})

	t.Run("should return error if unable to get prev latest schema", func(t *testing.T) {
		svc, nsService, schemaProvider, schemaRepo, newrelic, _, _, _ := getSvc()
		parsedSchema := &mocks.ParsedSchema{}
		nsName := "testNamespace"
		data := []byte("data")
		nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{Format: "protobuf"}, nil)
		schemaProvider.On("ParseSchema", "protobuf", data).Return(parsedSchema, nil)
		schemaRepo.On("GetLatestVersion", mock.Anything, nsName, "a").Return(int32(2), errors.New("some other error apart from noRowsError"))
		var called bool
		var compatibility bool
		newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
		newrelic.On("StartGenericSegment", mock.Anything, "Compatibility checker").Return(func() { compatibility = true })
		_, err := svc.Create(ctx, nsName, "a", &schema.Metadata{}, data)
		assert.Error(t, err)
		schemaRepo.AssertExpectations(t)
		nsService.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
		assert.True(t, compatibility)
	})
	t.Run("should return error if previous latest schema is not valid", func(t *testing.T) {
		svc, nsService, schemaProvider, schemaRepo, newrelic, _, _, _ := getSvc()
		parsedSchema := &mocks.ParsedSchema{}
		prevParsedSchema := &mocks.ParsedSchema{}
		nsName := "testNamespace"
		data := []byte("data aa")
		prevData := []byte("some prev data")
		nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{Format: "protobuf", Compatibility: "COMPATIBILITY_BACKWARD"}, nil)
		schemaProvider.On("ParseSchema", "protobuf", data).Return(parsedSchema, nil).Once()
		schemaRepo.On("GetMetadata", mock.Anything, nsName, "a").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("GetLatestVersion", mock.Anything, nsName, "a").Return(int32(3), nil)
		schemaRepo.On("Get", mock.Anything, nsName, "a", int32(3)).Return(prevData, nil)
		schemaProvider.On("ParseSchema", "protobuf", prevData).Return(prevParsedSchema, errors.New("parse error")).Once()
		var called, compatibility, metadata, dataCheck bool
		newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
		newrelic.On("StartGenericSegment", mock.Anything, "Compatibility checker").Return(func() { compatibility = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetData").Return(func() { dataCheck = true })
		_, err := svc.Create(ctx, nsName, "a", &schema.Metadata{Compatibility: "COMPATIBILITY_FORWARD"}, data)
		assert.Error(t, err)
		schemaRepo.AssertExpectations(t)
		nsService.AssertExpectations(t)
		parsedSchema.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, called)
		assert.True(t, compatibility)
		assert.True(t, metadata)
		assert.True(t, dataCheck)
	})

	t.Run("should return error if compatibility check fails", func(t *testing.T) {
		for _, test := range []struct {
			compatibility string
			compFn        string
			isError       bool
		}{
			{"COMPATIBILITY_BACKWARD", "IsBackwardCompatible", true},
			{"COMPATIBILITY_FORWARD", "IsForwardCompatible", true},
			{"COMPATIBILITY_FULL", "IsFullCompatible", true},
		} {
			t.Run(test.compatibility, func(t *testing.T) {
				svc, nsService, schemaProvider, schemaRepo, newrelic, _, _, _ := getSvc()
				parsedSchema := &mocks.ParsedSchema{}
				prevParsedSchema := &mocks.ParsedSchema{}
				nsName := "testNamespace"
				data := []byte("data")
				prevData := []byte("some prev data")
				var compErr error
				if test.isError {
					compErr = errors.New("compatibilit error")
				}
				nsService.On("Get", mock.Anything, nsName).Return(namespace.Namespace{Format: "protobuf", Compatibility: "COMPATIBILITY_BACKWARD"}, nil)
				schemaProvider.On("ParseSchema", "protobuf", data).Return(parsedSchema, nil).Once()
				schemaRepo.On("GetMetadata", mock.Anything, nsName, "a").Return(&schema.Metadata{Format: "protobuf"}, nil)
				schemaRepo.On("GetLatestVersion", mock.Anything, nsName, "a").Return(int32(3), nil)
				schemaRepo.On("Get", mock.Anything, nsName, "a", int32(3)).Return(prevData, nil)
				schemaProvider.On("ParseSchema", "protobuf", prevData).Return(prevParsedSchema, nil).Once()
				parsedSchema.On(test.compFn, prevParsedSchema).Return(compErr)
				var called, compatibility, metadata, dataCheck bool
				newrelic.On("StartGenericSegment", mock.Anything, "Create Schema Info").Return(func() { called = true })
				newrelic.On("StartGenericSegment", mock.Anything, "Compatibility checker").Return(func() { compatibility = true })
				newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
				newrelic.On("StartGenericSegment", mock.Anything, "GetData").Return(func() { dataCheck = true })
				_, err := svc.Create(ctx, nsName, "a", &schema.Metadata{Compatibility: test.compatibility}, data)
				assert.Error(t, err)
				schemaRepo.AssertExpectations(t)
				nsService.AssertExpectations(t)
				parsedSchema.AssertExpectations(t)
				newrelic.AssertExpectations(t)
				assert.True(t, called)
				assert.True(t, compatibility)
				assert.True(t, metadata)
				assert.True(t, dataCheck)
			})
		}
	})
}

func TestGetSchema(t *testing.T) {
	ctx := context.Background()
	nsName := "testNamespace"
	schemaName := "testSchema"
	t.Run("should return error if get metadata fails", func(t *testing.T) {
		svc, _, _, repo, newrelic, _, _, _ := getSvc()
		var metadata bool
		newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
		repo.On("GetMetadata", mock.Anything, nsName, schemaName).Return(&schema.Metadata{}, errors.New("get metadata error"))
		_, _, err := svc.Get(ctx, nsName, schemaName, int32(1))
		assert.NotNil(t, err)
		newrelic.AssertExpectations(t)
		repo.AssertExpectations(t)
		assert.True(t, metadata)
	})

	t.Run("should return error if getting data fails", func(t *testing.T) {
		svc, _, _, repo, newrelic, _, _, _ := getSvc()
		version := int32(1)
		var metadata, dataCheck bool
		newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetData").Return(func() { dataCheck = true })
		repo.On("GetMetadata", mock.Anything, nsName, schemaName).Return(&schema.Metadata{}, nil)
		repo.On("Get", mock.Anything, nsName, schemaName, version).Return(nil, errors.New("get data error"))
		_, _, err := svc.Get(ctx, nsName, schemaName, version)
		assert.NotNil(t, err)
		repo.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, metadata)
		assert.True(t, dataCheck)
	})

	t.Run("should return metadata along with schema data", func(t *testing.T) {
		svc, _, _, repo, newrelic, _, _, _ := getSvc()
		version := int32(1)
		data := []byte("data")
		meta := &schema.Metadata{Format: "protobuf"}
		var metadata, dataCheck bool
		newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetData").Return(func() { dataCheck = true })
		repo.On("GetMetadata", mock.Anything, nsName, schemaName).Return(meta, nil)
		repo.On("Get", mock.Anything, nsName, schemaName, version).Return(data, nil)
		actualMeta, actualData, err := svc.Get(ctx, nsName, schemaName, version)
		assert.Nil(t, err)
		assert.Equal(t, data, actualData)
		assert.Equal(t, meta, actualMeta)
		repo.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, metadata)
		assert.True(t, dataCheck)
	})

	t.Run("should cache schema data", func(t *testing.T) {
		nsService := &mocks.NamespaceService{}
		schemaProvider := &mocks.SchemaProvider{}
		repo := &mocks.SchemaRepository{}
		cache := &mocks.SchemaCache{}
		newrelic := &mocks2.NewRelic{}

		svc := schema.NewService(repo, schemaProvider, nsService, cache, newrelic, nil, nil, nil, nil)
		var metadata, dataCheck bool
		newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetData").Return(func() { dataCheck = true })
		version := int32(1)
		data := []byte("data")
		meta := &schema.Metadata{Format: "protobuf"}
		key := "testNamespace-testSchema-1"
		cache.On("Get", key).Return("", false)
		cache.On("Set", key, data, int64(len(data))).Return(true)
		repo.On("GetMetadata", mock.Anything, nsName, schemaName).Return(meta, nil)
		repo.On("Get", mock.Anything, nsName, schemaName, version).Return(data, nil)
		actualMeta, actualData, err := svc.Get(ctx, nsName, schemaName, version)
		assert.Nil(t, err)
		assert.Equal(t, data, actualData)
		assert.Equal(t, meta, actualMeta)
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, metadata)
		assert.True(t, dataCheck)
	})

	t.Run("should get data from cache if key exists", func(t *testing.T) {
		nsService := &mocks.NamespaceService{}
		schemaProvider := &mocks.SchemaProvider{}
		repo := &mocks.SchemaRepository{}
		cache := &mocks.SchemaCache{}
		newrelic := &mocks2.NewRelic{}

		svc := schema.NewService(repo, schemaProvider, nsService, cache, newrelic, nil, nil, nil, nil)
		var metadata, dataCheck bool
		newrelic.On("StartGenericSegment", mock.Anything, "GetMetaData").Return(func() { metadata = true })
		newrelic.On("StartGenericSegment", mock.Anything, "GetData").Return(func() { dataCheck = true })
		version := int32(1)
		data := []byte("data")
		meta := &schema.Metadata{Format: "protobuf"}
		key := "testNamespace-testSchema-1"
		cache.On("Get", key).Return(data, true)
		repo.On("GetMetadata", mock.Anything, nsName, schemaName).Return(meta, nil)
		actualMeta, actualData, err := svc.Get(ctx, nsName, schemaName, version)
		assert.Nil(t, err)
		assert.Equal(t, data, actualData)
		assert.Equal(t, meta, actualMeta)
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
		newrelic.AssertExpectations(t)
		assert.True(t, metadata)
		assert.True(t, dataCheck)
	})
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name         string
		fromVersion  string
		toVersion    string
		schemaRepo   func(*mocks.SchemaRepository)
		expectedFrom int32
		expectedTo   int32
		expectedErr  error
	}{
		{
			name:         "Valid fromVersion and toVersion",
			fromVersion:  "2",
			toVersion:    "5",
			schemaRepo:   nil,
			expectedFrom: 2,
			expectedTo:   5,
			expectedErr:  nil,
		},
		{
			name:         "Invalid fromVersion format",
			fromVersion:  "abc",
			toVersion:    "5",
			schemaRepo:   nil,
			expectedFrom: 0,
			expectedTo:   0,
			expectedErr:  fmt.Errorf("invalid fromVersion format - strconv.ParseInt: parsing \"abc\": invalid syntax"),
		},
		{
			name:         "Invalid toVersion format",
			fromVersion:  "2",
			toVersion:    "xyz",
			schemaRepo:   nil,
			expectedFrom: 0,
			expectedTo:   0,
			expectedErr:  fmt.Errorf("invalid toVersion format - strconv.ParseInt: parsing \"xyz\": invalid syntax"),
		},
		{
			name:         "fromVersion less than 1",
			fromVersion:  "0",
			toVersion:    "5",
			schemaRepo:   nil,
			expectedFrom: 0,
			expectedTo:   0,
			expectedErr:  fmt.Errorf("fromVersion should be greater than or equal to 1"),
		},
		{
			name:         "toVersion less than or equal to 1",
			fromVersion:  "2",
			toVersion:    "1",
			schemaRepo:   nil,
			expectedFrom: 0,
			expectedTo:   0,
			expectedErr:  fmt.Errorf("toVersion should be greater than 1"),
		},
		{
			name:         "Only toVersion provided",
			fromVersion:  "",
			toVersion:    "3",
			schemaRepo:   nil,
			expectedFrom: 2,
			expectedTo:   3,
			expectedErr:  nil,
		},
		{
			name:        "Only fromVersion provided",
			fromVersion: "2",
			toVersion:   "",
			schemaRepo: func(mockRepo *mocks.SchemaRepository) {
				mockRepo.On("GetLatestVersion", mock.Anything, "testNamespace", "testSchema").Return(int32(4), nil)
			},
			expectedFrom: 2,
			expectedTo:   4,
			expectedErr:  nil,
		},
		{
			name:        "Both fromVersion and toVersion not provided",
			fromVersion: "",
			toVersion:   "",
			schemaRepo: func(mockRepo *mocks.SchemaRepository) {
				mockRepo.On("GetLatestVersion", mock.Anything, "testNamespace", "testSchema").Return(int32(3), nil)
			},
			expectedFrom: 2,
			expectedTo:   3,
			expectedErr:  nil,
		},
		{
			name:         "fromVersion greater than or equal to toVersion",
			fromVersion:  "5",
			toVersion:    "3",
			schemaRepo:   nil,
			expectedFrom: 0,
			expectedTo:   0,
			expectedErr:  fmt.Errorf("'from' should be less than 'to'"),
		},
		{
			name:        "Error in fetching latest version",
			fromVersion: "",
			toVersion:   "",
			schemaRepo: func(mockRepo *mocks.SchemaRepository) {
				mockRepo.On("GetLatestVersion", mock.Anything, "testNamespace", "testSchema").Return(int32(0), errors.New("some error"))
			},
			expectedFrom: 0,
			expectedTo:   0,
			expectedErr:  fmt.Errorf("error getting latest version - some error"),
		},
		{
			name:        "Only one version exists for the schema",
			fromVersion: "",
			toVersion:   "",
			schemaRepo: func(mockRepo *mocks.SchemaRepository) {
				mockRepo.On("GetLatestVersion", mock.Anything, "testNamespace", "testSchema").Return(int32(1), nil)
			},
			expectedFrom: 0,
			expectedTo:   0,
			expectedErr:  fmt.Errorf("only one version exists for schema - testSchema"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, _, _, schemaRepo, _, _, _, _ := getSvc()

			if tt.schemaRepo != nil {
				tt.schemaRepo(schemaRepo)
			}

			fromVer, toVer, err := svc.GetVersions(ctx, "testNamespace", "testSchema", tt.fromVersion, tt.toVersion)
			assert.Equal(t, tt.expectedFrom, fromVer)
			assert.Equal(t, tt.expectedTo, toVer)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
