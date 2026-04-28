// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Limits applied to identity schemas before they reach the upstream
// jsonschema compiler. These bound the worst-case CPU and memory cost of
// compiling a customer-supplied schema, regardless of upstream behavior.
//
// The thresholds are intentionally conservative compared to realistic
// identity schemas (which rarely exceed 5-10 levels of nesting and 50
// properties). Operators with legitimate schemas above these limits should
// raise the bug rather than chase the cap.
const (
	// MaxSchemaBodyBytes caps the raw byte size of an identity schema fetched
	// from any source. Documents above this size cannot be cached or
	// compiled.
	MaxSchemaBodyBytes = 1 << 20 // 1 MiB

	// MaxSchemaNestingDepth caps the depth of nested objects/arrays inside
	// the parsed schema document. The upstream compiler has no depth cap;
	// Go stdlib's json.Decoder caps at 10000, but that is far above any
	// realistic identity schema.
	MaxSchemaNestingDepth = 32

	// MaxSchemaObjectKeys caps the number of keys in any single object in
	// the schema document, including `properties`, `patternProperties`,
	// `$defs`, and `definitions`. Each property compiles to a *Schema node.
	MaxSchemaObjectKeys = 1024

	// MaxSchemaArrayElements caps the length of any array in the schema
	// document, including `allOf`, `anyOf`, `oneOf`, and tuple `items`.
	MaxSchemaArrayElements = 128

	// MaxSchemaTotalNodes caps the total number of map+array nodes in the
	// document. Acts as a final backstop against schemas that stay within
	// per-level limits but compose them to produce a huge tree.
	MaxSchemaTotalNodes = 8192
)

// preValidateSchema walks a parsed identity schema document and rejects
// patterns that would let a customer-supplied schema crash kratos or pin
// pathological resources at compile or validate time.
//
// Specifically, this function rejects:
//
//   - A schema body whose decoded structure exceeds MaxSchemaNestingDepth,
//     MaxSchemaObjectKeys, MaxSchemaArrayElements, or MaxSchemaTotalNodes.
//   - Any cycle in the document's `$ref` chain graph: a sequence of `$ref`
//     nodes P₀ → P₁ → … → Pₙ → P₀ in which every step is a `$ref`. The
//     upstream compiler memoizes ref resolution to terminate compilation,
//     but the resulting `*Schema` graph contains the cycle, and at
//     validate time `s.Ref.validate(v)` recurses indefinitely and crashes
//     the process via fatal stack overflow (jsonschema/v3
//     schema.go:155-169 has no validate-time cycle guard, and stack
//     overflow is unrecoverable).
//   - A `pattern` or `patternProperties` key whose value is not a valid
//     regular expression — the upstream compiler uses regexp.MustCompile,
//     which panics, and the compile path has no panic recovery.
//
// Cycles whose intermediate nodes have non-`$ref` validation (e.g. a
// `$ref` to an ancestor that contains a `properties` map) are NOT
// rejected: each cycle iteration consumes at least one level of input, so
// recursion is bounded by JSON parser depth. Detecting the dangerous
// pure-ref subgraph is the goal.
//
// The function is generic over schema dialects and structural keywords:
// limits are applied to every map and array in the parsed tree, regardless
// of whether the surrounding keyword is `properties`, `default`, or
// anything else. This is intentional defense-in-depth.
func preValidateSchema(doc any) error {
	v := &preValidator{refs: map[string]string{}}
	if err := v.walk(doc, 0, ""); err != nil {
		return err
	}
	return v.detectRefCycles()
}

type preValidator struct {
	nodes int
	// refs maps each `$ref` location's JSON-pointer path to its target
	// JSON-pointer path. Targets that cannot be resolved as in-document
	// fragments (external URLs, malformed refs) are excluded — those go
	// through the loadRefURL scheme allowlist.
	refs map[string]string
}

func (p *preValidator) walk(v any, depth int, path string) error {
	if depth > MaxSchemaNestingDepth {
		return fmt.Errorf("identity schema rejected: nesting depth exceeds %d", MaxSchemaNestingDepth)
	}

	switch v := v.(type) {
	case map[string]any:
		p.nodes++
		if p.nodes > MaxSchemaTotalNodes {
			return fmt.Errorf("identity schema rejected: total node count exceeds %d", MaxSchemaTotalNodes)
		}
		if len(v) > MaxSchemaObjectKeys {
			return fmt.Errorf("identity schema rejected: object key count %d exceeds %d", len(v), MaxSchemaObjectKeys)
		}

		// Record `$ref` for cycle detection in detectRefCycles. Root
		// pointers (`#`, `#/`, empty) map to the empty path. Anything
		// without a `#/` prefix is external — out of scope here;
		// loadRefURL handles scheme enforcement.
		if ref, ok := v["$ref"].(string); ok {
			switch {
			case ref == "" || ref == "#" || ref == "#/":
				p.refs[path] = ""
			case strings.HasPrefix(ref, "#/"):
				p.refs[path] = strings.TrimPrefix(ref, "#")
			}
		}

		// Pre-compile `pattern` regexes so an invalid one returns a
		// kratos-side error instead of panicking deep in
		// regexp.MustCompile during the upstream compile.
		if pat, ok := v["pattern"].(string); ok {
			if _, err := regexp.Compile(pat); err != nil {
				return fmt.Errorf("identity schema rejected: invalid regex in pattern: %w", err)
			}
		}

		// patternProperties keys are themselves regexes.
		if patternProps, ok := v["patternProperties"].(map[string]any); ok {
			for raw := range patternProps {
				if _, err := regexp.Compile(raw); err != nil {
					return fmt.Errorf("identity schema rejected: invalid regex in patternProperties key %q: %w", raw, err)
				}
			}
		}

		for k, sub := range v {
			if err := p.walk(sub, depth+1, path+"/"+escapeJSONPointer(k)); err != nil {
				return err
			}
		}

	case []any:
		p.nodes++
		if p.nodes > MaxSchemaTotalNodes {
			return fmt.Errorf("identity schema rejected: total node count exceeds %d", MaxSchemaTotalNodes)
		}
		if len(v) > MaxSchemaArrayElements {
			return fmt.Errorf("identity schema rejected: array element count %d exceeds %d", len(v), MaxSchemaArrayElements)
		}
		for i, sub := range v {
			if err := p.walk(sub, depth+1, path+"/"+strconv.Itoa(i)); err != nil {
				return err
			}
		}
	}

	return nil
}

// detectRefCycles reports an error if the document's `$ref` chain graph
// contains a cycle. Walking from each `$ref` location, follow the target
// path. If the target is itself a `$ref` location, continue. If the chain
// revisits a location, the resulting `*Schema` graph has a cycle that
// crashes Validate via stack overflow.
//
// The chain ends as soon as it reaches a node that is not itself a `$ref`
// — that node has its own validation logic (`properties`, `type`, etc.)
// which consumes input on each cycle iteration, so the recursion is
// bounded. Only pure `$ref` chains form unbounded loops.
func (p *preValidator) detectRefCycles() error {
	for start := range p.refs {
		visited := map[string]struct{}{}
		cur := start
		var chain []string
		for {
			if _, ok := visited[cur]; ok {
				idx := slices.Index(chain, cur)
				cycle := append(slices.Clone(chain[idx:]), cur)
				return fmt.Errorf("identity schema rejected: self-referential $ref cycle: %s",
					formatRefCycle(cycle))
			}
			visited[cur] = struct{}{}
			chain = append(chain, cur)
			next, ok := p.refs[cur]
			if !ok {
				break
			}
			cur = next
		}
	}
	return nil
}

// escapeJSONPointer encodes a property name as a JSON-pointer reference
// token (RFC 6901): `~` → `~0`, `/` → `~1`. The order matters — `~` must be
// escaped first so a literal `/` does not collide with the escape produced
// for `~`.
func escapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}

// formatRefCycle renders a path slice as a human-readable arrow chain,
// rendering the empty (root) path as "#".
func formatRefCycle(paths []string) string {
	parts := make([]string, len(paths))
	for i, p := range paths {
		if p == "" {
			parts[i] = "#"
		} else {
			parts[i] = "#" + p
		}
	}
	return strings.Join(parts, " → ")
}
