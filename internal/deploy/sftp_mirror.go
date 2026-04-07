package deploy

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pkg/sftp"

	"github.com/SomniSom/docker-ops/internal/locale"
)

type localFile struct {
	rel      string
	abs      string
	info     os.FileInfo
}

// MirrorProjectTree uploads the project tree to remoteRoot via SFTP, mirroring rsync --delete semantics
// using size + mtime comparison (readme §4.3). Excluded paths are not sent; extra remote files are removed.
func MirrorProjectTree(c *sftp.Client, localRoot, remoteRoot string, patterns []string) error {
	localRoot = filepath.Clean(localRoot)
	remoteRoot = filepath.ToSlash(filepath.Clean(remoteRoot))

	var locals []localFile
	err := filepath.WalkDir(localRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(localRoot, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if PathExcluded(rel, patterns) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		locals = append(locals, localFile{rel: rel, abs: p, info: info})
		return nil
	})
	if err != nil {
		return err
	}

	localSet := make(map[string]struct{}, len(locals))
	for _, lf := range locals {
		localSet[lf.rel] = struct{}{}
	}

	remoteFiles, err := collectRemoteFileRelPaths(c, remoteRoot)
	if err != nil {
		return fmt.Errorf("%s: %w", locale.T("deploy.mirror.list"), err)
	}

	for rf := range remoteFiles {
		if _, ok := localSet[rf]; !ok {
			abs := remoteJoin(remoteRoot, rf)
			if err := c.Remove(abs); err != nil {
				return fmt.Errorf("%s: %w", locale.Tf("deploy.mirror.remove", abs), err)
			}
		}
	}

	slices.SortFunc(locals, func(a, b localFile) int {
		return strings.Compare(a.rel, b.rel)
	})

	for _, lf := range locals {
		rem := remoteJoin(remoteRoot, lf.rel)
		parent := path.Dir(filepath.ToSlash(rem))
		if parent != "." && parent != "/" {
			if err := sftpMkdirAll(c, parent); err != nil {
				return fmt.Errorf("%s: %w", locale.Tf("deploy.mirror.mkdir", parent), err)
			}
		}

		need, err := needsUpload(c, lf, rem)
		if err != nil || !need {
			if err != nil {
				return err
			}
			continue
		}
		if err := uploadFile(c, lf.abs, rem, lf.info); err != nil {
			return fmt.Errorf("%s: %w", locale.Tf("deploy.mirror.upload", lf.rel), err)
		}
	}

	return pruneEmptyDirs(c, remoteRoot)
}

func remoteJoin(root, rel string) string {
	root = filepath.ToSlash(filepath.Clean(root))
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "." {
		return root
	}
	return strings.TrimSuffix(root, "/") + "/" + rel
}

func collectRemoteFileRelPaths(c *sftp.Client, remoteRoot string) (map[string]struct{}, error) {
	remoteRoot = filepath.ToSlash(filepath.Clean(remoteRoot))
	out := make(map[string]struct{})
	var walk func(dir string) error
	walk = func(abs string) error {
		entries, err := c.ReadDir(abs)
		if err != nil {
			return err
		}
		for _, e := range entries {
			name := e.Name()
			if name == "." || name == ".." {
				continue
			}
			full := strings.TrimSuffix(abs, "/") + "/" + name
			rel, err := relUnderRoot(remoteRoot, full)
			if err != nil {
				return err
			}
			if e.IsDir() {
				if err := walk(full); err != nil {
					return err
				}
				continue
			}
			out[rel] = struct{}{}
		}
		return nil
	}
	if err := walk(remoteRoot); err != nil {
		return nil, err
	}
	return out, nil
}

func relUnderRoot(root, full string) (string, error) {
	root = filepath.ToSlash(filepath.Clean(root))
	full = filepath.ToSlash(filepath.Clean(full))
	if full == root {
		return ".", nil
	}
	prefix := strings.TrimSuffix(root, "/") + "/"
	if !strings.HasPrefix(full, prefix) {
		return "", fmt.Errorf("%s", locale.Tf("deploy.mirror.not_under", full, root))
	}
	return strings.TrimPrefix(full, prefix), nil
}

func needsUpload(c *sftp.Client, lf localFile, rem string) (bool, error) {
	st, err := c.Stat(rem)
	if err != nil {
		return true, nil
	}
	if lf.info.Size() != st.Size() {
		return true, nil
	}
	lt := lf.info.ModTime().Unix()
	rt := st.ModTime().Unix()
	if lt != rt {
		return true, nil
	}
	return false, nil
}

func uploadFile(c *sftp.Client, localPath, rem string, info os.FileInfo) error {
	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := c.Create(rem)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	mode := info.Mode().Perm() & 0o777
	if mode != 0 {
		_ = c.Chmod(rem, mode)
	}
	return nil
}

// pruneEmptyDirs removes empty directories under root, deepest first (best-effort).
func pruneEmptyDirs(c *sftp.Client, root string) error {
	root = filepath.ToSlash(filepath.Clean(root))
	var dirs []string
	var walk func(abs string)
	walk = func(abs string) {
		entries, err := c.ReadDir(abs)
		if err != nil {
			return
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if name == "." || name == ".." {
				continue
			}
			full := strings.TrimSuffix(abs, "/") + "/" + name
			walk(full)
		}
		if abs != root {
			dirs = append(dirs, abs)
		}
	}
	walk(root)
	slices.SortFunc(dirs, func(a, b string) int {
		return len(b) - len(a)
	})
	for _, d := range dirs {
		entries, err := c.ReadDir(d)
		if err != nil || len(entries) != 0 {
			continue
		}
		_ = c.RemoveDirectory(d)
	}
	return nil
}
