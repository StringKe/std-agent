package budget

import (
	"strings"
	"testing"
)

func TestEstimateTokensASCII(t *testing.T) {
	s := "hello world test" // 16 chars
	if got := EstimateTokens(s); got != 4 {
		t.Errorf("got %d, want 4 (16/4)", got)
	}
}

func TestEstimateTokensChinese(t *testing.T) {
	s := "你好世界" // 4 汉字
	if got := EstimateTokens(s); got != 2 {
		t.Errorf("got %d, want 2 (4*2/3)", got)
	}
}

func TestEstimateTokensMixed(t *testing.T) {
	s := "hello 世界" // 6 ascii (含空格) + 2 中文
	got := EstimateTokens(s)
	// 6/4 + 2*2/3 = 1 + 1 = 2
	if got != 2 {
		t.Errorf("got %d, want 2", got)
	}
}

func TestEstimateTokensEmpty(t *testing.T) {
	if got := EstimateTokens(""); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestEstimateTokensLargeASCII(t *testing.T) {
	s := strings.Repeat("a", 10000)
	if got := EstimateTokens(s); got != 2500 {
		t.Errorf("got %d, want 2500", got)
	}
}

func TestEstimateTokensLargeChinese(t *testing.T) {
	s := strings.Repeat("中", 1000)
	if got := EstimateTokens(s); got != 666 {
		t.Errorf("got %d, want 666 (1000*2/3)", got)
	}
}

func TestEstimateTokensPunctuation(t *testing.T) {
	// punctuation/whitespace 都按 ascii 计
	s := "  ,;.!?  "
	got := EstimateTokens(s)
	if got < 1 || got > 3 {
		t.Errorf("got %d, want 1..3", got)
	}
}
