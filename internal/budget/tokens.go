package budget

import "unicode"

// EstimateTokens 粗略估算 LLM tokens 数。
//
// 经验近似：
//   - ASCII 字符（含标点空白）约 4 chars / token（OpenAI BPE 经验值）
//   - 中文字符（CJK Unified Ideographs）约 1.5 chars / token（BPE 倾向把
//     单个汉字 split 为 1-2 tokens）
//   - 其他 Unicode 字符按 ascii 处理（保守）
//
// 不是精确的 tiktoken / Claude tokenizer 实现，仅用于 budget 提示，
// 误差范围约 ±30%。需要精确估算可在 cli 端引入 tiktoken-go 等库。
func EstimateTokens(s string) int {
	asciiCount := 0
	chineseCount := 0
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			chineseCount++
		} else {
			asciiCount++
		}
	}
	return asciiCount/4 + (chineseCount*2)/3
}
