package project

import "testing"

func Test_Root(t *testing.T) {
	var got string
	got = Root()
	t.Logf("Root: %v", got)

	got = Root("cmd", "internal")
	t.Logf("Root: %v", got)

}
