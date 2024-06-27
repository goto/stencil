package schema

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/goto/stencil/core/changedetector"
	stencilv1beta1 "github.com/goto/stencil/proto/gotocompany/stencil/v1beta1"
)

type Metadata struct {
	Authority     string
	Format        string
	Compatibility string
	SourceURL     string
}

type SchemaInfo struct {
	ID       string `json:"id"`
	Version  int32  `json:"version"`
	Location string `json:"location"`
}

type SchemaFile struct {
	ID     string
	Types  []string
	Fields []string
	Data   []byte
}

type UpdateSchemaRequest struct {
	Namespace  string
	Schema     string
	Metadata   *Metadata
	SchemaFile *SchemaFile
}

type Repository interface {
	Create(ctx context.Context, request *UpdateSchemaRequest, versionID string, commitSHA string) (version int32, err error)
	List(context.Context, string) ([]Schema, error)
	ListVersions(context.Context, string, string) ([]int32, error)
	Get(context.Context, string, string, int32) ([]byte, error)
	GetLatestVersion(context.Context, string, string) (int32, error)
	GetMetadata(context.Context, string, string) (*Metadata, error)
	UpdateMetadata(context.Context, string, string, *Metadata) (*Metadata, error)
	Delete(context.Context, string, string) error
	DeleteVersion(context.Context, string, string, int32) error
	GetSchemaID(ctx context.Context, ns string, sc string) (int32, error)
}

type ParsedSchema interface {
	IsBackwardCompatible(ParsedSchema) error
	IsForwardCompatible(ParsedSchema) error
	IsFullCompatible(ParsedSchema) error
	Format() string
	GetCanonicalValue() *SchemaFile
}

type Provider interface {
	ParseSchema(format string, data []byte) (ParsedSchema, error)
}

type Cache interface {
	Get(interface{}) (interface{}, bool)
	Set(interface{}, interface{}, int64) bool
}

type Schema struct {
	Name          string
	Format        string
	Compatibility string
	Authority     string
	SourceURL     string
}

type ChangeDetectorService interface {
	IdentifySchemaChange(ctx context.Context, request *changedetector.ChangeRequest) (*stencilv1beta1.SchemaChangedEvent, error)
}

type Producer interface {
	Write(topic string, protoMessage proto.Message) error
}

type NotificationEventRepository interface {
	Create(ctx context.Context, event changedetector.NotificationEvent) (changedetector.NotificationEvent, error)
	Update(ctx context.Context, Id string, success bool) (changedetector.NotificationEvent, error)
	GetByNameSpaceSchemaVersionAndSuccess(ctx context.Context, namespace string, schemaID int32, versionID string, success bool) (changedetector.NotificationEvent, error)
}
