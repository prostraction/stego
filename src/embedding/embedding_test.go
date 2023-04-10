package embedding

import (
	"testing"
)

func TestBoolArrayToString(t *testing.T) {
	testWord := ""
	want := ""
	for i := -255; i < 513; i++ {
		testWord += string(uint8(i))
		want += string(uint8(i))
		boolArr := StringToBoolArray(testWord)
		msg := BoolArrayToString(boolArr)
		if msg != want {
			t.Fatalf(`boolArrayToString() = %q, want match for %q, (%d)`, msg, want, uint8(i))
		}
	}
}
