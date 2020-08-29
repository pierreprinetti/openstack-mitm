package proxy

import (
	"testing"
)

func TestSplitAlias(t *testing.T) {
	for _, tc := range [...]struct {
		input      string
		alias      string
		targetPath string
	}{
		{
			"/alias/target-path",
			"alias",
			"/target-path",
		},
		{
			"/alias/target-path/",
			"alias",
			"/target-path/",
		},
		{
			"/alias/target/path",
			"alias",
			"/target/path",
		},
		{
			"/alias/tar get/path",
			"alias",
			"/tar get/path",
		},
		{
			"/alias",
			"alias",
			"/",
		},
		{
			"/",
			"",
			"/",
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			alias, targetPath := splitAlias(tc.input)

			if alias != tc.alias {
				t.Errorf("expected alias %q, found %q", tc.alias, alias)
			}

			if targetPath != tc.targetPath {
				t.Errorf("expected targetPath %q, found %q", tc.targetPath, targetPath)
			}
		})
	}
}

// func TestNew(t *testing.T) {
// 	u, err := url.Parse("https://kaizen.massopen.cloud:13000/")
// 	if err != nil {
// 		panic(err)
// 	}
// 	targets := NewAddressBook("localhost:2443")
// 	targets.Set("auth", *u)

// 	input,err:= url.Parse("")
// }
