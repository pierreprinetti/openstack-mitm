// Copyright 2020 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
