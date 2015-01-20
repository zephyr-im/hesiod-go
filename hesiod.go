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

// Package hesiod provides a Go hesiod protocol implementation.
//
// Hesiod is a DNS-based directory service developed at MIT for
// Project Athena. The Go implementation provides a significantly
// simplified API than the C library.
//
// More information about Hesiod is available from the USENIX paper
// available from
//   ftp://athena-dist.mit.edu/pub/ATHENA/usenix/hesiod.PS
package hesiod

import (
	"fmt"
	"net"
	"strings"
)

// A Config contains the Hesiod name server prefix as well as the
// realm suffix. See AthenaConfig for an example of a typical
// configuration.
type Config struct {
	Nameserver string
	Realm string
}

// AthenaConfig is the default configuration for Hesiod, for use at
// MIT.
var AthenaConfig = &Config{
	Nameserver: ".ns",
	Realm: ".athena.mit.edu",
}

// The Hesiod service itself.
type Hesiod struct {
	// The Hesiod documentation refers to nameserver as "lhs"
	nameserver string
	// and realm as "rhs"
	realm string
	// For testing only
	lookup lookupInterface
}

// DefaultHesiodConfigFile is a path to the systemd-wide hesiod.conf
// file. On UNIX-like systems, this should be /etc/hesiod.conf.
const DefaultHesiodConfigFile = "/etc/hesiod.conf"

// ServiceEntry contains the information similar to getservbyname(3)
// method, and is used by GetServiceByName().
type ServiceEntry struct {
	ServiceName string
	Port int
	Protocol string
}

// NewHesiod initializes a Hesiod resolver with the machine's local
// Hesiod configuration.
func NewHesiod() *Hesiod {
	// Hesiod, being created at MIT, officially defaults to MIT's
	// ATHENA realm unless a configuration file is specified.
	//
	// TODO(achernya): Implement parsing hesiod.conf and make that
	// the default.
	return NewHesiodWithConfig(AthenaConfig)
}

// NewHesiodWithConfig allows the application to override the
// machine's local Hesiod configuration with the specified
// configuration. In the C Hesiod library, this was controlled through
// the HESIOD_CONFIG environment variable, which is intentionally
// unsupported in hesiod-go.
func NewHesiodWithConfig(config *Config) *Hesiod {
	return &Hesiod{config.Nameserver, config.Realm, lookupInterfaceFunc(net.LookupTXT)}
}

func (hesiod *Hesiod) prepareDNSName(question string, queryType string) (dns string, err error) {
	splitQuestion := strings.SplitN(question, "@", 2)
	realm := hesiod.realm
	if len(splitQuestion) == 2 {
		// This hesiod name contains a redirect to another
		// realm. If it's qualified (i.e., contains '.', then
		// everything after the '@' becomes the realm. If not,
		// then we need to look up that name under the
		// 'rhs-extension' query type to find the realm
		realm = splitQuestion[1]
		if !strings.ContainsRune(realm, '.') {
			var candidates []string
			candidates, err = hesiod.Resolve(realm, "rhs-extension")
			if err != nil {
				return
			}
			realm = candidates[0]
		}
		realm = "." + realm
	}
	dns = splitQuestion[0] + "." + queryType + hesiod.nameserver + realm
	return
}

// Resolve performs the Hesiod resolution of the given query and its
// type. For example, try Resolve("hesiod", "sloc") or
// Resolve("achernya", "passwd") with the Athena configuration.
func (hesiod *Hesiod) Resolve(question string, queryType string) (txt []string, err error) {
	var dns string
	dns, err = hesiod.prepareDNSName(question, queryType)
	if err != nil {
		return
	}
	txt, err = hesiod.lookup.LookupTXT(dns)
	return
}

// GetServiceByName queries Hesiod for the service information
// (protocol, port).
func (hesiod *Hesiod) GetServiceByName(service string, proto string) (result ServiceEntry, err error) {
	var candidates []string
	candidates, err = hesiod.Resolve(service, "service")
	for _, candidate := range candidates {
		var resolvedService, protocol string
		var port int
		_, err = fmt.Sscanf(candidate, "%s %s %d", &resolvedService, &protocol, &port)
		if err != nil {
			return
		}
		if protocol != proto {
			continue
		}
		result.ServiceName = resolvedService
		result.Port = port
		result.Protocol = protocol
		return
	}
	return
}
