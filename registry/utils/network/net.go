// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package network

import (
	"net"
	"time"
)

// Network-related utility functions

const (
	waitInterval = time.Duration(100) * time.Millisecond
	waitTimeout  = time.Duration(2) * time.Minute
)

// GetPrivateIP returns a private IP address, or panics if no IP is available.
func GetPrivateIP() net.IP {
	addr := getPrivateIPIfAvailable()
	if addr.IsUnspecified() {
		panic("No private IP address is available")
	}
	return addr
}

// GetPrivateIPv4 returns a private IPv4 address, or panics if no IP is available.
func GetPrivateIPv4() net.IP {
	addr := getPrivateIPv4IfAvailable()
	if addr.IsUnspecified() {
		panic("No private IP address is available")
	}
	return addr
}

// WaitForPrivateNetwork blocks until a private IP address is available, or a timeout is reached.
// Returns 'true' if a private IP is available before timeout is reached, and 'false' otherwise.
func WaitForPrivateNetwork() bool {
	deadline := time.Now().Add(waitTimeout)
	for {
		addr := getPrivateIPIfAvailable()
		if !addr.IsUnspecified() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(waitInterval)
	}
}

// WaitForPrivateNetworkIPv4 blocks until a private IP address is available, or a timeout is reached.
// Returns 'true' if a private IP is available before timeout is reached, and 'false' otherwise.
func WaitForPrivateNetworkIPv4() bool {
	deadline := time.Now().Add(waitTimeout)
	for {
		addr := getPrivateIPv4IfAvailable()
		if !addr.IsUnspecified() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(waitInterval)
	}
}

// Returns a private IP address, or unspecified IP (0.0.0.0) if no IP is available
func getPrivateIPIfAvailable() net.IP {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}
		if !ip.IsLoopback() {
			return ip
		}
	}
	return net.IPv4zero
}

// Returns a private IPv4 address, or unspecified IP (0.0.0.0) if no IP is available
func getPrivateIPv4IfAvailable() net.IP {
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	return net.IPv4zero
}
