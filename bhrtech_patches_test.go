package goja

import (
	"reflect"
	"testing"
)

// Tests for BHRTech fork patches (see BHR_VERSIONING.md). They live in their
// own file so that merges from upstream do not conflict with them.

// bhrtechSetting is filled from JS object literals. With a lowercasing
// FieldNameMapper its JS-facing names are label, kind, enabled, limit and data.
// The json tags are ignored by that mapper.
type bhrtechSetting struct {
	owner string // unexported: must never be filled from JS

	Label   string `json:"label"`
	Kind    string // deliberately untagged: must map the same as tagged fields
	Enabled bool   `json:"enabled"`
	Limit   int32  `json:"maxLimit"` // tag deliberately differs from the mapped name "limit"
	Data    any    `json:"data"`
}

type bhrtechRegistry struct {
	added []*bhrtechSetting
}

func (r *bhrtechRegistry) Add(s *bhrtechSetting) *bhrtechSetting {
	r.added = append(r.added, s)
	return s
}

func newBHRTechTestRuntime() *Runtime {
	vm := New()
	vm.SetFieldNameMapper(UncapFieldNameMapper())
	return vm
}

// When a JS object is converted to a Go struct and a property is missing under
// the mapped field name, it is looked up under the Go field name instead.
func TestBHRTechExportToStructGoFieldNameFallback(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   bhrtechSetting
	}{
		{
			name:   "capitalised keys",
			script: `({Label: "a", Kind: "b", Enabled: true, Limit: 3})`,
			want:   bhrtechSetting{Label: "a", Kind: "b", Enabled: true, Limit: 3},
		},
		{
			name:   "mapped keys",
			script: `({label: "a", kind: "b", enabled: true, limit: 3})`,
			want:   bhrtechSetting{Label: "a", Kind: "b", Enabled: true, Limit: 3},
		},
		{
			name:   "mixed keys",
			script: `({label: "a", Kind: "b"})`,
			want:   bhrtechSetting{Label: "a", Kind: "b"},
		},
		{
			name:   "mapped name wins over Go field name",
			script: `({label: "a", Label: "b"})`,
			want:   bhrtechSetting{Label: "a"},
		},
		{
			name:   "object into interface field",
			script: `({Data: {count: 1}})`,
			want:   bhrtechSetting{Data: map[string]any{"count": int64(1)}},
		},
		{
			name:   "json tag name is ignored",
			script: `({maxLimit: 3})`,
			want:   bhrtechSetting{},
		},
		{
			name:   "unexported field is never filled",
			script: `({owner: "a", Owner: "b"})`,
			want:   bhrtechSetting{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vm := newBHRTechTestRuntime()
			v, err := vm.RunString(tc.script)
			if err != nil {
				t.Fatal(err)
			}
			var got bhrtechSetting
			if err := vm.ExportTo(v, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

// The fallback also applies when goja converts a JS object into a struct
// pointer argument of a Go method. Reading the struct from JS still uses only
// the mapped names.
func TestBHRTechStructPointerArgumentGoFieldNameFallback(t *testing.T) {
	vm := newBHRTechTestRuntime()
	reg := &bhrtechRegistry{}
	if err := vm.Set("registry", reg); err != nil {
		t.Fatal(err)
	}
	v, err := vm.RunString(`
	var s = registry.add({Label: "a", Enabled: true, Limit: 3});
	[s.label, s.enabled, s.limit, typeof s.Label].join("|");
	`)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.String(); got != "a|true|3|undefined" {
		t.Fatalf("unexpected JS view of the struct: %q", got)
	}
	if len(reg.added) != 1 {
		t.Fatalf("expected 1 added setting, got %d", len(reg.added))
	}
	want := bhrtechSetting{Label: "a", Enabled: true, Limit: 3}
	if got := *reg.added[0]; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

// Without a FieldNameMapper the mapped name is the Go field name, so the
// fallback never applies and lowercase keys do not fill capitalised fields.
func TestBHRTechExportToStructWithoutMapper(t *testing.T) {
	vm := New()
	v, err := vm.RunString(`({Label: "a", kind: "b"})`)
	if err != nil {
		t.Fatal(err)
	}
	var got bhrtechSetting
	if err := vm.ExportTo(v, &got); err != nil {
		t.Fatal(err)
	}
	want := bhrtechSetting{Label: "a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
