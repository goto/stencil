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
// Service.GetLineage tests
// ---------------------------------------------------------------------------

func TestGetLineage(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error if GetLatest fails", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(0), errors.New("not found"))
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		_, err := svc.GetLineage(ctx, "ns1", "Root", 5, schema.LineageDirectionDownstream)

		assert.Error(t, err)
	})

	t.Run("should return empty downstream lineage when schema has no dependents", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Child").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Child").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Child", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetLineage(ctx, "ns1", "Child", 5, schema.LineageDirectionDownstream)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, schema.LineageDirectionDownstream, resp.Direction)
		assert.Equal(t, "ns1", resp.RootSchema.NamespaceID)
		assert.Equal(t, "Child", resp.RootSchema.SchemaName)
		assert.Empty(t, resp.Downstream)
		assert.Empty(t, resp.Upstream)
		assert.Equal(t, 0, resp.Summary.TotalCount)
	})

	t.Run("should return downstream lineage for direct dependent", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetLineage(ctx, "ns1", "Root", 5, schema.LineageDirectionDownstream)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Root", resp.RootSchema.SchemaName)
		assert.Equal(t, 1, resp.Summary.DownstreamCount)
		assert.Equal(t, 1, resp.Summary.TotalCount)
		assert.Equal(t, "Child", resp.Downstream[0].SchemaName)
		assert.Equal(t, 1, resp.Downstream[0].Level)
		assert.Equal(t, []string{"Root", "Child"}, resp.Downstream[0].Path)
	})

	t.Run("should return upstream lineage", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildMultiLevelDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "GrandChild").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "GrandChild").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "GrandChild", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetLineage(ctx, "ns1", "GrandChild", 10, schema.LineageDirectionUpstream)

		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Summary.UpstreamCount)
		assert.Equal(t, 2, resp.Summary.TotalCount)
		assert.Equal(t, "Child", resp.Upstream[0].SchemaName)
		assert.Equal(t, "Root", resp.Upstream[1].SchemaName)
		assert.Equal(t, []string{"GrandChild", "Child"}, resp.Upstream[0].Path)
	})

	t.Run("should default to both directions when direction is empty", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildMultiLevelDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Child").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Child").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Child", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetLineage(ctx, "ns1", "Child", 10, "")

		assert.NoError(t, err)
		assert.Equal(t, schema.LineageDirectionBoth, resp.Direction)
		assert.Equal(t, 1, resp.Summary.DownstreamCount)
		assert.Equal(t, 1, resp.Summary.UpstreamCount)
		assert.Equal(t, 2, resp.Summary.TotalCount)
		assert.Equal(t, "GrandChild", resp.Downstream[0].SchemaName)
		assert.Equal(t, "Root", resp.Upstream[0].SchemaName)
	})

	t.Run("should respect level=1 and not traverse deeper", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildMultiLevelDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetLineage(ctx, "ns1", "Root", 1, schema.LineageDirectionDownstream)

		assert.NoError(t, err)
		assert.Equal(t, 1, resp.Summary.DownstreamCount)
		assert.Equal(t, "Child", resp.Downstream[0].SchemaName)
	})

	t.Run("should traverse multi-level downstream lineage", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildMultiLevelDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetLineage(ctx, "ns1", "Root", 10, schema.LineageDirectionDownstream)

		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Summary.DownstreamCount)

		names := map[string]bool{}
		for _, is := range resp.Downstream {
			names[is.SchemaName] = true
		}
		assert.True(t, names["Child"])
		assert.True(t, names["GrandChild"])
	})

	t.Run("should use default level=10 when 0 is passed", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		pkg := "mypackage"
		data := buildMultiLevelDescriptorSetBytes(pkg)

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		resp, err := svc.GetLineage(ctx, "ns1", "Root", 0, schema.LineageDirectionDownstream) // 0 → default 10

		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Summary.DownstreamCount)
	})

	t.Run("should return error when schema bytes are invalid protobuf", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return([]byte("not valid proto"), nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		_, err := svc.GetLineage(ctx, "ns1", "Root", 5, schema.LineageDirectionDownstream)

		assert.Error(t, err)
	})

	t.Run("should return error on invalid direction", func(t *testing.T) {
		svc, _, _, schemaRepo, newrelic, _, _, _ := getSvc()
		data := buildDescriptorSetBytes("mypackage")

		schemaRepo.On("GetLatestVersion", mock.Anything, "ns1", "Root").Return(int32(1), nil)
		schemaRepo.On("GetMetadata", mock.Anything, "ns1", "Root").Return(&schema.Metadata{Format: "protobuf"}, nil)
		schemaRepo.On("Get", mock.Anything, "ns1", "Root", int32(1)).Return(data, nil)
		newrelic.On("StartGenericSegment", mock.Anything, mock.Anything).Return(func() {})

		_, err := svc.GetLineage(ctx, "ns1", "Root", 5, schema.LineageDirection("sideways"))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid direction")
	})
}

// ---------------------------------------------------------------------------
// LineageResponse struct tests
// ---------------------------------------------------------------------------

func TestLineageStructs(t *testing.T) {
	t.Run("LineageResponse summary counts", func(t *testing.T) {
		resp := schema.LineageResponse{
			RootSchema: schema.RootSchemaRef{NamespaceID: "ns", SchemaName: "Foo"},
			Downstream: []schema.LineageSchema{
				{SchemaName: "Bar"},
			},
			Upstream: []schema.LineageSchema{
				{SchemaName: "Baz"},
			},
			Summary: schema.LineageSummary{DownstreamCount: 1, UpstreamCount: 1, TotalCount: 2},
		}
		assert.Equal(t, 2, resp.Summary.TotalCount)
	})
}
