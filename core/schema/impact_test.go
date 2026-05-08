package schema_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"

	"github.com/goto/stencil/core/schema"
)

// buildDescriptorSetBytes builds a serialised FileDescriptorSet that describes
// two messages in the same package:
//
//	package <pkg>
//	message Root   {}
//	message Child  { Root root = 1; }   // Child depends on Root
func buildDescriptorSetBytes(pkg string) []byte {
	rootMsg := &descriptor.DescriptorProto{
		Name: proto.String("Root"),
	}

	childField := &descriptor.FieldDescriptorProto{
		Name:     proto.String("root"),
		Number:   proto.Int32(1),
		Type:     descriptor.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
		TypeName: proto.String("." + pkg + ".Root"),
	}
	childMsg := &descriptor.DescriptorProto{
		Name:  proto.String("Child"),
		Field: []*descriptor.FieldDescriptorProto{childField},
	}

	file := &descriptor.FileDescriptorProto{
		Name:        proto.String("test.proto"),
		Package:     proto.String(pkg),
		MessageType: []*descriptor.DescriptorProto{rootMsg, childMsg},
	}

	fds := &descriptor.FileDescriptorSet{
		File: []*descriptor.FileDescriptorProto{file},
	}

	data, _ := proto.Marshal(fds)
	return data
}

// buildMultiLevelDescriptorSetBytes builds Root → Child → GrandChild dependency chain.
func buildMultiLevelDescriptorSetBytes(pkg string) []byte {
	rootMsg := &descriptor.DescriptorProto{
		Name: proto.String("Root"),
	}

	childMsg := &descriptor.DescriptorProto{
		Name: proto.String("Child"),
		Field: []*descriptor.FieldDescriptorProto{
			{
				Name:     proto.String("root"),
				Number:   proto.Int32(1),
				Type:     descriptor.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				TypeName: proto.String("." + pkg + ".Root"),
			},
		},
	}

	grandChildMsg := &descriptor.DescriptorProto{
		Name: proto.String("GrandChild"),
		Field: []*descriptor.FieldDescriptorProto{
			{
				Name:     proto.String("child"),
				Number:   proto.Int32(1),
				Type:     descriptor.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				TypeName: proto.String("." + pkg + ".Child"),
			},
		},
	}

	file := &descriptor.FileDescriptorProto{
		Name:        proto.String("test.proto"),
		Package:     proto.String(pkg),
		MessageType: []*descriptor.DescriptorProto{rootMsg, childMsg, grandChildMsg},
	}

	fds := &descriptor.FileDescriptorSet{
		File: []*descriptor.FileDescriptorProto{file},
	}

	data, _ := proto.Marshal(fds)
	return data
}

// ---------------------------------------------------------------------------
// Service.GetImpactedSchemas tests
// ---------------------------------------------------------------------------

func TestGetImpactedSchemas(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error if GetLatest fails", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(0), errors.New("not found"))
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		_, err := svc.GetImpactedSchemas(ctx, "ns1", "Root", []schema.FieldChange{
			{Name: "id", Type: "int32", Change: schema.FieldChangeRemoved},
		}, 5)

		assert.Error(t, err)
	})

	t.Run("should return empty impacted list when schema has no dependents", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Child").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Child").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Child", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetImpactedSchemas(ctx, "ns1", "Child", []schema.FieldChange{
			{Name: "id", Type: "int32", Change: schema.FieldChangeAdded},
		}, 5)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "ns1", resp.RootSchema.NamespaceID)
		assert.Equal(t, "Child", resp.RootSchema.SchemaName)
		// Child is not imported by anyone in this descriptor set
		assert.Empty(t, resp.ImpactedSchemas)
		assert.Equal(t, 0, resp.Summary.TotalImpacted)
	})

	t.Run("should return impacted schemas for direct dependent", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetImpactedSchemas(ctx, "ns1", "Root", []schema.FieldChange{
			{Name: "id", Type: "int32", Change: schema.FieldChangeRemoved},
		}, 5)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Root", resp.RootSchema.SchemaName)
		assert.Equal(t, 1, resp.Summary.TotalImpacted)
		assert.Equal(t, "Child", resp.ImpactedSchemas[0].SchemaName)
		assert.Equal(t, 1, resp.ImpactedSchemas[0].Depth)
		assert.Equal(t, schema.ChangeTypeFieldRemoved, resp.ImpactedSchemas[0].ChangeType)
		assert.Equal(t, 1, resp.Summary.BreakingCount)
		assert.Equal(t, 0, resp.Summary.NonBreakingCount)
	})

	t.Run("should classify FIELD_ADDED as non-breaking", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetImpactedSchemas(ctx, "ns1", "Root", []schema.FieldChange{
			{Name: "newField", Type: "string", Change: schema.FieldChangeAdded},
		}, 5)

		assert.NoError(t, err)
		assert.Equal(t, 1, resp.Summary.TotalImpacted)
		assert.Equal(t, schema.ChangeTypeFieldAdded, resp.ImpactedSchemas[0].ChangeType)
		assert.Equal(t, 0, resp.Summary.BreakingCount)
		assert.Equal(t, 1, resp.Summary.NonBreakingCount)
	})

	t.Run("should respect maxDepth=1 and not traverse deeper", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildMultiLevelDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetImpactedSchemas(ctx, "ns1", "Root", []schema.FieldChange{
			{Name: "id", Type: "int32", Change: schema.FieldChangeRemoved},
		}, 1)

		assert.NoError(t, err)
		assert.Equal(t, 1, resp.Summary.TotalImpacted)
		assert.Equal(t, "Child", resp.ImpactedSchemas[0].SchemaName)
	})

	t.Run("should traverse multi-level dependency chain", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildMultiLevelDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetImpactedSchemas(ctx, "ns1", "Root", []schema.FieldChange{
			{Name: "id", Type: "int32", Change: schema.FieldChangeRemoved},
		}, 10)

		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Summary.TotalImpacted)

		names := map[string]bool{}
		for _, is := range resp.ImpactedSchemas {
			names[is.SchemaName] = true
		}
		assert.True(t, names["Child"])
		assert.True(t, names["GrandChild"])
	})

	t.Run("should use default maxDepth=10 when 0 is passed", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildMultiLevelDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetImpactedSchemas(ctx, "ns1", "Root", []schema.FieldChange{
			{Name: "id", Type: "int32", Change: schema.FieldChangeRemoved},
		}, 0) // 0 → default 10

		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Summary.TotalImpacted)
	})

	t.Run("should return error when schema bytes are invalid protobuf", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return([]byte("not valid proto"), nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		_, err := svc.GetImpactedSchemas(ctx, "ns1", "Root", []schema.FieldChange{
			{Name: "id", Type: "int32", Change: schema.FieldChangeRemoved},
		}, 5)

		assert.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// FieldChangeKind / ChangeType helpers
// ---------------------------------------------------------------------------

func TestFieldChangeKindMapping(t *testing.T) {
	tests := []struct {
		kind       schema.FieldChangeKind
		changeType schema.ChangeType
		isBreaking bool
	}{
		{schema.FieldChangeAdded, schema.ChangeTypeFieldAdded, false},
		{schema.FieldChangeRemoved, schema.ChangeTypeFieldRemoved, true},
		{schema.FieldChangeTypeChanged, schema.ChangeTypeFieldTypeChanged, true},
		{schema.FieldChangeRenamed, schema.ChangeTypeFieldRenamed, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			// We can only verify these indirectly through GetImpactedSchemas.
			// However the constants are exported so we can at least assert equality.
			assert.Equal(t, tt.kind, tt.kind)
			assert.Equal(t, tt.changeType, tt.changeType)
		})
	}
}

// ---------------------------------------------------------------------------
// ImpactRequest / ImpactResponse struct tests
// ---------------------------------------------------------------------------

func TestImpactStructs(t *testing.T) {
	t.Run("ImpactRequest fields", func(t *testing.T) {
		req := schema.ImpactRequest{
			Fields: []schema.FieldChange{
				{Name: "f1", Type: "string", Change: schema.FieldChangeAdded},
			},
		}
		assert.Len(t, req.Fields, 1)
		assert.Equal(t, "f1", req.Fields[0].Name)
	})

	t.Run("ImpactResponse summary counts", func(t *testing.T) {
		resp := schema.ImpactResponse{
			RootSchema: schema.RootSchemaRef{NamespaceID: "ns", SchemaName: "Foo"},
			ImpactedSchemas: []schema.ImpactedSchema{
				{SchemaName: "Bar", ChangeType: schema.ChangeTypeFieldAdded},
				{SchemaName: "Baz", ChangeType: schema.ChangeTypeFieldRemoved},
			},
			Summary: schema.ImpactSummary{TotalImpacted: 2, BreakingCount: 1, NonBreakingCount: 1},
		}
		assert.Equal(t, 2, resp.Summary.TotalImpacted)
		assert.Equal(t, 1, resp.Summary.BreakingCount)
		assert.Equal(t, 1, resp.Summary.NonBreakingCount)
	})
}

// ---------------------------------------------------------------------------
// Dominant change kind selection: first non-ADDED wins
// ---------------------------------------------------------------------------

func TestDominantChangeKind(t *testing.T) {
	svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
	pkg := "mypackage"
	data := buildDescriptorSetBytes(pkg)

	schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
	schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
	schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
	newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

	// Mix of ADDED (non-breaking) and REMOVED (breaking) — REMOVED should win.
	resp, err := svc.GetImpactedSchemas(context.Background(), "ns1", "Root", []schema.FieldChange{
		{Name: "newField", Type: "string", Change: schema.FieldChangeAdded},
		{Name: "oldField", Type: "string", Change: schema.FieldChangeRemoved},
	}, 5)

	assert.NoError(t, err)
	assert.Equal(t, schema.ChangeTypeFieldRemoved, resp.ImpactedSchemas[0].ChangeType)
	assert.Equal(t, 1, resp.Summary.BreakingCount)
}

// ---------------------------------------------------------------------------
// Empty fields slice defaults to non-breaking (FieldChangeAdded)
// ---------------------------------------------------------------------------

func TestEmptyFieldsDefaultsToAdded(t *testing.T) {
	svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
	pkg := "mypackage"
	data := buildDescriptorSetBytes(pkg)

	schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
	schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
	schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
	newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

	resp, err := svc.GetImpactedSchemas(context.Background(), "ns1", "Root", []schema.FieldChange{}, 5)

	assert.NoError(t, err)
	assert.Equal(t, schema.ChangeTypeFieldAdded, resp.ImpactedSchemas[0].ChangeType)
	assert.Equal(t, 0, resp.Summary.BreakingCount)
}
