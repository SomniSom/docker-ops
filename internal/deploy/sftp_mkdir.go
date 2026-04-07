package deploy

import (
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
)

// sftpMkdirAll creates path and parents on the SFTP server (Unix paths).
func sftpMkdirAll(c *sftp.Client, abs string) error {
	abs = filepath.ToSlash(filepath.Clean(abs))
	if abs == "" || abs == "." {
		return nil
	}
	parts := strings.Split(strings.TrimPrefix(abs, "/"), "/")
	cur := ""
	if strings.HasPrefix(abs, "/") {
		cur = "/"
	}
	for _, p := range parts {
		if p == "" {
			continue
		}
		if cur == "/" {
			cur = "/" + p
		} else if cur == "" {
			cur = p
		} else {
			cur = cur + "/" + p
		}
		if _, err := c.Stat(cur); err != nil {
			if err := c.Mkdir(cur); err != nil {
				return err
			}
		}
	}
	return nil
}
