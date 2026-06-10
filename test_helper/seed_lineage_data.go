//go:build ignore

// seed_lineage_data.go runs an end-to-end lineage API check:
//  1. compile descriptor from test_helper/lineage_seed.proto
//  2. upload ONE schema container (like real usage: esb-log-entities)
//  3. call lineage using /types/<FQN proto message>/lineage
//  4. assert expected upstream/downstream nodes using full FQNs in the response
//
// Run:
//
//	go run ./test_helper/seed_lineage_data.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	baseURL     = "http://localhost:8080"
	namespaceID = "gotocompany"
	schemaID    = "esb-log-entities"
	protoPath   = "./test_helper/lineage_seed.proto"
)

type lineageResponse struct {
	Direction  string        `json:"direction"`
	Downstream []lineageNode `json:"downstream"`
	Upstream   []lineageNode `json:"upstream"`
	Summary    struct {
		DownstreamCount int `json:"downstream_count"`
		UpstreamCount   int `json:"upstream_count"`
		TotalCount      int `json:"total_count"`
	} `json:"summary"`
}

type lineageNode struct {
	TypeName string `json:"type_name"`
}

type e2eCase struct {
	name               string
	typeName           string
	query              string
	wantStatus         int
	wantDownstreamSet  []string
	wantUpstreamSet    []string
	wantDownstreamSize int
	wantUpstreamSize   int
}

func main() {
	// 1) Compile proto into descriptor bytes
	descBytes := buildDescriptorFromProto()
	fmt.Printf("✓ FileDescriptorSet built (%d bytes)\n", len(descBytes))

	// 2) Create namespace and seed one schema container
	createNamespace()
	deleteSchema(schemaID)
	uploadSchema(schemaID, descBytes)

	// 3) Execute E2E assertions
	runE2ECases()
}

// buildDescriptorFromProto compiles the proto file using protoc and returns
// descriptor bytes used by schema uploads.
func buildDescriptorFromProto() []byte {
	protocBin, err := exec.LookPath("protoc")
	if err != nil {
		log.Fatalf("protoc not found in PATH: %v", err)
	}

	absProtoPath, err := filepath.Abs(protoPath)
	if err != nil {
		log.Fatalf("resolve proto path: %v", err)
	}
	protoDir := filepath.Dir(absProtoPath)

	descOut, err := os.CreateTemp("", "lineage_seed_*.desc")
	if err != nil {
		log.Fatalf("create temp descriptor file: %v", err)
	}
	descOutPath := descOut.Name()
	_ = descOut.Close()
	defer os.Remove(descOutPath)

	cmd := exec.Command(
		protocBin,
		"-I", protoDir,
		fmt.Sprintf("--descriptor_set_out=%s", descOutPath),
		"--include_imports",
		absProtoPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("protoc failed: %v\n%s", err, string(out))
	}

	b, err := os.ReadFile(descOutPath)
	if err != nil {
		log.Fatalf("read descriptor file: %v", err)
	}
	return b
}

func createNamespace() {
	body, _ := json.Marshal(map[string]interface{}{
		"id":            namespaceID,
		"format":        "FORMAT_PROTOBUF",
		"compatibility": "COMPATIBILITY_BACKWARD",
		"description":   "Lineage E2E seed namespace",
	})

	resp, err := http.Post(baseURL+"/v1beta1/namespaces", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("create namespace: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		fmt.Printf("✓ Namespace '%s' created (HTTP %d)\n", namespaceID, resp.StatusCode)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("ℹ Namespace '%s' already exists — skipping\n", namespaceID)
	} else {
		fmt.Printf("⚠ Namespace create response (HTTP %d): %s\n", resp.StatusCode, string(respBody))
	}
}

func deleteSchema(schemaName string) {
	url := fmt.Sprintf("%s/v1beta1/namespaces/%s/schemas/%s", baseURL, namespaceID, schemaName)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return // best-effort
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("⟳ Schema '%s/%s' deleted before re-seed\n", namespaceID, schemaName)
	}
}

func uploadSchema(schemaName string, descBytes []byte) {
	url := fmt.Sprintf("%s/v1beta1/namespaces/%s/schemas/%s", baseURL, namespaceID, schemaName)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(descBytes))
	req.Header.Set("X-Format", "FORMAT_PROTOBUF")
	req.Header.Set("X-Compatibility", "COMPATIBILITY_BACKWARD")
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("upload schema: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		fmt.Printf("✓ Schema '%s/%s' uploaded (HTTP %d): %s\n", namespaceID, schemaName, resp.StatusCode, string(respBody))
	} else {
		log.Fatalf("upload schema failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}
}

func runE2ECases() {
	base := fmt.Sprintf("%s/v1beta1/namespaces/%s/schemas/%s", baseURL, namespaceID, schemaID)

	cases := []e2eCase{
		{
			name:               "container + type_name downstream location-like root",
			typeName:           "gotocompany.events.Product",
			query:              "direction=downstream",
			wantStatus:         http.StatusOK,
			wantDownstreamSet:  []string{"gotocompany.events.Order.Item", "gotocompany.events.Order", "gotocompany.events.Cart", "gotocompany.events.Payment"},
			wantDownstreamSize: 4,
		},
		{
			name:               "container + type_name downstream nested receipt root",
			typeName:           "gotocompany.events.Payment.Receipt",
			query:              "direction=downstream",
			wantStatus:         http.StatusOK,
			wantDownstreamSet:  []string{"gotocompany.events.Payment", "gotocompany.events.Ledger"},
			wantDownstreamSize: 2,
		},
		{
			name:             "container + type_name upstream cart",
			typeName:         "gotocompany.events.Cart",
			query:            "direction=upstream",
			wantStatus:       http.StatusOK,
			wantUpstreamSet:  []string{"gotocompany.events.User", "gotocompany.events.Order.Item", "gotocompany.events.Address", "gotocompany.events.Product"},
			wantUpstreamSize: 4,
		},
		{
			name:       "invalid direction",
			typeName:   "gotocompany.events.Product",
			query:      "direction=sideways",
			wantStatus: http.StatusBadRequest,
		},

		// ── enum use-cases ────────────────────────────────────────────────────────

		// DataMessage depends on User (→ Address), DataTypes.Enum (→ DataTypes), and Status.
		{
			name:       "upstream of DataMessage includes user chain + enum deps",
			typeName:   "gotocompany.events.DataMessage",
			query:      "direction=upstream",
			wantStatus: http.StatusOK,
			wantUpstreamSet: []string{
				"gotocompany.events.User",
				"gotocompany.events.Address",
				"gotocompany.events.DataTypes",
				"gotocompany.events.DataTypes.Enum",
				"gotocompany.events.Status",
			},
			wantUpstreamSize: 5,
		},

		// DataTypes.Enum (nested enum) is referenced by DataMessage.
		{
			name:               "downstream of nested enum DataTypes.Enum",
			typeName:           "gotocompany.events.DataTypes.Enum",
			query:              "direction=downstream",
			wantStatus:         http.StatusOK,
			wantDownstreamSet:  []string{"gotocompany.events.DataMessage"},
			wantDownstreamSize: 1,
		},

		// DataTypes (the wrapper message) transitively reaches DataMessage through DataTypes.Enum.
		{
			name:               "downstream of DataTypes wrapper message",
			typeName:           "gotocompany.events.DataTypes",
			query:              "direction=downstream",
			wantStatus:         http.StatusOK,
			wantDownstreamSet:  []string{"gotocompany.events.DataTypes.Enum", "gotocompany.events.DataMessage"},
			wantDownstreamSize: 2,
		},

		// Status (top-level enum) is referenced by both Order and DataMessage.
		{
			name:               "downstream of top-level enum Status",
			typeName:           "gotocompany.events.Status",
			query:              "direction=downstream",
			wantStatus:         http.StatusOK,
			wantDownstreamSet:  []string{"gotocompany.events.Order", "gotocompany.events.DataMessage"},
			wantDownstreamSize: 2,
		},
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🧪 Running lineage E2E cases")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, tc := range cases {
		url := fmt.Sprintf("%s/types/%s/lineage?%s", base, tc.typeName, tc.query)
		if err := runCase(tc, url); err != nil {
			log.Fatalf("✗ %s: %v", tc.name, err)
		}
		fmt.Printf("✓ %s\n", tc.name)
	}

	fmt.Println("\n✅ All lineage E2E cases passed")
}

func runCase(tc e2eCase, url string) error {
	resp, body, err := doGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != tc.wantStatus {
		return fmt.Errorf("status mismatch: got=%d want=%d body=%s", resp.StatusCode, tc.wantStatus, string(body))
	}

	if tc.wantStatus != http.StatusOK {
		return nil
	}

	var parsed lineageResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("decode response: %w body=%s", err, string(body))
	}

	downSet := extractNames(parsed.Downstream)
	upSet := extractNames(parsed.Upstream)

	if tc.wantDownstreamSize >= 0 && tc.wantDownstreamSize != len(downSet) && len(tc.wantDownstreamSet) > 0 {
		return fmt.Errorf("downstream size mismatch: got=%d want=%d gotSet=%v", len(downSet), tc.wantDownstreamSize, downSet)
	}
	if tc.wantUpstreamSize >= 0 && tc.wantUpstreamSize != len(upSet) && len(tc.wantUpstreamSet) > 0 {
		return fmt.Errorf("upstream size mismatch: got=%d want=%d gotSet=%v", len(upSet), tc.wantUpstreamSize, upSet)
	}

	if err := ensureContainsAll(downSet, tc.wantDownstreamSet, "downstream"); err != nil {
		return err
	}
	if err := ensureContainsAll(upSet, tc.wantUpstreamSet, "upstream"); err != nil {
		return err
	}

	return nil
}

func doGet(url string) (*http.Response, []byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("http get %s: %w", url, err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return nil, nil, fmt.Errorf("read body: %w", err)
	}
	return resp, body, nil
}

func extractNames(nodes []lineageNode) []string {
	set := map[string]struct{}{}
	for _, n := range nodes {
		if n.TypeName == "" {
			continue
		}
		set[n.TypeName] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for n := range set {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func ensureContainsAll(got []string, want []string, label string) error {
	if len(want) == 0 {
		return nil
	}
	gotMap := map[string]struct{}{}
	for _, g := range got {
		gotMap[g] = struct{}{}
	}
	missing := []string{}
	for _, w := range want {
		if _, ok := gotMap[w]; !ok {
			missing = append(missing, w)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s missing expected nodes: %s (got=%v)", label, strings.Join(missing, ", "), got)
	}
	return nil
}
