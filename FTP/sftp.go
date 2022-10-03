package sftp

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	Username     string
	Password     string
	PrivateKey   string
	Server       string
	KeyExchanges []string

	Timeout time.Duration
}

type Client struct {
	config     Config
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func New(config Config) (*Client, error) {
	c := &Client{
		config: config,
	}
 
	if err := c.connect(); err != nil {
		return nil, err
	}
 
	return c, nil
}

func (c *Client) Create(filePath string) (io.ReadWriteCloser, error) {
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
 
	return c.sftpClient.Create(filePath)
}

func (c *Client) Upload(source io.Reader, destination io.Writer, size int) error {
	if err := c.connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
 
	chunk := make([]byte, size)
 
	for {
		num, err := source.Read(chunk)
		if err == io.EOF {
			tot, err := destination.Write(chunk[:num])
			if err != nil {
				return err
			}
 
			if tot != len(chunk[:num]) {
				return fmt.Errorf("failed to write stream")
			}
 
			return nil
		}
 
		if err != nil {
			return err
		}
 
		tot, err := destination.Write(chunk[:num])
		if err != nil {
			return err
		}
 
		if tot != len(chunk[:num]) {
			return fmt.Errorf("failed to write stream")
		}
	}
}

func (c *Client) Download(filePath string) (io.ReadCloser, error) {
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
 
	return c.sftpClient.Open(filePath)
}

func (c *Client) Info(filePath string) (os.FileInfo, error) {
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
 
	info, err := c.sftpClient.Lstat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file stats: %w", err)
	}
 
	return info, nil
}

func (c *Client) Close() {
	if c.sftpClient != nil {
		c.sftpClient.Close()
	}
	if c.sshClient != nil {
		c.sshClient.Close()
	}
}

func (c *Client) connect() error {
	if c.sshClient != nil {
		_, _, err := c.sshClient.SendRequest("keepalive", false, nil)
		if err == nil {
			return nil
		}
	}
 
	auth := ssh.Password(c.config.Password)
	if c.config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(c.config.PrivateKey))
		if err != nil {
			return fmt.Errorf("ssh parse private key: %w", err)
		}
		auth = ssh.PublicKeys(signer)
	}
 
	cfg := &ssh.ClientConfig{
		User: c.config.Username,
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: func(string, net.Addr, ssh.PublicKey) error { return nil },
		Timeout:         c.config.Timeout,
		Config: ssh.Config{
			KeyExchanges: c.config.KeyExchanges,
		},
	}
 
	sshClient, err := ssh.Dial("tcp", c.config.Server, cfg)
	if err != nil {
		return fmt.Errorf("ssh dial: %w", err)
	}
	c.sshClient = sshClient
 
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("sftp new client: %w", err)
	}
	c.sftpClient = sftpClient
 
	return nil
}