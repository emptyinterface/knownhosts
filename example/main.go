package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/emptyinterface/knownhosts"
	"golang.org/x/crypto/ssh"
)

var (
	u, _       = user.Current()
	ident      = flag.String("i", filepath.Join(u.HomeDir, ".ssh/id_rsa"), "key to use")
	username   = flag.String("u", u.Username, "user to connect as")
	knownHosts = flag.String("h", filepath.Join(u.HomeDir, ".ssh/known_hosts"), "known hosts to verify against")
)

func NewSSHConfig(username, ident, knownHosts string) *ssh.ClientConfig {

	data, err := ioutil.ReadFile(ident)
	if err != nil {
		log.Fatal(err)
	}

	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		log.Fatal(err)
	}

	key := knownhosts.NewHostKeyFile(knownHosts)
	check := knownhosts.NewHostKeyChecker(key)

	return &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: check.Check,
	}

}

func main() {

	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Please specify a host:port and command to run")
		fmt.Printf("%s 10.0.0.1:22 whoami\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	config := NewSSHConfig(*username, *ident, *knownHosts)

	fmt.Println("Dialing", args[0])

	client, err := ssh.Dial("tcp", args[0], config)
	if err != nil {
		log.Fatal(err)
	}

	sess, err := client.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	sess.Stdin = os.Stdin
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	cmd := strings.Join(args[1:], " ")

	fmt.Printf("running %q\n", cmd)

	if err := sess.Run(cmd); err != nil {
		log.Fatal(err)
	}

}
