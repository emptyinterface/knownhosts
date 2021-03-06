// This package was originally imported from github.com/coreos/fleet/ssh.
// As such it is a derivative work under the Apache 2.0 license under which
// it was originally open-sourced.  All original licensing documents have been
// included without modification.  This package redistributed under:
//
// Copyright 2016 Empty Interface LLC
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

// Copyright 2014 CoreOS, Inc.
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

package knownhosts

import (
	"bytes"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"golang.org/x/crypto/ssh"
)

const (
	hostLine           = "192.0.2.10:2222 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC3y6omlFuiBQfV2lqwqt3EuQHXLxvghhdfyZ840je6pRNnidgfCTmzNgIjmqdfkCwIthh+fhArkFPWIT6dRwim4hhtbpum7AzAay1h6mmLsmJVJQ/nK+zLwQ4JHs6+Tfj6F3iXJyrZR9JMTeLLs0mEd+VNHbX3LxIh7nXk5IM0G5LP2nnIYG96Luu4WunJzFsDVFLgxMl66T9VBYeAIbfUeCoCDYMmJK7kTleLD1XfL2KdoHkh0t9fkJVA5XJUZJPh3PJw+mT7eP3meAMc8EzyCGcRm+5GQzAe2/M4dNaZ5iqF7YIO7HJpA8UyAE+Dgd9WqhoBX/6ItdcuDXVAy63v\n"
	addrInHostLine     = "192.0.2.10:2222"
	hostFile           = "testdata/known_hosts"
	wrongAuthorizedKey = "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAzJjWHWVDum5WukrlWTYPtPN/Ny8BTXzhHFf89vejOQukQNMPcoohjSOBkrFZXQMLQ0s/RqpTKly1omdo8TgfUE5f7rgegwPhzleuxw/Q/XJJJiiCi7KHSQv9Vs+fNlMr14VsF8JStpKei5jD/moM1Pk/q5asYtY9I4+0rJRq1KbFPR4gTGlCqZApvJWfEHlgQxwlug6zFKaVy3vG04ggvS4GREd6XQeVjAE5cPY31Yrtdgll/BETHAxvy1+ucWxiFy6BNrqPni6XSOkSZc44EEIj4TCRAQdv5nZyd2VKPQHENYLDaC9KkxllZdqNuJuXx9stRv8auwOFRnF+JSk+7Q=="
	hostFileBackup     = "testdata/known_hosts_backup"
	wrongHostFile      = "testdata/wrong_known_hosts"
	badHostFile        = "testdata/bad_known_hosts"
)

func trustHostAlways(addr, algo, fingerprint string) bool {
	return true
}

func trustHostNever(addr, algo, fingerprint string) bool {
	return false
}

func TestAskToTrustHost(t *testing.T) {

	responses := map[string]bool{
		"Y":        true,
		"YES":      true,
		"Yes":      true,
		"y":        true,
		"yes":      true,
		"N":        false,
		"NO":       false,
		"No":       false,
		"n":        false,
		"no":       false,
		"anythign": false,
		"":         false,
	}

	for resp, expected := range responses {
		// using a pipe to simulate input from stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Error(err)
		}
		stdin := os.Stdin
		os.Stdin = r
		w.WriteString(resp)
		w.Close()
		if actual := askToTrustHost("addr", "algo", "fingerprint"); actual != expected {
			t.Errorf("Expected %v, got %v for %q", expected, actual, resp)
		}
		r.Close()
		os.Stdin = stdin
	}

}

// TestHostKeyChecker tests to check existing key
func TestHostKeyChecker(t *testing.T) {
	keyFile := NewHostKeyFile(hostFile)
	checker := NewHostKeyChecker(keyFile)

	addr, key, _ := parseKnownHostsLine([]byte(hostLine))
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	if err := checker.Check("localhost", tcpAddr, key); err != nil {
		t.Fatalf("checker should succeed for %v: %v", tcpAddr.String(), err)
	}

	wrongKey, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(wrongAuthorizedKey))
	if err := checker.Check("localhost", tcpAddr, wrongKey); err != ErrUnmatchKey {
		t.Fatalf("checker should fail with %v", ErrUnmatchKey)
	}
}

// TestHostKeyCheckerInteraction tests to check nonexisting key
func TestHostKeyCheckerInteraction(t *testing.T) {
	os.Remove(hostFileBackup)
	defer os.Remove(hostFileBackup)

	keyFile := NewHostKeyFile(hostFileBackup)
	checker := NewHostKeyChecker(keyFile)

	addr, key, _ := parseKnownHostsLine([]byte(hostLine))
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	// Refuse to add new host key
	checker.trustHost = trustHostNever
	if err := checker.Check("localhost", tcpAddr, key); err != ErrUntrustHost {
		t.Fatalf("checker should fail to put %v, %v in known_hosts", addr, tcpAddr.String())
	}

	// Accept to add new host key
	checker.trustHost = trustHostAlways
	if err := checker.Check("localhost", tcpAddr, key); err != nil {
		t.Fatalf("checker should succeed to put %v, %v in known_hosts", addr, tcpAddr.String())
	}

	// Use authorized key that have been added
	checker.trustHost = trustHostNever
	if err := checker.Check("localhost", tcpAddr, key); err != nil {
		t.Fatalf("checker should succeed to put %v, %v in known_hosts", addr, tcpAddr.String())
	}
}

// TestHostLine tests how to parse and render host line
func TestHostLine(t *testing.T) {
	addr, key, _ := parseKnownHostsLine([]byte(hostLine))
	if addr != addrInHostLine {
		t.Fatalf("addr is %v instead of %v", addr, addrInHostLine)
	}
	if key.Type() != ssh.KeyAlgoRSA {
		t.Fatalf("key type is %v instead of %v", key.Type(), ssh.KeyAlgoRSA)
	}

	line := renderHostLine(addr, key)
	if string(line) != hostLine {
		t.Fatal("unmatched host line after save and load")
	}
}

// TestHostKeyFile tests to read and write from HostKeyFile
func TestHostKeyFile(t *testing.T) {
	os.Remove(hostFileBackup)
	defer os.Remove(hostFileBackup)

	in := NewHostKeyFile(hostFile)
	out := NewHostKeyFile(hostFileBackup)

	hostKeys, err := in.GetHostKeys()
	if err != nil {
		t.Fatal("reading host file error:", err)
	}

	for i, v := range hostKeys {
		for _, k := range v {
			if err = out.PutHostKey(i, k); err != nil {
				t.Fatal("append error:", err)
			}
		}
	}

	keysByte, _ := ioutil.ReadFile(hostFile)
	keysByteBackup, _ := ioutil.ReadFile(hostFileBackup)
	keyBytes := bytes.Split(keysByte, []byte{'\n'})
	keyBytesBackup := bytes.Split(keysByteBackup, []byte{'\n'})
	for _, keyByte := range keyBytes {
		find := false
		for _, keyByteBackup := range keyBytesBackup {
			find = bytes.Compare(keyByte, keyByteBackup) == 0
			if find {
				break
			}
		}
		if !find {
			t.Fatalf("host file difference")
		}
	}
}

// TestHostKeyFile tests that reading and writing the wrong host key file fails
func TestWrongHostKeyFile(t *testing.T) {
	// Non-existent host key file should fail
	f := NewHostKeyFile(wrongHostFile)
	_, err := f.GetHostKeys()
	if err == nil {
		t.Fatal("should fail to read wrong host file")
	}
	if _, ok := err.(*os.PathError); !ok {
		t.Fatalf("should fail to read wrong host file due to file miss, but got %v", err)
	}

	// Create a host key file we do not have permission to read
	os.OpenFile(wrongHostFile, os.O_CREATE, 0000)
	defer os.Remove(wrongHostFile)
	// If run as root, drop privileges temporarily
	if id := syscall.Geteuid(); id == 0 {
		if err := syscall.Setuid(12345); err != nil {
			t.Fatalf("error setting uid: %v", err)
		}
		defer syscall.Setuid(id)
	}
	err = f.PutHostKey("", nil)
	if err == nil {
		t.Fatal("should fail to write wrong host file")
	}
	if !os.IsPermission(err) {
		t.Fatalf("should fail to write wrong host file due to permission denied, but got %v", err)
	}
}

// TestHostKeyFile tests to read from bad HostKeyFile
func TestBadHostKeyFile(t *testing.T) {
	f := NewHostKeyFile(badHostFile)
	hostKeys, _ := f.GetHostKeys()
	if len(hostKeys) > 0 {
		t.Fatal("read key from bad host file")
	}
}

// TestAlgorithmString tests the string representation of key algorithm
func TestAlgorithmString(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{ssh.KeyAlgoRSA, "RSA"},
		{ssh.KeyAlgoDSA, "DSA"},
		{ssh.KeyAlgoECDSA256, "ECDSA"},
		{ssh.KeyAlgoECDSA384, "ECDSA"},
		{ssh.KeyAlgoECDSA521, "ECDSA"},
		{"UNKNOWN", "UNKNOWN"},
	}
	for _, test := range tests {
		out := algoString(test.in)
		if out != test.out {
			t.Errorf("bad algo string for %s: got %s, want %s", test.in, out, test.out)
		}
	}

}

func TestMD5String(t *testing.T) {
	sum := [16]byte{0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	if md5String(sum) != "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff" {
		t.Fatal("wrong md5 string conversion")
	}
}

func TestAddrToHostPort(t *testing.T) {
	keyFile := NewHostKeyFile(hostFile)
	checker := NewHostKeyChecker(keyFile)

	badAddrs := []string{
		"12:12:12",
		"foobar:baz",
		"[12:323",
		"[127.0.0.1:]",
		// raw IPv6 addresses should fail
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		"2001:db8:85a3:0:0:8a2e:370:7334",
		"2001:db8:85a3::8a2e:370:7334",
		"::1",
		"::",
		// IPv6 addresses without ports should fail
		"[2001:db8:85a3::8a2e:370:7334]",
		"[::1]",
	}

	for _, a := range badAddrs {
		_, err := checker.addrToHostPort(a)
		if err == nil {
			t.Errorf("addr %v did not fail hostport conversion!", a)
		}
	}

	goodAddrs := []struct {
		in  string
		out string
	}{
		{"foo.com", "foo.com"},
		{"127.0.0.1", "127.0.0.1"},
		{"127.0.0.1:0", "127.0.0.1"},
		{"127.0.0.1:" + strconv.Itoa(sshDefaultPort), "127.0.0.1"},
		{"127.0.0.1:12345", "[127.0.0.1]:12345"},
		{"foo.com:" + strconv.Itoa(sshDefaultPort), "foo.com"},
		{"foo.com:2222", "[foo.com]:2222"},
		// escaped IPv6 addresses with ports should succeed
		{"[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:22", "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]"},
		{"[2001:db8:85a3:0:0:8a2e:370:7334]:12345", "[2001:db8:85a3:0:0:8a2e:370:7334]:12345"},
		{"[2001:db8:85a3::8a2e:370:7334]:12345", "[2001:db8:85a3::8a2e:370:7334]:12345"},
		{"[::1]:22", "[::1]"},
	}

	for _, a := range goodAddrs {
		got, err := checker.addrToHostPort(a.in)
		if err != nil {
			t.Errorf("addr %s failed hostport conversation: %v", a.in, err)
			continue
		}
		if got != a.out {
			t.Errorf("bad hostport conversion for %s: got %s, want %s", a.in, got, a.out)
		}
	}
}

func TestResolvePath(t *testing.T) {

	usr, err := user.Current()
	if err != nil {
		t.Error(err)
	}

	if len(usr.HomeDir) == 0 {
		t.Error("Unable to determine home dir for user")
	}

	if err := os.Setenv("TESTA", "a"); err != nil {
		t.Error(err)
	}
	if err := os.Setenv("TESTB", "b"); err != nil {
		t.Error(err)
	}
	if err := os.Setenv("TESTC", "c"); err != nil {
		t.Error(err)
	}

	paths := map[string]string{
		"~/.ssh":                     filepath.Join(usr.HomeDir, ".ssh"),
		"$TESTA/$TESTB/$TESTC/.ssh/": "a/b/c/.ssh",
		"/home/~/$TESTA/":            "/home/~/a",
		"~/$TESTA/.ssh/":             filepath.Join(usr.HomeDir, "/a/.ssh"),
	}

	for path, expected := range paths {
		if actual := resolvePath(path); actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	}

}
