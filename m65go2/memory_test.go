package m65go2

import (
	"testing"
)

func TestSamePage(t *testing.T) {
	for a := uint16(0x0000); ; a += 0x0100 {
		for b := uint16(0x0000); ; b += 0x0100 {
			if a>>8 == b>>8 && !SamePage(a, b) {
				t.Errorf("Bad result for SamePage(%#04x, %#04x)\n", a, b)
			} else if a>>8 != b>>8 && SamePage(a, b) {
				t.Errorf("Bad result for SamePage(%#04x, %#04x)\n", a, b)
			}

			if b == 0xff00 {
				break
			}
		}

		if a == 0xff00 {
			break
		}
	}
}
