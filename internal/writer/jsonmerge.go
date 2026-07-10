package writer

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MergeJSON 把 fragment（JSON object）深合并进 existing（JSON object），返回
// 格式化后的合并结果。合并语义面向"往用户配置里注册路径"场景：
//
//   - object：递归合并
//   - array：并集（保留 existing 顺序，fragment 中缺失的元素按序追加）
//   - scalar：existing 优先（不覆盖用户已有取值）
//
// existing 为空时直接格式化 fragment。existing 不是合法 JSON（如 JSONC 注释）
// 时返回 error，调用方应跳过写入而不是破坏用户文件。
func MergeJSON(existing, fragment []byte) ([]byte, error) {
	var frag map[string]any
	if err := json.Unmarshal(fragment, &frag); err != nil {
		return nil, fmt.Errorf("parse fragment: %w", err)
	}

	merged := frag
	if len(bytes.TrimSpace(existing)) > 0 {
		var base map[string]any
		if err := json.Unmarshal(existing, &base); err != nil {
			return nil, fmt.Errorf("parse existing: %w", err)
		}
		merged = mergeValue(base, frag).(map[string]any)
	}

	out, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// mergeValue 按 MergeJSON 语义合并两个已解码的 JSON 值
func mergeValue(base, frag any) any {
	baseMap, bOK := base.(map[string]any)
	fragMap, fOK := frag.(map[string]any)
	if bOK && fOK {
		for k, fv := range fragMap {
			if bv, ok := baseMap[k]; ok {
				baseMap[k] = mergeValue(bv, fv)
			} else {
				baseMap[k] = fv
			}
		}
		return baseMap
	}
	baseArr, bOK := base.([]any)
	fragArr, fOK := frag.([]any)
	if bOK && fOK {
		for _, fv := range fragArr {
			if !containsJSONValue(baseArr, fv) {
				baseArr = append(baseArr, fv)
			}
		}
		return baseArr
	}
	// scalar 或类型不一致：existing 优先
	return base
}

// containsJSONValue 判断数组是否已含等价元素（JSON 序列化后比较）
func containsJSONValue(arr []any, v any) bool {
	vb, _ := json.Marshal(v)
	for _, e := range arr {
		eb, _ := json.Marshal(e)
		if bytes.Equal(eb, vb) {
			return true
		}
	}
	return false
}
