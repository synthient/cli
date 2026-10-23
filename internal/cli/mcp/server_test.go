package mcp

import (
	"encoding/json"
	"net/netip"
	"testing"

	"github.com/synthient/synthient-go"
)

func TestIPLookupSchemaAcceptsAddress(t *testing.T) {
	schema, err := ipLookupSchema()
	if err != nil {
		t.Fatalf("ipLookupSchema returned error: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatalf("resolving schema: %v", err)
	}

	output := ipLookupOutput{IPs: []synthient.IPInfo{{IP: netip.MustParseAddr("8.8.8.8")}}}
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("marshaling output: %v", err)
	}
	var value any
	err = json.Unmarshal(data, &value)
	if err != nil {
		t.Fatalf("unmarshaling output: %v", err)
	}

	err = resolved.Validate(value)
	if err != nil {
		t.Fatalf("output does not match schema: %v", err)
	}
}
