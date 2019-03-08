// Copyright 2019 The SwiftShader Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package testlist provides utilities for handling test lists.
package testlist

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"../cause"
)

// API is an enumerator of graphics APIs.
type API string

// Graphics APIs.
const (
	EGL    = API("egl")
	GLES2  = API("gles2")
	GLES3  = API("gles3")
	Vulkan = API("vulkan")
)

// Group is a list of tests to be run for a single API.
type Group struct {
	Name  string
	File  string
	API   API
	Tests []string
}

// Lists is the full list of tests to be run.
type Lists []Group

// Hash returns a SHA1 hash of the set of tests.
func (l Lists) Hash() string {
	h := sha1.New()
	if err := gob.NewEncoder(h).Encode(l); err != nil {
		panic(cause.Wrap(err, "Could not encode testlist to produce hash"))
	}
	var hash [20]byte
	copy(hash[:], h.Sum(nil))
	return hex.EncodeToString(hash[:])
}

// Load loads the test list json file and returns the full set of tests.
func Load(root, jsonPath string) (Lists, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, cause.Wrap(err, "Couldn't get absolute path of '%s'", root)
	}

	jsonPath, err = filepath.Abs(jsonPath)
	if err != nil {
		return nil, cause.Wrap(err, "Couldn't get absolute path of '%s'", jsonPath)
	}

	i, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return nil, cause.Wrap(err, "Couldn't read test list from '%s'", jsonPath)
	}

	var jsonGroups []struct {
		Name     string
		API      string
		TestFile string `json:"tests"`
	}
	if err := json.NewDecoder(bytes.NewReader(i)).Decode(&jsonGroups); err != nil {
		return nil, cause.Wrap(err, "Couldn't parse '%s'", jsonPath)
	}

	dir := filepath.Dir(jsonPath)

	out := make(Lists, len(jsonGroups))
	for i, jsonGroup := range jsonGroups {
		path := filepath.Join(dir, jsonGroup.TestFile)
		tests, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, cause.Wrap(err, "Couldn't read '%s'", tests)
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil, cause.Wrap(err, "Couldn't get relative path for '%s'", path)
		}
		group := Group{
			Name: jsonGroup.Name,
			File: relPath,
			API:  API(jsonGroup.API),
		}
		for _, line := range strings.Split(string(tests), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				group.Tests = append(group.Tests, line)
			}
		}
		sort.Strings(group.Tests)
		out[i] = group
	}

	return out, nil
}
