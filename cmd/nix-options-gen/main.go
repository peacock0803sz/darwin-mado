// Package main implements the JSON Schema -> Nix options codegen script.
// It reads config.v1.schema.json and emits nix/generated-options.nix.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// JSONSchema represents a JSON Schema Draft-07 document or sub-schema.
type JSONSchema struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Properties  map[string]*JSONSchema `json:"properties"`
	Items       *JSONSchema            `json:"items"`
	Enum        []string               `json:"enum"`
	Pattern     string                 `json:"pattern"`
	Default     any                    `json:"default"`
	Required    []string               `json:"required"`
	AnyOf       []*JSONSchema          `json:"anyOf"`
	Minimum     *float64               `json:"minimum"`
	MinItems    *int                   `json:"minItems"`
	MaxItems    *int                   `json:"maxItems"`
	MinLength   *int                   `json:"minLength"`
}

// assertCtx carries the traversal context for assertion generation.
// As we recurse into nested schemas, wrapFn grows with builtins.all lambdas.
type assertCtx struct {
	// nullGuards: top-level null checks; if any is true, skip the whole assertion
	nullGuards []string
	// wrapFn wraps an inner boolean Nix expression with builtins.all traversals
	wrapFn func(inner string) string
	// refFn returns the Nix expression for a named field in the current item scope
	refFn func(fieldName string) string
}

var rootCtx = assertCtx{
	wrapFn: func(inner string) string { return inner },
	refFn:  func(name string) string { return "cfg.settings." + name },
}

type nixAssert struct {
	condition string
	message   string
}

type generator struct {
	asserts []nixAssert
}

func main() {
	schemaPath := flag.String("schema", "", "path to JSON Schema file")
	outputPath := flag.String("output", "", "path to output Nix file")
	flag.Parse()

	if *schemaPath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: nix-options-gen -schema <path> -output <path>")
		os.Exit(1)
	}

	data, err := os.ReadFile(*schemaPath) //nolint:gosec // G304: path is CLI arg
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading schema: %v\n", err)
		os.Exit(1)
	}

	var schema JSONSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing schema: %v\n", err)
		os.Exit(1)
	}

	g := &generator{}
	out := g.generate(&schema, *schemaPath)

	if err := os.WriteFile(*outputPath, []byte(out), 0o644); err != nil { //nolint:gosec // G306
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}
}

func (g *generator) generate(schema *JSONSchema, schemaPath string) string {
	var sb strings.Builder
	sb.WriteString("# AUTO-GENERATED from " + schemaPath + "\n")
	sb.WriteString("# Do not edit manually. Run: go run ./cmd/nix-options-gen -schema schemas/config.v1.schema.json -output nix/generated-options.nix\n")
	sb.WriteString("{ lib }:\n{\n  options = {\n")

	for _, name := range sortedKeys(schema.Properties) {
		prop := schema.Properties[name]
		// R-007: all top-level settings use nullOr with default = null
		sb.WriteString(g.emitOption(name, prop, true, 2, rootCtx))
	}

	sb.WriteString("  };\n  assertions = cfg: [\n")
	for _, a := range g.asserts {
		sb.WriteString("    {\n")
		sb.WriteString("      assertion = " + a.condition + ";\n")
		sb.WriteString("      message = " + nixStr(a.message) + ";\n")
		sb.WriteString("    }\n")
	}
	sb.WriteString("  ];\n}\n")
	return sb.String()
}

// emitOption writes a single lib.mkOption declaration at the given indent level.
func (g *generator) emitOption(name string, prop *JSONSchema, nullable bool, indent int, ctx assertCtx) string {
	pad := strings.Repeat("  ", indent)
	var sb strings.Builder

	nixType := g.buildType(name, prop, nullable, indent, ctx)
	desc := prop.Description
	if desc == "" {
		desc = name
	}

	sb.WriteString(pad + name + " = lib.mkOption {\n")
	sb.WriteString(pad + "  type = " + nixType + ";\n")
	if nullable {
		sb.WriteString(pad + "  default = null;\n")
	}
	sb.WriteString(pad + "  description = " + nixStr(desc) + ";\n")
	sb.WriteString(pad + "};\n")
	return sb.String()
}

// buildType returns the Nix type expression, wrapping in nullOr when nullable.
func (g *generator) buildType(name string, prop *JSONSchema, nullable bool, indent int, ctx assertCtx) string {
	base := g.baseType(name, prop, nullable, indent, ctx)
	if nullable {
		return "lib.types.nullOr (" + base + ")"
	}
	return base
}

// baseType returns the unwrapped Nix type, generating assertions as a side effect.
// nullable indicates whether the field is optional (affects assertion null guards).
func (g *generator) baseType(name string, prop *JSONSchema, nullable bool, indent int, ctx assertCtx) string {
	switch prop.Type {
	case "string":
		if len(prop.Enum) > 0 {
			return "lib.types.enum [ " + joinEnum(prop.Enum) + " ]"
		}
		if prop.Pattern != "" {
			fieldExpr := ctx.refFn(name)
			inner := "builtins.match " + nixStr(prop.Pattern) + " " + fieldExpr + " != null"
			if nullable {
				inner = fieldExpr + " == null || " + inner
			}
			wrapped := ctx.wrapFn(inner)
			g.asserts = append(g.asserts, nixAssert{
				condition: makeGuard(ctx.nullGuards) + wrapped,
				message:   name + " must match pattern " + prop.Pattern,
			})
		}
		return "lib.types.str"

	case "integer":
		if prop.Minimum != nil {
			if *prop.Minimum >= 1 {
				return "lib.types.ints.positive"
			}
			if *prop.Minimum == 0 {
				return "lib.types.ints.unsigned"
			}
		}
		return "lib.types.int"

	case "array":
		fieldExpr := ctx.refFn(name)
		if prop.MinItems != nil {
			minItems := *prop.MinItems
			inner := fmt.Sprintf("builtins.length %s >= %d", fieldExpr, minItems)
			if nullable {
				inner = fieldExpr + " == null || " + inner
			}
			g.asserts = append(g.asserts, nixAssert{
				condition: makeGuard(ctx.nullGuards) + ctx.wrapFn(inner),
				message:   fmt.Sprintf("%s must have at least %d item(s)", name, minItems),
			})
		}
		if prop.MaxItems != nil {
			maxItems := *prop.MaxItems
			inner := fmt.Sprintf("builtins.length %s <= %d", fieldExpr, maxItems)
			if nullable {
				inner = fieldExpr + " == null || " + inner
			}
			g.asserts = append(g.asserts, nixAssert{
				condition: makeGuard(ctx.nullGuards) + ctx.wrapFn(inner),
				message:   fmt.Sprintf("%s must have at most %d item(s)", name, maxItems),
			})
		}
		if prop.Items != nil {
			itemType := g.itemType(name, prop.Items, nullable, indent, ctx)
			return "lib.types.listOf (" + itemType + ")"
		}
		return "lib.types.listOf lib.types.str"

	case "object":
		if len(prop.Properties) > 0 {
			return g.submoduleType(prop, indent, ctx)
		}
		return "lib.types.attrs"
	}
	return "lib.types.str"
}

// itemType generates the Nix type for array items, updating context for object items.
func (g *generator) itemType(listName string, items *JSONSchema, listNullable bool, indent int, ctx assertCtx) string {
	if items.Type == "object" {
		binding := pickBinding(listName)
		listExpr := ctx.refFn(listName)
		parentWrap := ctx.wrapFn

		// New guards: outer guards + the list itself being null
		newGuards := make([]string, len(ctx.nullGuards))
		copy(newGuards, ctx.nullGuards)
		if listNullable {
			newGuards = append(newGuards, listExpr+" == null")
		}

		itemCtx := assertCtx{
			nullGuards: newGuards,
			wrapFn: func(inner string) string {
				return parentWrap("builtins.all (" + binding + ": " + inner + ") " + listExpr)
			},
			refFn: func(fieldName string) string { return binding + "." + fieldName },
		}

		// anyOf on the items schema (e.g., rule must have position or size)
		if len(items.AnyOf) > 0 {
			g.addAnyOfAssertion(items.AnyOf, itemCtx, binding)
		}

		return g.submoduleType(items, indent, itemCtx)
	}

	// String items with minLength: generate assertion at list item level
	if items.Type == "string" && items.MinLength != nil && *items.MinLength > 0 {
		binding := "s"
		listExpr := ctx.refFn(listName)
		parentWrap := ctx.wrapFn
		newGuards := make([]string, len(ctx.nullGuards))
		copy(newGuards, ctx.nullGuards)
		if listNullable {
			newGuards = append(newGuards, listExpr+" == null")
		}
		inner := fmt.Sprintf("builtins.stringLength %s >= %d", binding, *items.MinLength)
		wrapped := parentWrap("builtins.all (" + binding + ": " + inner + ") " + listExpr)
		g.asserts = append(g.asserts, nixAssert{
			condition: makeGuard(newGuards) + wrapped,
			message:   listName + " items must be non-empty strings",
		})
	}

	// Primitive item type (no new context needed, just the type)
	return g.primitiveType(items)
}

// primitiveType returns a Nix type for a primitive schema (no assertions).
func (g *generator) primitiveType(prop *JSONSchema) string {
	switch prop.Type {
	case "integer":
		if prop.Minimum != nil {
			if *prop.Minimum >= 1 {
				return "lib.types.ints.positive"
			}
			if *prop.Minimum == 0 {
				return "lib.types.ints.unsigned"
			}
		}
		return "lib.types.int"
	case "string":
		return "lib.types.str"
	}
	return "lib.types.str"
}

// submoduleType generates a lib.types.submodule { options = {...}; } expression.
// indent is the indent level of the containing lib.mkOption declaration.
func (g *generator) submoduleType(schema *JSONSchema, indent int, ctx assertCtx) string {
	// options = { at indent+2, properties at indent+3, closing } at indent+1
	padOpts := strings.Repeat("  ", indent+2)
	padClose := strings.Repeat("  ", indent+1)

	var sb strings.Builder
	sb.WriteString("lib.types.submodule {\n")
	sb.WriteString(padOpts + "options = {\n")

	requiredSet := toSet(schema.Required)
	for _, name := range sortedKeys(schema.Properties) {
		prop := schema.Properties[name]
		nullable := !requiredSet[name]
		sb.WriteString(g.emitOption(name, prop, nullable, indent+3, ctx))
	}

	sb.WriteString(padOpts + "};\n")
	sb.WriteString(padClose + "}")
	return sb.String()
}

// addAnyOfAssertion handles anyOf on an items schema (e.g., rule must have position or size).
func (g *generator) addAnyOfAssertion(anyOf []*JSONSchema, ctx assertCtx, binding string) {
	parts := make([]string, 0, len(anyOf))
	for _, alt := range anyOf {
		if len(alt.Required) > 0 {
			fieldParts := make([]string, 0, len(alt.Required))
			for _, req := range alt.Required {
				fieldParts = append(fieldParts, binding+"."+req+" != null")
			}
			parts = append(parts, "("+strings.Join(fieldParts, " && ")+")")
		}
	}
	if len(parts) == 0 {
		return
	}
	inner := strings.Join(parts, " || ")
	g.asserts = append(g.asserts, nixAssert{
		condition: makeGuard(ctx.nullGuards) + ctx.wrapFn(inner),
		message:   "Each preset rule must have at least 'position' or 'size'",
	})
}

// makeGuard produces the null-guard prefix: "A == null || B == null || " or "".
func makeGuard(guards []string) string {
	if len(guards) == 0 {
		return ""
	}
	return strings.Join(guards, " || ") + " || "
}

func pickBinding(listName string) string {
	switch listName {
	case "presets":
		return "p"
	case "rules":
		return "r"
	default:
		if len(listName) > 0 {
			return string(listName[0])
		}
		return "x"
	}
}

func joinEnum(vals []string) string {
	quoted := make([]string, len(vals))
	for i, v := range vals {
		quoted[i] = `"` + v + `"`
	}
	return strings.Join(quoted, " ")
}

func nixStr(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `${`, `\${`)
	return `"` + s + `"`
}

func sortedKeys(m map[string]*JSONSchema) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func toSet(slice []string) map[string]bool {
	m := make(map[string]bool, len(slice))
	for _, v := range slice {
		m[v] = true
	}
	return m
}
