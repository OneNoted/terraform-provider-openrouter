package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderSchemaMarksManagementAPIKeySensitive(t *testing.T) {
	p := New()
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	attr, ok := resp.Schema.Attributes["management_api_key"]
	if !ok {
		t.Fatal("management_api_key attribute missing")
	}
	if !attr.IsSensitive() {
		t.Fatal("management_api_key must be sensitive")
	}
}
