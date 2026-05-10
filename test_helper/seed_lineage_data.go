//go:build ignore

// seed_lineage_data.go seeds the local Stencil server with protobuf schemas
// that demonstrate multi-level lineage relationships.
//
// Schema hierarchy (package: gotocompany.data):
//
//	Address  {}
//	User     { address Address }        → User depends on Address
//	Order    { user    User    }        → Order depends on User
//	Payment  { order   Order   }        → Payment depends on Order
//
// Lineage for "User":
//
//	upstream   → Address
//	downstream → Order → Payment
//
// Run: go run ./test_helper/seed_lineage_data.go
//
// By default this script compiles ./test_helper/lineage_seed.proto to a
// temporary descriptor set via protoc.
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
)

const (
	baseURL     = "http://localhost:8080"
	namespaceID = "gotocompany"
	protoPath   = "./test_helper/lineage_seed.proto"
)

// Each schema name matches a proto message name so GetLineage can find the root FQN.
var schemaNames = []string{"Address", "User", "Order", "Payment"}

func main() {
	// ── 1. Build FileDescriptorSet from proto ───────────────────────────────
	descBytes := buildDescriptorFromProto()
	fmt.Printf("✓ FileDescriptorSet built (%d bytes)\n", len(descBytes))

	// ── 2. Create namespace ─────────────────────────────────────────────────
	createNamespace()

	// ── 3. Upload schema under each message name ─────────────────────────
	for _, name := range schemaNames {
		uploadSchema(name, descBytes)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅  Seed complete! Use the curls below to test the lineage API:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	printCurls()
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

func printCurls() {
	// Lineage is keyed by the schema name = proto message name
	// Our main focus is "User": upstream=Address, downstream=Order→Payment
	lineageBase := fmt.Sprintf("%s/v1beta1/namespaces/%s/schemas/User/lineage", baseURL, namespaceID)
	cases := []struct {
		label string
		url   string
	}{
		{"Lineage for 'User' — direction=both (default), level=10\n  Expected: upstream=[Address], downstream=[Order, Payment]", lineageBase + "?level=10&direction=both"},
		{"Lineage for 'User' — upstream only\n  Expected: [Address]", lineageBase + "?direction=upstream"},
		{"Lineage for 'User' — downstream only\n  Expected: [Order, Payment]", lineageBase + "?direction=downstream"},
		{"Lineage for 'User' — downstream, max 1 level (direct only)\n  Expected: [Order]", lineageBase + "?direction=downstream&level=1"},
		{"Lineage for 'Order' — both\n  Expected: upstream=[User, Address], downstream=[Payment]", fmt.Sprintf("%s/v1beta1/namespaces/%s/schemas/Order/lineage?direction=both", baseURL, namespaceID)},
		{"Lineage for 'Address' — downstream only\n  Expected: [User, Order, Payment]", fmt.Sprintf("%s/v1beta1/namespaces/%s/schemas/Address/lineage?direction=downstream", baseURL, namespaceID)},
		{"Invalid direction — expects HTTP 400", lineageBase + "?direction=sideways"},
	}

	for _, c := range cases {
		fmt.Printf("\n# %s\ncurl -s '%s' | python3 -m json.tool\n", c.label, c.url)
	}

	fmt.Println()
	fmt.Println("# ── Other useful endpoints ─────────────────────────────────")
	fmt.Printf("\n# List namespaces\ncurl -s '%s/v1beta1/namespaces' | python3 -m json.tool\n", baseURL)
	fmt.Printf("\n# List schemas in namespace '%s'\ncurl -s '%s/v1beta1/namespaces/%s/schemas' | python3 -m json.tool\n", namespaceID, baseURL, namespaceID)
}
