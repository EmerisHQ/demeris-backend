package database

import (
	"fmt"
	"io/ioutil"
	"net"

	"github.com/allinbits/navigator-backend/config"
	"golang.org/x/crypto/ssh"
)

// NewSshConnection is used to create an SSH connection to the server hosting the db.
// This will be used for development purposes
func NewSshConnection(c *config.Config) (*ssh.Client, error) {

	sshConfig := &ssh.ClientConfig{
		User: c.SshUser,
		Auth: []ssh.AuthMethod{
			publicKeyFile(c.KeyFile),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	return ssh.Dial("tcp", fmt.Sprintf("%s:%s", c.SshHost, c.SshPort), sshConfig)
}

// publicKeyFile takes a file path to a public key and returns it as an auth method.
func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}
