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
// Nested messages (Order.Item, Payment.Receipt) are traversed automatically via the
// same descriptor — no separate upload needed for them.
var schemaNames = []string{"Address", "User", "Product", "Invoice", "Order", "Payment", "Item", "Receipt", "Cart", "Ledger"}

func main() {
	// ── 1. Build FileDescriptorSet from proto ───────────────────────────────
	descBytes := buildDescriptorFromProto()
	fmt.Printf("✓ FileDescriptorSet built (%d bytes)\n", len(descBytes))

	// ── 2. Create namespace ─────────────────────────────────────────────────
	createNamespace()

	// ── 3. Upload schema under each message name ─────────────────────────
	for _, name := range schemaNames {
		deleteSchema(name) // clear stale versions before re-seeding
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

func printCurls() {
	type cas struct{ label, url string }

	base := fmt.Sprintf("%s/v1beta1/namespaces/%s/schemas", baseURL, namespaceID)

	cases := []cas{
		{
			"Lineage for 'User' — both, level=10\n  Expected: upstream=[Address], downstream=[Order, Payment]",
			base + "/User/lineage?level=10&direction=both",
		},
		{
			"Lineage for 'User' — upstream only\n  Expected: [Address]",
			base + "/User/lineage?direction=upstream",
		},
		{
			"Lineage for 'User' — downstream only\n  Expected: [Order, Payment]",
			base + "/User/lineage?direction=downstream",
		},
		{
			"Lineage for 'User' — downstream, level=1 (direct only)\n  Expected: [Order]",
			base + "/User/lineage?direction=downstream&level=1",
		},
		{
			"Lineage for 'Order' — both\n  Expected: upstream=[User, Address], downstream=[Payment]",
			base + "/Order/lineage?direction=both",
		},
		{
			"Lineage for 'Address' — downstream\n  Expected: [User, Order, Payment]",
			base + "/Address/lineage?direction=downstream",
		},
		{
			// Nested message: Order.Item depends on Product;
			// GetLineage for 'Product' will find it as the root FQN 'gotocompany.events.Order.Item.Product'? No —
			// Product is a top-level message. Order.Item's field has TypeName='.gotocompany.events.Product'.
			// Lineage(Product) downstream → Order.Item (nested msg inside Order)
			"Inner use-case: Lineage for 'Product' — downstream\n  Expected: downstream=[Item] (Order.Item depends on Product)",
			base + "/Product/lineage?direction=downstream",
		},
		{
			// Nested message: Payment.Receipt depends on Invoice
			"Inner use-case: Lineage for 'Invoice' — downstream\n  Expected: downstream=[Receipt] (Payment.Receipt depends on Invoice)",
			base + "/Invoice/lineage?direction=downstream",
		},
		{
			"Outer ref use-case: Lineage for 'Cart' — upstream\n  Expected: upstream=[Item, Product, User, Address] (Cart uses Order.Item which uses Product; also uses User)",
			base + "/Cart/lineage?direction=upstream",
		},
		{
			"Outer ref use-case: Lineage for 'Ledger' — upstream\n  Expected: upstream=[Receipt, Invoice] (Ledger uses Payment.Receipt which uses Invoice)",
			base + "/Ledger/lineage?direction=upstream",
		},
		{
			"Outer ref use-case: Lineage for 'Item' — downstream\n  Expected: downstream=[Order, Payment, Cart] (Item is used by both Order (parent) and Cart (outer ref))",
			base + "/Item/lineage?direction=downstream",
		},
		{
			"Outer ref use-case: Lineage for 'Receipt' — downstream\n  Expected: downstream=[Payment, Ledger] (Receipt is used by both Payment (parent) and Ledger (outer ref))",
			base + "/Receipt/lineage?direction=downstream",
		},
		{
			"Invalid direction — expects HTTP 400",
			base + "/User/lineage?direction=sideways",
		},
	}

	for _, c := range cases {
		fmt.Printf("\n# %s\ncurl -s '%s' | python3 -m json.tool\n", c.label, c.url)
	}

	fmt.Println()
	fmt.Println("# ── Other useful endpoints ─────────────────────────────────")
	fmt.Printf("\n# List namespaces\ncurl -s '%s/v1beta1/namespaces' | python3 -m json.tool\n", baseURL)
	fmt.Printf("\n# List schemas in namespace '%s'\ncurl -s '%s/v1beta1/namespaces/%s/schemas' | python3 -m json.tool\n", namespaceID, baseURL, namespaceID)
}
