package changedetector

import (
	"container/list"
	"context"
	"errors"
	"log"
	"math"
	"time"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/goto/stencil/pkg/newrelic"
	stencilv1beta1 "github.com/goto/stencil/proto/gotocompany/stencil/v1beta1"
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
	prevFds, err := GetDescriptorSet(request.OldData)
	if err != nil {
		log.Printf("unable to get file descriptor set from previous schema data %s\n", err.Error())
		return nil, errors.New("unable to get file descriptor set from previous schema data")
	}
	currentFds, err := GetDescriptorSet(request.NewData)
	if err != nil {
		log.Printf("unable to get file descriptor set from current schema data %s\n", err.Error())
		return nil, errors.New("unable to get file descriptor set from current schema data")
	}
	sce := &stencilv1beta1.SchemaChangedEvent{
		EventId:        uuid.New().String(),
		EventTimestamp: timestamppb.New(time.Now()),
		NamespaceName:  request.NamespaceID,
		SchemaName:     request.SchemaName,
		Version:        request.Version,
		Metadata: &stencilv1beta1.Metadata{
			SourceUrl: request.SourceURL,
			CommitSha: request.CommitSHA,
		},
	}
	setDirectlyImpactedSchemasAndFields(currentFds, prevFds, sce)
	reverseDependencies := getReverseDependencies(currentFds)
	/*
		If depth is negative then show all impacted schemas
	*/
	if request.Depth < 0 {
		request.Depth = math.MaxInt32
	}
	for _, schema := range sce.UpdatedSchemas {
		appendImpactedDependents(sce, schema, getDependentImpactedSchemas(reverseDependencies, schema, request.Depth))
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
	prevMessageMap := collectAllMessages(prevFds)
	packageEnumMap := getPackageEnumMap(prevFds)
	for _, fileDesc := range currentFds.GetFile() {
		pkg := fileDesc.GetPackage()
		for _, newMessageDesc := range fileDesc.GetMessageType() {
			compareMessageTree(pkg, newMessageDesc, prevMessageMap, sce)
		}
		compareEnumDescriptors(fileDesc, packageEnumMap, sce)
	}
}

// collectAllMessages returns a flat map of fully-qualified name -> descriptor
// for ALL messages (top-level and nested) across all files in the descriptor set.
func collectAllMessages(fds *descriptor.FileDescriptorSet) map[string]*descriptor.DescriptorProto {
	result := make(map[string]*descriptor.DescriptorProto)
	for _, fileDesc := range fds.GetFile() {
		pkg := fileDesc.GetPackage()
		for _, msgDesc := range fileDesc.GetMessageType() {
			walkMessageTree(pkg, msgDesc, result)
		}
	}
	return result
}

// walkMessageTree recursively registers a message and all its nested messages
// into the result map under their fully-qualified names (e.g. "test.Outer.Inner").
func walkMessageTree(prefix string, msg *descriptor.DescriptorProto, result map[string]*descriptor.DescriptorProto) {
	fullName := prefix + "." + msg.GetName()
	result[fullName] = msg
	for _, nested := range msg.GetNestedType() {
		walkMessageTree(fullName, nested, result)
	}
}

// compareMessageTree recursively compares a message and all its nested messages,
// recording any changes into the SchemaChangedEvent.
func compareMessageTree(prefix string, newMsg *descriptor.DescriptorProto, prevMessageMap map[string]*descriptor.DescriptorProto, sce *stencilv1beta1.SchemaChangedEvent) {
	fullName := prefix + "." + newMsg.GetName()
	oldMsg := prevMessageMap[fullName]
	if oldMsg == nil {
		sce.UpdatedSchemas = append(sce.UpdatedSchemas, fullName)
		appendImpactedFields(sce, fullName, GetImpactedMessageFields(oldMsg, newMsg))
	} else if !proto.Equal(oldMsg, newMsg) {
		sce.UpdatedSchemas = append(sce.UpdatedSchemas, fullName)
		appendImpactedFields(sce, fullName, GetImpactedMessageFields(oldMsg, newMsg))
		compareEnumDescInMessageDesc(oldMsg, newMsg, fullName, sce)
	}
	// Always recurse into nested messages so each level is independently evaluated.
	for _, nested := range newMsg.GetNestedType() {
		compareMessageTree(fullName, nested, prevMessageMap, sce)
	}
}

func compareEnumDescInMessageDesc(oldMessageDesc, newMessageDesc *descriptorpb.DescriptorProto, messageName string, sce *stencilv1beta1.SchemaChangedEvent) {
	for _, newEnumDesc := range newMessageDesc.GetEnumType() {
		enumName := messageName + "." + newEnumDesc.GetName()
		oldEnumDesc := findEnumDescriptorFromMessageDescriptor(oldMessageDesc, newEnumDesc.GetName())
		if oldEnumDesc == nil {
			sce.UpdatedSchemas = append(sce.UpdatedSchemas, enumName)
			appendImpactedFields(sce, enumName, GetImpactedEnumFields(oldEnumDesc, newEnumDesc))
			continue
		}
		if !proto.Equal(oldEnumDesc, newEnumDesc) {
			sce.UpdatedSchemas = append(sce.UpdatedSchemas, enumName)
			appendImpactedFields(sce, enumName, GetImpactedEnumFields(oldEnumDesc, newEnumDesc))
		}
	}
}

// Key can be messageName or enumName
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
		if !proto.Equal(oldEnumDesc, newEnumDesc) {
			sce.UpdatedSchemas = append(sce.UpdatedSchemas, enumName)
			appendImpactedFields(sce, enumName, GetImpactedEnumFields(oldEnumDesc, newEnumDesc))
		}
	}
}

/*
packageEnumMap is map having all the enums inside a package
[com.goto.bookinglog][ServiceTypeEnum]=ServiceTypeEnumDescriptor
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

func getReverseDependencies(fileDescriptorSet *descriptor.FileDescriptorSet) map[string][]string {
	reverseDependencies := make(map[string][]string)
	for _, fileDescriptor := range fileDescriptorSet.GetFile() {
		pkg := fileDescriptor.GetPackage()
		for _, messageDescriptor := range fileDescriptor.GetMessageType() {
			buildReverseDepsFromMessage(pkg, messageDescriptor, reverseDependencies)
		}
	}
	return reverseDependencies
}

// buildReverseDepsFromMessage recursively registers reverse dependencies for a message
// and all its nested messages, so that e.g. "test.Outer.Inner" is a valid lookup key.
func buildReverseDepsFromMessage(prefix string, msg *descriptor.DescriptorProto, reverseDeps map[string][]string) {
	msgFullName := prefix + "." + msg.GetName()
	for _, fieldDescriptor := range msg.GetField() {
		fieldType := fieldDescriptor.GetTypeName()
		/*Check if the field type is a message (nested message or imported message)
		Ref:https://cloud.google.com/java/docs/reference/protobuf/latest/com.google.protobuf.DescriptorProtos.FieldDescriptorProto#com_google_protobuf_DescriptorProtos_FieldDescriptorProto_getType__:~:text=The%20type.-,getTypeName(),-public%20String%20getTypeName
		*/
		if fieldType != "" && fieldType[0] == '.' {
			dependentMessage := fieldType[1:]
			reverseDeps[dependentMessage] = append(reverseDeps[dependentMessage], msgFullName)
		}
	}
	for _, nested := range msg.GetNestedType() {
		buildReverseDepsFromMessage(msgFullName, nested, reverseDeps)
	}
}

func getDependentImpactedSchemas(reverseDependencies map[string][]string, impactedSchema string, depth int32) []string {
	visitedMessages := make(map[string]bool)
	var dependentImpactedSchemas []string
	queue := list.New()
	queue.PushBack(impactedSchema)
	visitedMessages[impactedSchema] = true
	for queue.Len() > 0 && depth >= 0 {
		size := queue.Len()
		for i := 0; i < size; i++ {
			currentMessage := queue.Front().Value.(string)
			queue.Remove(queue.Front())
			dependentImpactedSchemas = append(dependentImpactedSchemas, currentMessage)
			for _, neighbor := range reverseDependencies[currentMessage] {
				if !visitedMessages[neighbor] {
					queue.PushBack(neighbor)
					visitedMessages[neighbor] = true
				}
			}
		}
		depth--
	}
	return dependentImpactedSchemas
}
