package schema

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	"github.com/goto/stencil/core/changedetector"
)

type LineageDirection string

const (
	LineageDirectionBoth       LineageDirection = "both"
	LineageDirectionDownstream LineageDirection = "downstream"
	LineageDirectionUpstream   LineageDirection = "upstream"
)

// RootSchemaRef identifies the schema on which lineage is requested.
type RootSchemaRef struct {
	NamespaceID string `json:"namespace_id"`
	SchemaID    string `json:"schema_id"`
	TypeName    string `json:"type_name"`
}

// LineageSchema represents a single schema in the lineage graph.
type LineageSchema struct {
	NamespaceID string   `json:"namespace_id"`
	SchemaID    string   `json:"schema_id"`
	TypeName    string   `json:"type_name"`
	Level       int      `json:"level"`
	Path        []string `json:"path"`
}

// LineageSummary aggregates lineage counts.
type LineageSummary struct {
	DownstreamCount int `json:"downstream_count"`
	UpstreamCount   int `json:"upstream_count"`
	TotalCount      int `json:"total_count"`
}

// LineageResponse is the full response body for the lineage endpoint.
type LineageResponse struct {
	RootSchema RootSchemaRef    `json:"root_schema"`
	Direction  LineageDirection `json:"direction"`
	Downstream []LineageSchema  `json:"downstream,omitempty"`
	Upstream   []LineageSchema  `json:"upstream,omitempty"`
	Summary    LineageSummary   `json:"summary"`
}

func NormalizeLineageDirection(direction LineageDirection) (LineageDirection, error) {
	switch direction {
	case "", LineageDirectionBoth:
		return LineageDirectionBoth, nil
	case LineageDirectionDownstream:
		return LineageDirectionDownstream, nil
	case LineageDirectionUpstream:
		return LineageDirectionUpstream, nil
	default:
		return "", fmt.Errorf("invalid direction: %q", direction)
	}
}

// buildDependencyMaps constructs both forward and reverse dependency maps.
func buildDependencyMaps(fds *descriptor.FileDescriptorSet) (map[string][]string, map[string][]string) {
	forwardDeps := make(map[string][]string)
	reverseDeps := make(map[string][]string)
	for _, fd := range fds.GetFile() {
		pkg := fd.GetPackage()
		for _, msg := range fd.GetMessageType() {
			buildDepsFromMsg(pkg, msg, forwardDeps, reverseDeps)
		}
	}
	return forwardDeps, reverseDeps
}

func buildDepsFromMsg(prefix string, msg *descriptor.DescriptorProto, forwardDeps, reverseDeps map[string][]string) {
	fqn := prefix + "." + msg.GetName()
	for _, field := range msg.GetField() {
		typeName := field.GetTypeName()
		if typeName != "" && typeName[0] == '.' {
			dep := typeName[1:]
			forwardDeps[fqn] = append(forwardDeps[fqn], dep)
			reverseDeps[dep] = append(reverseDeps[dep], fqn)
		}
	}
	for _, nested := range msg.GetNestedType() {
		buildDepsFromMsg(fqn, nested, forwardDeps, reverseDeps)
	}
}

// findRootFQNs looks up all fully-qualified names that match
// schemaName (either exact match or suffix ".schemaName").
func findRootFQNs(fds *descriptor.FileDescriptorSet, schemaName string) []string {
	allFQNs := make(map[string]struct{})
	for _, fd := range fds.GetFile() {
		pkg := fd.GetPackage()
		for _, msg := range fd.GetMessageType() {
			collectFQNs(pkg, msg, allFQNs)
		}
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

// walkLineage performs a BFS from the rootFQN over the provided graph,
// returning every reachable schema (excluding the root itself) with level and path.
func walkLineage(graph map[string][]string, rootFQN string, maxLevel int) []bfsNode {
	type item struct {
		name string
		path []string
	}

	visited := map[string]bool{rootFQN: true}
	queue := []item{{name: rootFQN, path: []string{rootFQN}}}
	var results []bfsNode
	level := 0

	for len(queue) > 0 && level < maxLevel {
		level++
		nextQueue := []item{}
		for _, cur := range queue {
			for _, relation := range graph[cur.name] {
				if visited[relation] {
					continue
				}
				visited[relation] = true
				newPath := make([]string, len(cur.path)+1)
				copy(newPath, cur.path)
				newPath[len(cur.path)] = relation
				results = append(results, bfsNode{fqn: relation, level: level, path: newPath})
				nextQueue = append(nextQueue, item{name: relation, path: newPath})
			}
		}
		queue = nextQueue
	}
	return results
}

type bfsNode struct {
	fqn   string
	level int
	path  []string
}

func convertLineageNodes(namespaceID, schemaID string, nodes []bfsNode) []LineageSchema {
	lineage := make([]LineageSchema, 0, len(nodes))
	for _, node := range nodes {
		fullPath := make([]string, len(node.path))
		copy(fullPath, node.path)
		lineage = append(lineage, LineageSchema{
			NamespaceID: namespaceID,
			SchemaID:    schemaID,
			TypeName:    node.fqn,
			Level:       node.level,
			Path:        fullPath,
		})
	}
	return lineage
}

// computeLineage is the core logic: given raw schema bytes, a root schema name,
// level limit, and traversal direction, it returns a LineageResponse.
func computeLineage(data []byte, namespaceID, schemaID, rootType string, level int, direction LineageDirection) (*LineageResponse, error) {
	fds, err := changedetector.GetDescriptorSet(data)
	if err != nil {
		return nil, err
	}
	direction, err = NormalizeLineageDirection(direction)
	if err != nil {
		return nil, err
	}

	forwardDeps, reverseDeps := buildDependencyMaps(fds)
	roots := findRootFQNs(fds, rootType)

	resp := &LineageResponse{
		RootSchema: RootSchemaRef{
			NamespaceID: namespaceID,
			SchemaID:    schemaID,
			TypeName:    rootType,
		},
		Direction: direction,
	}

	downstreamSeen := map[string]bool{}
	upstreamSeen := map[string]bool{}
	for _, rootFQN := range roots {
		if direction == LineageDirectionBoth || direction == LineageDirectionDownstream {
			nodes := walkLineage(reverseDeps, rootFQN, level)
			filtered := make([]bfsNode, 0, len(nodes))
			for _, node := range nodes {
				if downstreamSeen[node.fqn] {
					continue
				}
				downstreamSeen[node.fqn] = true
				filtered = append(filtered, node)
			}
			resp.Downstream = append(resp.Downstream, convertLineageNodes(namespaceID, schemaID, filtered)...)
		}

		if direction == LineageDirectionBoth || direction == LineageDirectionUpstream {
			nodes := walkLineage(forwardDeps, rootFQN, level)
			filtered := make([]bfsNode, 0, len(nodes))
			for _, node := range nodes {
				if upstreamSeen[node.fqn] {
					continue
				}
				upstreamSeen[node.fqn] = true
				filtered = append(filtered, node)
			}
			resp.Upstream = append(resp.Upstream, convertLineageNodes(namespaceID, schemaID, filtered)...)
		}
	}

	resp.Summary.DownstreamCount = len(resp.Downstream)
	resp.Summary.UpstreamCount = len(resp.Upstream)
	resp.Summary.TotalCount = resp.Summary.DownstreamCount + resp.Summary.UpstreamCount

	return resp, nil
}

// lastSegment returns the last dot-separated segment of a fully-qualified name.
func lastSegment(fqn string) string {
	if idx := strings.LastIndex(fqn, "."); idx >= 0 {
		return fqn[idx+1:]
	}
	return fqn
}
