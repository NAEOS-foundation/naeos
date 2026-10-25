// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package publicadapter

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/metadata"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/schemaregistry"
)

func TestToPublicRoundTripsAndMatchesModelSchema(t *testing.T) {
	source := &model.NEIR{
		Project: &project.Project{
			Name: "public-neir-check", Version: "1.2.3", Authors: []string{"NAEOS"},
		},
		Modules: []module.Module{{
			Name: "core", Path: "./core", Dependencies: []string{"shared"},
		}},
		Services: []service.Service{{
			Name: "api", Kind: service.KindHTTP, Port: 8080,
			Endpoints: []service.Endpoint{{Method: "GET", Path: "/health", Action: "health"}},
		}},
		Metadata: &metadata.Metadata{
			NEIRVersion: "2.0.0", SchemaVersion: "0.5.0", ProjectVersion: "0.1.0",
			Source: &metadata.SourceRef{Kind: "file", Ref: "naeos.yaml"},
		},
	}

	public, err := ToPublic(source)
	if err != nil {
		t.Fatalf("ToPublic: %v", err)
	}
	if public.Project == nil || public.Project.Name != "public-neir-check" || len(public.Modules) != 1 || len(public.Services) != 1 {
		t.Fatalf("public DTO lost typed fields: %#v", public)
	}

	data, err := json.Marshal(public)
	if err != nil {
		t.Fatalf("marshal public DTO: %v", err)
	}
	var got, want any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	internalData, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(internalData, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("public conversion changed model JSON\n got: %#v\nwant: %#v", got, want)
	}

	client := schemaregistry.NewNEIRClient("file://../../../site/public/schemaregistry/v1/neir.json")
	schema, err := client.FetchSchema()
	if err != nil {
		t.Fatalf("load public model schema: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	validation := schemaregistry.ValidateJSONSchema(document, schema)
	if !validation.Valid {
		t.Fatalf("public DTO failed model schema validation: %+v", validation.Errors)
	}
}

func TestToPublicNil(t *testing.T) {
	document, err := ToPublic(nil)
	if err != nil || document != nil {
		t.Fatalf("ToPublic(nil) = (%v, %v), want (nil, nil)", document, err)
	}
}
