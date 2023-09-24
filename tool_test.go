package poh

import "testing"

func TestBytesToIntHashXX(t *testing.T) {
	h := BytesToIntHashXX([]byte("123123123123213213123213123123213"))

	if h != 6871299203570162080 {
		t.Fatal(h)
	}

	h = StringToIntHashXX("1231231231232132131232131231232136")

	if h != -356980692338892050 {
		t.Fatal(h)
	}
}
