// Copyright 2014 The hesiod-go authors. All rights reserved.
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

package hesiod

import (
	"errors"
	"strings"
	"testing"
)

func TestPrepareDNSName(t *testing.T) {
	hesiod := NewHesiodWithConfig(AthenaConfig)
	name, err := hesiod.prepareDNSName("zephyr", "sloc")
	if err != nil {
		t.FailNow()
	}
	t.Logf("Resolved <zephyr, sloc> to <%s>", name)
	if name != "zephyr.sloc.ns.athena.mit.edu" {
		t.Fail()
	}
	name, err = hesiod.prepareDNSName("zephyr@sipb.mit.edu", "sloc")
	if err != nil {
		t.FailNow()
	}
	t.Logf("Resolved <zephyr@sipb.mit.edu, sloc> to <%s>", name)
	if name != "zephyr.sloc.ns.sipb.mit.edu" {
		t.Fail()
	}
}

type mockResolver map[string][]string

func (mock mockResolver) LookupTXT(name string) ([]string, error) {
	value, ok := mock[name]
	if !ok {
		return nil, errors.New("Mock resolver cannot find specified value")
	}
	return value, nil
}

func TestResolve(t *testing.T) {
	hesiod := NewHesiodWithConfig(AthenaConfig)
	hesiod.lookup = mockResolver(map[string][]string {
		"zephyr.sloc.ns.athena.mit.edu": {
			"ARILINN.MIT.EDU",
			"NESKAYA.MIT.EDU",
		},
	})
	results, err := hesiod.Resolve("zephyr", "sloc")
	if err != nil {
		t.FailNow()
	}
	t.Logf("Found Athena Zephyr servers: %s", results)
	pristine := []string{"ARILINN.MIT.EDU", "NESKAYA.MIT.EDU"}
	if strings.Join(results, " ") != strings.Join(pristine, " ") {
		t.Fail()
	}

	hesiod.lookup = mockResolver(map[string][]string {
	})
	results, err = hesiod.Resolve("zephyr", "sloc")
	if err == nil {
		t.FailNow()
	}
}

func TestGetServiceByName(t *testing.T) {
	hesiod := NewHesiodWithConfig(AthenaConfig)
	hesiod.lookup = mockResolver(map[string][]string {
		"zephyr.service.ns.athena.mit.edu": {
			"zephyr udp 2102",
		},
	})
	result, err := hesiod.GetServiceByName("zephyr", "udp")
	if err != nil {
		t.FailNow()
	}
	t.Logf("Resolved service <zephyr,udp> to <%s,%s,%d>", result.ServiceName, result.Protocol, result.Port)
	if result.ServiceName != "zephyr" || result.Protocol != "udp" || result.Port != 2102 {
		t.Fail()
	}
	hesiod.lookup = mockResolver(map[string][]string {
		"zephyr.service.ns.athena.mit.edu": {
			"fake tcp 12",
		},
	})
	result, err = hesiod.GetServiceByName("zephyr", "udp")
	if err != nil {
		t.FailNow()
	}
	t.Logf("Resolved fake service <zephyr,udp> to <%s,%s,%d>", result.ServiceName, result.Protocol, result.Port)
	if result.ServiceName == "zephyr" || result.Protocol == "udp" || result.Port == 2102 {
		t.Fail()
	}


}
