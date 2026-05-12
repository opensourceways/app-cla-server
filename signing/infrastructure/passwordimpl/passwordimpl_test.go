package passwordimpl

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()
	if cfg.MinLength != 8 {
		t.Errorf("expected 8, got %d", cfg.MinLength)
	}
	if cfg.MaxLength != 20 {
		t.Errorf("expected 20, got %d", cfg.MaxLength)
	}
	if cfg.MaxNumOfConsecutiveChars != 2 {
		t.Errorf("expected 2, got %d", cfg.MaxNumOfConsecutiveChars)
	}
	if cfg.MinNumOfKindOfPasswordChar != 3 {
		t.Errorf("expected 3, got %d", cfg.MinNumOfKindOfPasswordChar)
	}
}

func TestConfigSetDefaultKeepsCustom(t *testing.T) {
	cfg := &Config{
		MinLength:                  10,
		MaxLength:                  30,
		MaxNumOfConsecutiveChars:   3,
		MinNumOfKindOfPasswordChar: 4,
	}
	cfg.SetDefault()
	if cfg.MinLength != 10 {
		t.Errorf("expected 10, got %d", cfg.MinLength)
	}
	if cfg.MinNumOfKindOfPasswordChar != 4 {
		t.Errorf("expected 4, got %d", cfg.MinNumOfKindOfPasswordChar)
	}
}

func TestHasMultiChars(t *testing.T) {
	impl := &passwordImpl{
		cfg: Config{MinNumOfKindOfPasswordChar: 3},
	}

	if !impl.hasMultiChars([]byte("Aa1!")) {
		t.Error("expected true for mixed chars")
	}
	if !impl.hasMultiChars([]byte("Aa1aA1")) {
		t.Error("expected true for mixed chars")
	}
	if impl.hasMultiChars([]byte("aaaaaa")) {
		t.Error("expected false for all lowercase")
	}
	if impl.hasMultiChars([]byte("ABCDEF")) {
		t.Error("expected false for all uppercase")
	}
	if impl.hasMultiChars([]byte("123456")) {
		t.Error("expected false for all digits")
	}
}

func TestMarkCharType(t *testing.T) {
	impl := &passwordImpl{}

	tests := []struct {
		c    byte
		idx  int
		name string
	}{
		{'a', 0, "lowercase"},
		{'z', 0, "lowercase"},
		{'A', 1, "uppercase"},
		{'Z', 1, "uppercase"},
		{'0', 2, "digit"},
		{'9', 2, "digit"},
		{'!', 3, "special"},
		{'@', 3, "special"},
	}

	for _, tt := range tests {
		part := make([]bool, 4)
		impl.markCharType(tt.c, part)
		if !part[tt.idx] {
			t.Errorf("expected part[%d] to be true for %s char '%c'", tt.idx, tt.name, tt.c)
		}
	}
}

func TestCountCharTypes(t *testing.T) {
	impl := &passwordImpl{}

	if n := impl.countCharTypes([]bool{true, false, false, false}); n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
	if n := impl.countCharTypes([]bool{true, true, true, false}); n != 3 {
		t.Errorf("expected 3, got %d", n)
	}
	if n := impl.countCharTypes([]bool{true, true, true, true}); n != 4 {
		t.Errorf("expected 4, got %d", n)
	}
	if n := impl.countCharTypes([]bool{false, false, false, false}); n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestHasConsecutive(t *testing.T) {
	impl := &passwordImpl{
		cfg: Config{MaxNumOfConsecutiveChars: 2},
	}

	// 3 consecutive a's > threshold 2, should be detected
	if !impl.hasConsecutive([]byte("aaa")) {
		t.Error("expected true for 'aaa' (3 consecutive)")
	}
	// 2 consecutive a's is not > 2, should not be detected
	if impl.hasConsecutive([]byte("aa")) {
		t.Error("expected false for 'aa' (2 consecutive, not > threshold)")
	}
	// 3 consecutive a's at start should still be detected
	if !impl.hasConsecutive([]byte("aaabbb")) {
		t.Error("expected true for 'aaabbb' (3 consecutive a's)")
	}
}

func TestHasConsecutiveNoConsecutive(t *testing.T) {
	impl := &passwordImpl{
		cfg: Config{MaxNumOfConsecutiveChars: 2},
	}
	if impl.hasConsecutive([]byte("ababab")) {
		t.Error("expected false for no consecutive")
	}
	if impl.hasConsecutive([]byte("a")) {
		t.Error("expected false for single char")
	}
	if impl.hasConsecutive([]byte("")) {
		t.Error("expected false for empty")
	}
}

func TestGenChar(t *testing.T) {
	impl := &passwordImpl{}
	v := byte(100)

	for i := 0; i < 3; i++ {
		c := impl.genChar(i, v)
		switch i % 3 {
		case 0:
			if c < 'a' || c > 'z' {
				t.Errorf("genChar(0) expected lowercase, got '%c'", c)
			}
		case 1:
			if c < 'A' || c > 'Z' {
				t.Errorf("genChar(1) expected uppercase, got '%c'", c)
			}
		case 2:
			if c < '0' || c > '9' {
				t.Errorf("genChar(2) expected digit, got '%c'", c)
			}
		}
	}
}

func TestIsValid(t *testing.T) {
	impl := &passwordImpl{
		cfg: Config{
			MinLength:                  8,
			MaxLength:                  20,
			MaxNumOfConsecutiveChars:   2,
			MinNumOfKindOfPasswordChar: 3,
		},
	}

	validPw, _ := dp.NewPassword([]byte("Abc123!@"))
	if !impl.IsValid(validPw) {
		t.Error("expected valid password")
	}

	// Too short
	shortPw, _ := dp.NewPassword([]byte("Ab1!"))
	if impl.IsValid(shortPw) {
		t.Error("expected invalid for too short")
	}

	// Too long
	longBytes := make([]byte, 21)
	for i := range longBytes {
		longBytes[i] = 'a'
	}
	longPw, _ := dp.NewPassword(longBytes)
	if impl.IsValid(longPw) {
		t.Error("expected invalid for too long")
	}

	// Invalid characters (spaces)
	spacePw, _ := dp.NewPassword([]byte("Abc 123!"))
	if impl.IsValid(spacePw) {
		t.Error("expected invalid for space character")
	}
}

func TestIsValidNotEnoughCharTypes(t *testing.T) {
	impl := &passwordImpl{
		cfg: Config{
			MinLength:                  4,
			MaxLength:                  20,
			MaxNumOfConsecutiveChars:   2,
			MinNumOfKindOfPasswordChar: 3,
		},
	}
	// Only lowercase and uppercase (2 types)
	pw, _ := dp.NewPassword([]byte("AbcdEfgh"))
	if impl.IsValid(pw) {
		t.Error("expected invalid for only 2 char types")
	}
}

func TestGoodFormat(t *testing.T) {
	impl := &passwordImpl{
		cfg: Config{
			MaxNumOfConsecutiveChars:   2,
			MinNumOfKindOfPasswordChar: 3,
		},
	}

	if !impl.goodFormat([]byte("Abc123!@")) {
		t.Error("expected good format")
	}
	if impl.goodFormat([]byte("aaa123!")) {
		t.Error("expected bad format due to consecutive chars")
	}
}

func TestNewPasswordImpl(t *testing.T) {
	cfg := &Config{MinLength: 8, MaxLength: 20}
	impl := NewPasswordImpl(cfg)
	if impl == nil {
		t.Fatal("expected non-nil")
	}
	if impl.cfg.MinLength != 8 {
		t.Errorf("expected 8, got %d", impl.cfg.MinLength)
	}
}

func TestPasswordNew(t *testing.T) {
	impl := &passwordImpl{
		cfg: Config{
			MinLength:                  8,
			MaxLength:                  20,
			MaxNumOfConsecutiveChars:   2,
			MinNumOfKindOfPasswordChar: 3,
		},
	}

	pw, err := impl.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := pw.Password()
	if len(b) != 8 {
		t.Errorf("expected length 8, got %d", len(b))
	}
	if !impl.IsValid(pw) {
		t.Error("generated password should be valid")
	}
}
