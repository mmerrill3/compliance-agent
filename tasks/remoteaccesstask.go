package tasks

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

// RemoteAccessTask provides details on compliance for remote access to a host
type RemoteAccessTask struct {
	User, KeyFile, Host, FileName string
}

// Build is responsible for building file to put into S3
func (b *RemoteAccessTask) Build(cmd string) (output string, err error) {
	if output, err = b.accessAndRun(cmd); err != nil {
		glog.Infof("Error on running remote command %v", err)
	}
	return
}

func (b *RemoteAccessTask) GetFileName() string {
	return b.FileName
}

func (b *RemoteAccessTask) accessAndRun(cmd string) (output string, err error) {
	key, err := ioutil.ReadFile(b.KeyFile)
	if err != nil {
		glog.Fatalf("Unable to read private key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            b.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", b.Host), config)
	if err != nil {
		panic(err)
	}

	defer client.Close()
	session, err := client.NewSession()

	if err != nil {
		glog.Fatalf("unable to create session: %s", err)
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)
	output = stdoutBuf.String()
	return output, nil

}
