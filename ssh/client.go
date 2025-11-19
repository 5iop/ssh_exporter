package ssh

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client SSH客户端
type Client struct {
	config *ssh.ClientConfig
	host   string
	port   int
	conn   *ssh.Client
}

// NewClient 创建新的SSH客户端
func NewClient(host, user, password string, port int) *Client {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	return &Client{
		config: config,
		host:   host,
		port:   port,
	}
}

// Connect 连接到SSH服务器
func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := ssh.Dial("tcp", addr, c.config)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	c.conn = conn
	return nil
}

// ExecuteCommand 执行命令
func (c *Client) ExecuteCommand(command string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		// 即使有错误，也返回输出（某些命令可能会有部分输出）
		return string(output), err
	}

	return string(output), nil
}

// Close 关闭连接
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
