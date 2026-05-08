package schema

import (
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	"github.com/goto/stencil/core/changedetector"
)

// FieldChangeKind enumerates the kinds of field-level changes a caller can report.
type FieldChangeKind string

const (
	FieldChangeAdded       FieldChangeKind = "ADDED"
	FieldChangeRemoved     FieldChangeKind = "REMOVED"
	FieldChangeTypeChanged FieldChangeKind = "TYPE_CHANGED"
	FieldChangeRenamed     FieldChangeKind = "RENAMED"
)

// ChangeType is the derived impact classification surfaced in the response.
type ChangeType string

const (
	ChangeTypeFieldAdded       ChangeType = "FIELD_ADDED"
	ChangeTypeFieldRemoved     ChangeType = "FIELD_REMOVED"
	ChangeTypeFieldTypeChanged ChangeType = "FIELD_TYPE_CHANGED"
	ChangeTypeFieldRenamed     ChangeType = "FIELD_RENAMED"
)

// FieldChange describes a single proposed field mutation on the root schema.
type FieldChange struct {
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	Change FieldChangeKind `json:"change"`
}

// ImpactRequest is the request body for the impact analysis endpoint.
type ImpactRequest struct {
	Fields []FieldChange `json:"fields"`
}

// RootSchemaRef identifies the schema on which the change is proposed.
type RootSchemaRef struct {
	NamespaceID string `json:"namespace_id"`
	SchemaName  string `json:"schema_name"`
}

// ImpactedSchema represents a single schema that is transitively affected.
type ImpactedSchema struct {
	NamespaceID string     `json:"namespace_id"`
	SchemaName  string     `json:"schema_name"`
	ChangeType  ChangeType `json:"change_type"`
	Depth       int        `json:"depth"`
	ImportPath  []string   `json:"import_path"`
}

// ImpactSummary aggregates the impact counts.
type ImpactSummary struct {
	TotalImpacted    int `json:"total_impacted"`
	BreakingCount    int `json:"breaking_count"`
	NonBreakingCount int `json:"non_breaking_count"`
}

// ImpactResponse is the full response body for the impact endpoint.
type ImpactResponse struct {
	RootSchema      RootSchemaRef    `json:"root_schema"`
	ImpactedSchemas []ImpactedSchema `json:"impacted_schemas"`
	Summary         ImpactSummary    `json:"summary"`
}

// fieldChangeToChangeType maps a FieldChangeKind to the derived ChangeType.
func fieldChangeToChangeType(k FieldChangeKind) ChangeType {
	switch k {
	case FieldChangeAdded:
		return ChangeTypeFieldAdded
	case FieldChangeRemoved:
		return ChangeTypeFieldRemoved
	case FieldChangeTypeChanged:
		return ChangeTypeFieldTypeChanged
	case FieldChangeRenamed:
		return ChangeTypeFieldRenamed
	default:
		return ChangeTypeFieldRemoved
	}
}

// isBreaking returns true when the change type is considered breaking.
func isBreaking(ct ChangeType) bool {
	return ct != ChangeTypeFieldAdded
}

// buildReverseDepMap constructs a map of fullyQualifiedName → []dependents
// from the given FileDescriptorSet (i.e. "who imports me").
func buildReverseDepMap(fds *descriptor.FileDescriptorSet) map[string][]string {
	revDeps := make(map[string][]string)
	for _, fd := range fds.GetFile() {
		pkg := fd.GetPackage()
		for _, msg := range fd.GetMessageType() {
			buildReverseDepsFromMsg(pkg, msg, revDeps)
		}
	}
	return revDeps
}

func buildReverseDepsFromMsg(prefix string, msg *descriptor.DescriptorProto, revDeps map[string][]string) {
	fqn := prefix + "." + msg.GetName()
	for _, field := range msg.GetField() {
		typeName := field.GetTypeName()
		if typeName != "" && typeName[0] == '.' {
			dep := typeName[1:]
			revDeps[dep] = append(revDeps[dep], fqn)
		}
	}
	for _, nested := range msg.GetNestedType() {
		buildReverseDepsFromMsg(fqn, nested, revDeps)
	}
}

// findRootFQNs looks up all fully-qualified names in revDeps that match
// schemaName (either exact match or suffix ".schemaName").
func findRootFQNs(revDeps map[string][]string, fds *descriptor.FileDescriptorSet, schemaName string) []string {
	// Collect all known FQNs from the descriptor set.
	allFQNs := make(map[string]struct{})
	for _, fd := range fds.GetFile() {
		pkg := fd.GetPackage()
		for _, msg := range fd.GetMessageType() {
			collectFQNs(pkg, msg, allFQNs)
		}
	}
	// Also gather keys already in revDeps (schemas that are imported by others).
	for k := range revDeps {
		allFQNs[k] = struct{}{}
	}

	var roots []string
	for fqn := range allFQNs {
		if fqn == schemaName || strings.HasSuffix(fqn, "."+schemaName) {
			roots = append(roots, fqn)
		}
	}
	return roots
}

func collectFQNs(prefix string, msg *descriptor.DescriptorProto, out map[string]struct{}) {
	fqn := prefix + "." + msg.GetName()
	out[fqn] = struct{}{}
	for _, nested := range msg.GetNestedType() {
		collectFQNs(fqn, nested, out)
	}
}

// bfsImpacted performs a BFS from the rootFQN over the reverse dependency graph,
// returning every reachable schema (excluding the root itself) with depth and import path.
func bfsImpacted(revDeps map[string][]string, rootFQN string, maxDepth int) []bfsNode {
	type item struct {
		name string
		path []string
	}

	visited := map[string]bool{rootFQN: true}
	queue := []item{{name: rootFQN, path: []string{rootFQN}}}
	var results []bfsNode
	depth := 0

	for len(queue) > 0 && depth < maxDepth {
		depth++
		nextQueue := []item{}
		for _, cur := range queue {
			for _, dep := range revDeps[cur.name] {
				if visited[dep] {
					continue
				}
				visited[dep] = true
				newPath := make([]string, len(cur.path)+1)
				copy(newPath, cur.path)
				newPath[len(cur.path)] = dep
				results = append(results, bfsNode{fqn: dep, depth: depth, importPath: newPath})
				nextQueue = append(nextQueue, item{name: dep, path: newPath})
			}
		}
		queue = nextQueue
	}
	return results
}

type bfsNode struct {
	fqn        string
	depth      int
	importPath []string
}

// computeImpact is the core logic: given raw schema bytes, a root schema name,
// proposed field changes, and a max depth, it returns an ImpactResponse.
func computeImpact(data []byte, namespaceID, schemaName string, fields []FieldChange, maxDepth int) (*ImpactResponse, error) {
	fds, err := changedetector.GetDescriptorSet(data)
	if err != nil {
		return nil, err
	}

	revDeps := buildReverseDepMap(fds)
	roots := findRootFQNs(revDeps, fds, schemaName)

	// Determine the dominant change type (prefer breaking over non-breaking).
	dominantKind := FieldChangeAdded
	for _, f := range fields {
		if f.Change != FieldChangeAdded {
			dominantKind = f.Change
			break
		}
	}
	changeType := fieldChangeToChangeType(dominantKind)

	resp := &ImpactResponse{
		RootSchema: RootSchemaRef{
			NamespaceID: namespaceID,
			SchemaName:  schemaName,
		},
	}

	seen := map[string]bool{}
	for _, rootFQN := range roots {
		nodes := bfsImpacted(revDeps, rootFQN, maxDepth)
		for _, node := range nodes {
			if seen[node.fqn] {
				continue
			}
			seen[node.fqn] = true

			// Convert import path FQNs to short names for readability.
			shortPath := make([]string, len(node.importPath))
			for i, p := range node.importPath {
				shortPath[i] = lastSegment(p)
			}

			resp.ImpactedSchemas = append(resp.ImpactedSchemas, ImpactedSchema{
				NamespaceID: namespaceID,
				SchemaName:  lastSegment(node.fqn),
				ChangeType:  changeType,
				Depth:       node.depth,
				ImportPath:  shortPath,
			})
		}
	}

	// Build summary.
	for _, is := range resp.ImpactedSchemas {
		resp.Summary.TotalImpacted++
		if isBreaking(is.ChangeType) {
			resp.Summary.BreakingCount++
		} else {
			resp.Summary.NonBreakingCount++
		}
	}

	return resp, nil
}

// lastSegment returns the last dot-separated segment of a fully-qualified name.
func lastSegment(fqn string) string {
	if idx := strings.LastIndex(fqn, "."); idx >= 0 {
		return fqn[idx+1:]
	}
	return fqn
}
