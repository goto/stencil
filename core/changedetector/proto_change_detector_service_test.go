package changedetector_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goto/stencil/core/changedetector"
	"github.com/goto/stencil/formats/protobuf"
	stencilv1beta2 "github.com/goto/stencil/proto/gotocompany/stencil/v1beta1"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
)

func getSvc() *changedetector.Service {
	svc := changedetector.NewService()
	return svc
}

func TestIdentifySchemaChange(t *testing.T) {
	t.Run("should return error if previous schema data is nil", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
			NewData:     []byte("b"),
		}
		_, err := svc.IdentifySchemaChange(request)
		assert.NotNil(t, err)
		assert.Equal(t, errors.New("previous schema data is nil"), err)
	})

	t.Run("should return error if unable to get file descriptor from current schema data", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
			OldData:     []byte("a"),
			NewData:     []byte("b"),
		}
		_, err := svc.IdentifySchemaChange(request)
		assert.NotNil(t, err)
		assert.Equal(t, errors.New("unable getSchemaChangeEvent get file descriptor set from current schema data"), err)
	})

	t.Run("should return error if unable to get file descriptor from previous schema data", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
			OldData:     []byte("a"),
		}
		request.NewData = getDescriptorData(t, "./testdata/input", true, []string{"schema_with_no_dependency_v1.proto"})
		_, err := svc.IdentifySchemaChange(request)
		assert.NotNil(t, err)
		assert.Equal(t, errors.New("unable getSchemaChangeEvent get file descriptor set from previous schema data"), err)
	})
	t.Run("should return schema change event when any new message added in latest schema", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
		}
		oldData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_no_dependency_v1.proto"})
		newData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_dependency_v1.proto"})
		request.OldData = oldData
		request.NewData = newData
		actual, err := svc.IdentifySchemaChange(request)
		expected := getSchemaChangeEvent("./testdata/output/sce_new_message_added.json")
		assert.Nil(t, err)
		assert.NotNil(t, actual)
		assertSchemaChangeEvent(t, expected, actual)
	})
	t.Run("should return schema change event when any new field added in latest schema with no dependent schema", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
		}
		oldData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_no_dependency_v1.proto"})
		newData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_no_dependency_v2.proto"})
		request.OldData = oldData
		request.NewData = newData
		actual, err := svc.IdentifySchemaChange(request)
		expected := getSchemaChangeEvent("./testdata/output/sce_fields_updated_with_no_dependencies.json")
		assert.Nil(t, err)
		assert.NotNil(t, actual)
		assertSchemaChangeEvent(t, expected, actual)
	})

	t.Run("should return schema change event when any new field added in latest schema with dependent schema", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
		}
		oldData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_dependency_v1.proto"})
		newData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_dependency_v2.proto"})
		request.OldData = oldData
		request.NewData = newData
		actual, err := svc.IdentifySchemaChange(request)
		expected := getSchemaChangeEvent("./testdata/output/sce_fields_updated_with_dependencies.json")
		assert.Nil(t, err)
		assert.NotNil(t, actual)
		assertSchemaChangeEvent(t, expected, actual)
	})
	t.Run("should return schema change event when enum field added inside message in latest schema", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
		}
		oldData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_dependency_v1.proto"})
		newData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_enum_inside_message_v2.proto"})
		request.OldData = oldData
		request.NewData = newData
		actual, err := svc.IdentifySchemaChange(request)
		expected := getSchemaChangeEvent("./testdata/output/sce_enum_fields_inside_message_added.json")
		assert.Nil(t, err)
		assert.NotNil(t, actual)
		assertSchemaChangeEvent(t, expected, actual)
	})

	t.Run("should return schema change event when enum field updated inside message in latest schema", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
		}
		oldData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_enum_inside_message_v1.proto"})
		newData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_enum_inside_message_v2.proto"})
		request.OldData = oldData
		request.NewData = newData
		actual, err := svc.IdentifySchemaChange(request)
		expected := getSchemaChangeEvent("./testdata/output/sce_enum_fields_inside_message_updated.json")
		assert.Nil(t, err)
		assert.NotNil(t, actual)
		assertSchemaChangeEvent(t, expected, actual)
	})

	t.Run("should return schema change event when enum field added in latest schema", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
		}
		oldData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_no_dependency_v1.proto"})
		newData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_enum_added.proto"})
		request.OldData = oldData
		request.NewData = newData
		actual, err := svc.IdentifySchemaChange(request)
		expected := getSchemaChangeEvent("./testdata/output/sce_enum_added.json")
		assert.Nil(t, err)
		assert.NotNil(t, actual)
		assertSchemaChangeEvent(t, expected, actual)
	})

	t.Run("should return schema change event when enum field updated in latest schema", func(t *testing.T) {
		svc := getSvc()
		request := &changedetector.ChangeRequest{
			NamespaceID: "testNamespace",
			SchemaName:  "testSchemaName",
			Version:     int32(1),
		}
		oldData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_enum_added.proto"})
		newData := getDescriptorData(t, "./testdata/input", true, []string{"schema_with_enum_updated.proto"})
		request.OldData = oldData
		request.NewData = newData
		actual, err := svc.IdentifySchemaChange(request)
		expected := getSchemaChangeEvent("./testdata/output/sce_enum_updated.json")
		assert.Nil(t, err)
		assert.NotNil(t, actual)
		assertSchemaChangeEvent(t, expected, actual)
	})

}

func assertSchemaChangeEvent(t *testing.T, expected, actual *stencilv1beta2.SchemaChangedEvent) {
	t.Helper()
	assert.Equal(t, expected.NamespaceName, actual.NamespaceName)
	assert.Equal(t, expected.SchemaName, actual.SchemaName)
	assert.Equal(t, expected.Version, actual.Version)
	assert.ElementsMatch(t, expected.UpdatedSchemas, actual.UpdatedSchemas)
	assert.Equal(t, expected.UpdatedFields, actual.UpdatedFields)
	assert.Equal(t, expected.ImpactedSchemas, actual.ImpactedSchemas)
}

func getSchemaChangeEvent(filePath string) *stencilv1beta2.SchemaChangedEvent {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return nil
	}
	var event *stencilv1beta2.SchemaChangedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil
	}
	return event
}
func getDescriptorData(t *testing.T, path string, includeImports bool, protoFiles []string) []byte {
	t.Helper()
	root, _ := filepath.Abs(path)
	log.Println(t.Name())
	targetFile := filepath.Join(t.TempDir(), protobuf.GetRandomName())
	err := protobuf.RunProtoc(root, includeImports, targetFile, protoFiles)
	assert.NoError(t, err)
	data, err := ioutil.ReadFile(targetFile)
	assert.NoError(t, err)
	return data
}
