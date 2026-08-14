package models

import (
	"testing"

	"github.com/weibaohui/k8m/pkg/constants"
)

// TestBuiltinLuaScriptsHaveUniqueNames 锁定内置脚本源数据的 Name 唯一性。
// 回归保护：历史上曾出现两条 Name 完全相同但 ScriptCode 不同的条目，导致插件安装时
// SQLite 报 "UNIQUE constraint failed: inspection_lua_scripts.name"。
func TestBuiltinLuaScriptsHaveUniqueNames(t *testing.T) {
	if len(BuiltinLuaScripts) == 0 {
		t.Fatal("BuiltinLuaScripts is empty")
	}
	seen := make(map[string]struct{}, len(BuiltinLuaScripts))
	for i, s := range BuiltinLuaScripts {
		if _, dup := seen[s.Name]; dup {
			t.Errorf("duplicate builtin script Name at index %d: %q (script_code=%q)", i, s.Name, s.ScriptCode)
		}
		seen[s.Name] = struct{}{}
		if s.ScriptType != constants.LuaScriptTypeBuiltin {
			t.Errorf("builtin script at index %d (%q) has ScriptType=%q, want %q", i, s.Name, s.ScriptType, constants.LuaScriptTypeBuiltin)
		}
	}
}

// TestBuiltinLuaScriptsHaveUniqueScriptCodes 锁定内置脚本源数据的 ScriptCode 唯一性。
func TestBuiltinLuaScriptsHaveUniqueScriptCodes(t *testing.T) {
	seen := make(map[string]struct{}, len(BuiltinLuaScripts))
	for i, s := range BuiltinLuaScripts {
		if s.ScriptCode == "" {
			t.Errorf("builtin script at index %d (%q) has empty ScriptCode", i, s.Name)
			continue
		}
		if _, dup := seen[s.ScriptCode]; dup {
			t.Errorf("duplicate builtin script ScriptCode at index %d: %q (name=%q)", i, s.ScriptCode, s.Name)
		}
		seen[s.ScriptCode] = struct{}{}
	}
}

// TestDedupeBuiltinLuaScripts 验证内置脚本去重逻辑在各种重名场景下的行为。
// 防御回归：当源数据中偶然出现重名条目时，插件安装/重载不应该因为 SQLite UNIQUE 约束失败。
func TestDedupeBuiltinLuaScripts(t *testing.T) {
	tests := []struct {
		name      string
		input     []InspectionLuaScript
		wantNames []string
		wantLen   int
	}{
		{
			name: "no duplicates",
			input: []InspectionLuaScript{
				{Name: "a", ScriptCode: "code-a", ScriptType: constants.LuaScriptTypeBuiltin},
				{Name: "b", ScriptCode: "code-b", ScriptType: constants.LuaScriptTypeBuiltin},
			},
			wantNames: []string{"a", "b"},
			wantLen:   2,
		},
		{
			name: "single duplicate keeps first occurrence",
			input: []InspectionLuaScript{
				{Name: "a", ScriptCode: "code-a-1", ScriptType: constants.LuaScriptTypeBuiltin},
				{Name: "b", ScriptCode: "code-b", ScriptType: constants.LuaScriptTypeBuiltin},
				{Name: "a", ScriptCode: "code-a-2", ScriptType: constants.LuaScriptTypeBuiltin},
			},
			wantNames: []string{"a", "b"},
			wantLen:   2,
		},
		{
			name: "multiple duplicates preserve order of first appearance",
			input: []InspectionLuaScript{
				{Name: "a", ScriptCode: "code-a-1", ScriptType: constants.LuaScriptTypeBuiltin},
				{Name: "b", ScriptCode: "code-b", ScriptType: constants.LuaScriptTypeBuiltin},
				{Name: "a", ScriptCode: "code-a-2", ScriptType: constants.LuaScriptTypeBuiltin},
				{Name: "b", ScriptCode: "code-b-2", ScriptType: constants.LuaScriptTypeBuiltin},
				{Name: "c", ScriptCode: "code-c", ScriptType: constants.LuaScriptTypeBuiltin},
			},
			wantNames: []string{"a", "b", "c"},
			wantLen:   3,
		},
		{
			name: "all entries duplicate the same name",
			input: []InspectionLuaScript{
				{Name: "same", ScriptCode: "code-1", ScriptType: constants.LuaScriptTypeBuiltin},
				{Name: "same", ScriptCode: "code-2", ScriptType: constants.LuaScriptTypeBuiltin},
				{Name: "same", ScriptCode: "code-3", ScriptType: constants.LuaScriptTypeBuiltin},
			},
			wantNames: []string{"same"},
			wantLen:   1,
		},
		{
			name:      "empty input returns empty slice",
			input:     []InspectionLuaScript{},
			wantNames: []string{},
			wantLen:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DedupeBuiltinLuaScripts(tc.input)
			if len(got) != tc.wantLen {
				t.Fatalf("len(got) = %d, want %d", len(got), tc.wantLen)
			}
			// Ensure the input is not mutated.
			if len(tc.input) > 0 && &tc.input[0] == &got[0] && len(tc.input) != len(got) {
				t.Fatalf("input was mutated: input len=%d got len=%d", len(tc.input), len(got))
			}
			// Validate order and names.
			for i, want := range tc.wantNames {
				if got[i].Name != want {
					t.Errorf("got[%d].Name = %q, want %q", i, got[i].Name, want)
				}
			}
		})
	}
}