package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"code.google.com/p/go.crypto/ssh"
	"github.com/docopt/docopt-go"
)

func main() {
	usage := `Wakka Wakka.

 Usage:
   myapp [-hvcn] [--key=SSHKEY] [--user=USERNAME] TARGETS...
   myapp --version

 Arguments:
   TARGETS  targets for ssh

 Options:
   -h --help                    show this help message and exit
   --version                    show version and exit
   -v --verbose                 print status messages
   -c --cats                    just c
   -n --names                   just n
   -k=SSHKEY --key=SSHKEY       the ssh key
   -u=USERNAME --user=USERNAME  the username`

	args, _ := docopt.Parse(usage, nil, true, "1.0.0", false)
	//fmt.Printf("%v\n", args)

	var k string
	u := os.Getenv("USER")
	c := args["--cats"].(bool)
	n := args["--names"].(bool)
	hosts := args["TARGETS"].([]string)

	if args["--key"] != nil {
		k = args["--key"].(string)
		_, err := os.Stat(k)
		if err != nil {
			log.Fatal(err)
		}
	}
	if args["--user"] != nil {
		u = args["--user"].(string)
	}
	if c {
		fmt.Println(c)
	}
	if n {
		fmt.Println(n)
	}
	Responses := make(chan string)

	var wg sync.WaitGroup
	wg.Add(len(hosts))

	for _, h := range hosts {
		h = h + ":22"
		go func(h, user, k string) {
			defer wg.Done()
			cmd := "date"
			RunRemote(cmd, user, h, k)
			Responses <- string("1")
		}(h, u, k)
	}

	go func() {
		for response := range Responses {
			fmt.Println(response)
		}
	}()

	wg.Wait()
}

func parsekey(file string) ssh.Signer {
	privateBytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal(err)
	}
	return private
}

//func RunRemote(cmd, user, server string, pkey ssh.Signer) {
func RunRemote(cmd, user, server, key string) {
	pkey := parsekey(key)
	auths := []ssh.AuthMethod{ssh.PublicKeys(pkey)}

	cfg := &ssh.ClientConfig{
		User: user,
		Auth: auths,
	}
	cfg.SetDefaults()

	client, err := ssh.Dial("tcp", server, cfg)
	if err != nil {
		log.Println(err)
		return
	}

	var session *ssh.Session

	session, err = client.NewSession()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	fmt.Println(stdoutBuf.String())
}
