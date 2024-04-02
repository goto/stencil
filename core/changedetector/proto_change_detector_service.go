package changedetector

import (
	"context"
	"errors"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/google/uuid"
	"github.com/goto/stencil/pkg/newrelic"
	stencilv1beta1 "github.com/goto/stencil/proto/gotocompany/stencil/v1beta1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
)

func NewService(nr newrelic.Service) *Service {
	return &Service{
		newrelic: nr,
	}
}

type Service struct {
	newrelic newrelic.Service
}

func (s *Service) IdentifySchemaChange(ctx context.Context, request *ChangeRequest) (*stencilv1beta1.SchemaChangedEvent, error) {
	endFunc := s.newrelic.StartGenericSegment(ctx, "Identify Schema Change")
	defer endFunc()
	if request.OldData == nil || len(request.OldData) == 0 {
		log.Println("IdentifySchemaChange failed as previous schema data is nil")
		return nil, errors.New("previous schema data is nil")
	}
	currentFds, err := GetDescriptorSet(request.NewData)
	if err != nil {
		log.Printf("unable getSchemaChangeEvent get file descriptor set from current schema data %v\n", err)
		return nil, errors.New("unable getSchemaChangeEvent get file descriptor set from current schema data")
	}
	prevFds, err := GetDescriptorSet(request.OldData)
	if err != nil {
		log.Printf("unable getSchemaChangeEvent get file descriptor set from previous schema data %v\n", err)
		return nil, errors.New("unable getSchemaChangeEvent get file descriptor set from previous schema data")
	}
	sce := &stencilv1beta1.SchemaChangedEvent{
		EventId:        uuid.New().String(),
		EventTimestamp: timestamppb.New(time.Now()),
		NamespaceName:  request.NamespaceName,
		SchemaName:     request.SchemaName,
		Version:        request.Version,
	}
	setDirectlyImpactedSchemasAndFields(currentFds, prevFds, sce)
	reverseDependencies := getReverseDependenciesGraph(currentFds)
	for _, schema := range sce.UpdatedSchemas {
		appendImpactedDependents(sce, schema, findDependentImpactedSchemas(reverseDependencies, schema))
	}
	return sce, nil
}

func appendImpactedDependents(sce *stencilv1beta1.SchemaChangedEvent, key string, impactedDependents []string) {
	if sce.ImpactedSchemas == nil {
		sce.ImpactedSchemas = make(map[string]*stencilv1beta1.ImpactedSchemas)
	}
	sce.ImpactedSchemas[key] = &stencilv1beta1.ImpactedSchemas{
		SchemaNames: impactedDependents,
	}
}

func setDirectlyImpactedSchemasAndFields(currentFds, prevFds *descriptor.FileDescriptorSet, sce *stencilv1beta1.SchemaChangedEvent) {
	packageMessageMap := getPackageMessageMap(prevFds)
	packageEnumMap := getPackageEnumMap(prevFds)
	for _, fd := range currentFds.GetFile() {
		for _, newMessageDesc := range fd.GetMessageType() {
			messageName := fd.GetPackage() + "." + newMessageDesc.GetName()
			oldMessageDesc := getMessageDescriptor(packageMessageMap, fd.GetPackage(), newMessageDesc.GetName())
			if oldMessageDesc == nil {
				sce.UpdatedSchemas = append(sce.UpdatedSchemas, messageName)
				appendImpactedFields(sce, messageName, GetImpactedMessageFields(oldMessageDesc, newMessageDesc))
				continue
			}
			if !proto.Equal(newMessageDesc, oldMessageDesc) {
				sce.UpdatedSchemas = append(sce.UpdatedSchemas, messageName)
				appendImpactedFields(sce, messageName, GetImpactedMessageFields(oldMessageDesc, newMessageDesc))
			}
			compareEnumDescInMessageDesc(newMessageDesc, oldMessageDesc, messageName, sce)
		}
		compareEnumDescriptors(fd, packageEnumMap, sce)
	}
}

func compareEnumDescInMessageDesc(newMessageDesc, oldMessageDesc *descriptorpb.DescriptorProto, messageName string, sce *stencilv1beta1.SchemaChangedEvent) {
	for _, newEnumDesc := range newMessageDesc.GetEnumType() {
		enumName := messageName + "." + newEnumDesc.GetName()
		oldEnumDesc := findEnumDescriptorFromMessageDescriptor(oldMessageDesc, newEnumDesc.GetName())
		if oldEnumDesc == nil {
			sce.UpdatedSchemas = append(sce.UpdatedSchemas, enumName)
			appendImpactedFields(sce, enumName, GetImpactedEnumFields(oldEnumDesc, newEnumDesc))
			continue
		}
		if !proto.Equal(newEnumDesc, oldEnumDesc) {
			sce.UpdatedSchemas = append(sce.UpdatedSchemas, enumName)
			appendImpactedFields(sce, enumName, GetImpactedEnumFields(oldEnumDesc, newEnumDesc))
		}
	}
}
func appendImpactedFields(sce *stencilv1beta1.SchemaChangedEvent, key string, impactedFields []string) {
	if sce.UpdatedFields == nil {
		sce.UpdatedFields = make(map[string]*stencilv1beta1.ImpactedFields)
	}
	if val, ok := sce.UpdatedFields[key]; ok {
		val.FieldNames = append(val.FieldNames, impactedFields...)
		return
	}
	sce.UpdatedFields[key] = &stencilv1beta1.ImpactedFields{
		FieldNames: impactedFields,
	}
}

func compareEnumDescriptors(fds *descriptorpb.FileDescriptorProto, packageEnumMap map[string]map[string]*descriptor.EnumDescriptorProto, sce *stencilv1beta1.SchemaChangedEvent) {
	for _, newEnumDesc := range fds.GetEnumType() {
		enumName := fds.GetPackage() + "." + newEnumDesc.GetName()
		oldEnumDesc := getEnumDescriptor(packageEnumMap, fds.GetPackage(), newEnumDesc.GetName())
		if oldEnumDesc == nil {
			sce.UpdatedSchemas = append(sce.UpdatedSchemas, enumName)
			appendImpactedFields(sce, enumName, GetImpactedEnumFields(oldEnumDesc, newEnumDesc))
			continue
		}
		if !proto.Equal(newEnumDesc, oldEnumDesc) {
			sce.UpdatedSchemas = append(sce.UpdatedSchemas, enumName)
			appendImpactedFields(sce, enumName, GetImpactedEnumFields(oldEnumDesc, newEnumDesc))
		}
	}
}

/*
packageMessageMap is map having all the messages inside a package
[BookingLog][BookingLogMessage]=BookingLogMessage
*/
func getPackageMessageMap(fileDescriptorSet *descriptor.FileDescriptorSet) map[string]map[string]*descriptor.DescriptorProto {
	packageMessageMap := make(map[string]map[string]*descriptor.DescriptorProto)
	for _, fileDescriptor := range fileDescriptorSet.GetFile() {
		pkgName := fileDescriptor.GetPackage()
		if _, ok := packageMessageMap[pkgName]; !ok {
			packageMessageMap[pkgName] = make(map[string]*descriptor.DescriptorProto)
		}
		for _, messageDescriptor := range fileDescriptor.GetMessageType() {
			packageMessageMap[pkgName][messageDescriptor.GetName()] = messageDescriptor
		}
	}
	return packageMessageMap
}

/*
packageEnumMap is map having all the enums inside a package
[BookingLog][ServiceType]=ServiceType
*/
func getPackageEnumMap(fileDescriptorSet *descriptor.FileDescriptorSet) map[string]map[string]*descriptor.EnumDescriptorProto {
	packageEnumMap := make(map[string]map[string]*descriptor.EnumDescriptorProto)
	for _, fileDescriptor := range fileDescriptorSet.GetFile() {
		pkgName := fileDescriptor.GetPackage()
		if _, ok := packageEnumMap[pkgName]; !ok {
			packageEnumMap[pkgName] = make(map[string]*descriptor.EnumDescriptorProto)
		}
		for _, enumDescriptor := range fileDescriptor.GetEnumType() {
			packageEnumMap[pkgName][enumDescriptor.GetName()] = enumDescriptor
		}
	}
	return packageEnumMap
}

func getMessageDescriptor(packageMessageMap map[string]map[string]*descriptor.DescriptorProto, packageName, messageName string) *descriptor.DescriptorProto {
	if packageMap, found := packageMessageMap[packageName]; found {
		if descriptor, found := packageMap[messageName]; found {
			return descriptor
		}
	}
	return nil
}

func getEnumDescriptor(packageEnumMap map[string]map[string]*descriptor.EnumDescriptorProto, packageName, enumName string) *descriptor.EnumDescriptorProto {
	if packageMap, found := packageEnumMap[packageName]; found {
		if descriptor, found := packageMap[enumName]; found {
			return descriptor
		}
	}
	return nil
}

func findEnumDescriptorFromMessageDescriptor(messageDescriptor *descriptor.DescriptorProto, enumName string) *descriptor.EnumDescriptorProto {
	for _, enumDescriptor := range messageDescriptor.GetEnumType() {
		if enumDescriptor.GetName() == enumName {
			return enumDescriptor
		}
	}
	return nil
}

func getReverseDependenciesGraph(fileDescriptorSet *descriptor.FileDescriptorSet) map[string][]string {
	reverseDependencies := make(map[string][]string)
	for _, fileDescriptor := range fileDescriptorSet.GetFile() {
		for _, messageDescriptor := range fileDescriptor.GetMessageType() {
			messageName := fileDescriptor.GetPackage() + "." + messageDescriptor.GetName()
			for _, fieldDescriptor := range messageDescriptor.GetField() {
				fieldType := fieldDescriptor.GetTypeName()
				// Check if the field type is a message (nested message or imported message)
				if fieldType != "" && fieldType[0] == '.' {
					dependentMessage := fieldType[1:]
					reverseDependencies[dependentMessage] = append(reverseDependencies[dependentMessage], messageName)
				}
			}
		}
	}
	return reverseDependencies
}

func findDependentImpactedSchemas(reverseDependencies map[string][]string, impactedSchema string) []string {
	visitedMessages := make(map[string]bool)
	var dependentImpactedSchemas []string
	findDependents(impactedSchema, reverseDependencies, visitedMessages, &dependentImpactedSchemas)
	return dependentImpactedSchemas
}

func findDependents(messageName string, reverseDependencies map[string][]string, visitedMessages map[string]bool, impactedMessages *[]string) {
	if visitedMessages[messageName] {
		return
	}
	visitedMessages[messageName] = true
	*impactedMessages = append(*impactedMessages, messageName)
	for _, dependent := range reverseDependencies[messageName] {
		findDependents(dependent, reverseDependencies, visitedMessages, impactedMessages)
	}
}
