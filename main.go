package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh/knownhosts"

	ansi "github.com/jhunt/go-ansi"
	"golang.org/x/crypto/ssh"
)

func main() {
	if len(os.Args) != 2 {
		bailWith(`Expected 1 argument, received %d
  Usage: get-host-key <hostname>`, len(os.Args)-1)
	}

	hostname := os.Args[1]
	schemeRegexp := regexp.MustCompile(fmt.Sprintf("^.*%s", regexp.QuoteMeta("://")))

	//Hostname needs port, so we'll give it port 22 if not provided... but
	//url.Parse gets confused if the URL is nothing but a hostname, so we'll give
	//it a temporary scheme.
	if !schemeRegexp.Match([]byte(hostname)) {
		hostname = "ssh://" + hostname
	}
	u, err := url.Parse(hostname)
	if err != nil {
		bailWith("Could not parse hostname as URL: %s", err)
	}

	if u.Port() == "" {
		u.Host = u.Host + ":22"
		hostname = u.String()
	}

	hostname = schemeRegexp.ReplaceAllString(hostname, "")

	sshConf := &ssh.ClientConfig{
		HostKeyCallback: getHostKey,
	}

	_, err = ssh.Dial("tcp", hostname, sshConf)
	if err != nil {
		//Not providing any auth, so I expect an auth failure
		if strings.Contains(err.Error(), "unable to authenticate") {
			os.Exit(0)
		}

		bailWith("Error on SSH connection: %s", err)
	}
}

func getHostKey(address string, _ net.Addr, key ssh.PublicKey) error {
	address = knownhosts.Normalize(address)
	fmt.Println(knownhosts.Line([]string{address}, key))
	return nil
}

func bailWith(format string, args ...interface{}) {
	ansi.Fprintf(os.Stderr, "@R{"+format+"}\n", args...)
	os.Exit(1)
}
