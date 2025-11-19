package ssh

import (
	"fmt"
	"os"
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
// 如果提供了privateKeyPath，将使用私钥认证；否则使用密码认证
func NewClient(host, user, password, privateKeyPath string, port int) (*Client, error) {
	var authMethods []ssh.AuthMethod

	// 优先使用私钥认证
	if privateKeyPath != "" {
		key, err := loadPrivateKey(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(key))
	} else if password != "" {
		// 使用密码认证
		authMethods = append(authMethods, ssh.Password(password))
	} else {
		return nil, fmt.Errorf("neither password nor private key provided")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 信任所有主机密钥
		Timeout:         10 * time.Second,
	}

	return &Client{
		config: config,
		host:   host,
		port:   port,
	}, nil
}

// loadPrivateKey 从文件加载SSH私钥
func loadPrivateKey(path string) (ssh.Signer, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// 尝试解析私钥（支持无密码保护的私钥）
	key, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
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
