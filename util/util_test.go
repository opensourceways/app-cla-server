package util

import "testing"

func TestMaskEmail(t *testing.T) {
	// 覆盖本地部分超4位、恰为4位、不足4位、多字节字符及各类非法输入
	cases := []struct {
		email string
		want  string
	}{
		{"abcdef@example.com", "abcd***@example.com"},   // 本地部分超过4位，仅展示前4位
		{"abcd@example.com", "abcd***@example.com"},     // 本地部分恰为4位，原样保留
		{"ab@example.com", "ab***@example.com"},         // 本地部分不足4位，原样保留
		{"你好世界你好@example.com", "你好世界***@example.com"}, // 本地部分含多字节字符，按rune切分前4位
		{"invalidemail", "invalidemail"},                // 缺少@分隔符，原样返回
		{"", ""},               // 空字符串，原样返回
		{"@example.com", "@example.com"}, // 本地部分为空，原样返回
		{"test@", "test@"},     // 后缀为空，原样返回
	}

	for _, c := range cases {
		if got := MaskEmail(c.email); got != c.want {
			t.Errorf("MaskEmail(%q) = %q, want %q", c.email, got, c.want)
		}
	}
}
