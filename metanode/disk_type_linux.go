//go:build linux

package metanode

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func isNvmeDisk(dirPath string) (bool, string, error) {
	if dirPath == "" {
		return false, "", fmt.Errorf("empty path")
	}

	p := filepath.Clean(dirPath)
	if rp, err := filepath.EvalSymlinks(p); err == nil && rp != "" {
		p = rp
	}

	source, err := mountSourceForPath(p)
	if err != nil {
		return false, "", err
	}
	if !strings.HasPrefix(source, "/dev/") {
		return false, source, nil
	}

	srcResolved := source
	if rr, err := filepath.EvalSymlinks(source); err == nil && rr != "" {
		srcResolved = rr
	}

	block := strings.TrimPrefix(srcResolved, "/dev/")
	if strings.HasPrefix(block, "nvme") {
		return true, source, nil
	}
	if isNvmeBySysfs(block) {
		return true, source, nil
	}
	return false, source, nil
}

type mountInfoEntry struct {
	mountPoint string
	source     string
}

func mountSourceForPath(absPath string) (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	var best mountInfoEntry
	bestLen := -1
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) != 2 {
			continue
		}
		left := strings.Fields(parts[0])
		right := strings.Fields(parts[1])
		if len(left) < 5 || len(right) < 2 {
			continue
		}
		mp := unescapeMountInfoPath(left[4])
		src := right[1]
		if !pathHasMountPrefix(absPath, mp) {
			continue
		}
		if len(mp) > bestLen {
			best = mountInfoEntry{mountPoint: mp, source: src}
			bestLen = len(mp)
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	if bestLen < 0 {
		return "", fmt.Errorf("no mountinfo entry matches path: %s", absPath)
	}
	return best.source, nil
}

func pathHasMountPrefix(p, mountPoint string) bool {
	if mountPoint == "/" || p == mountPoint {
		return true
	}
	if strings.HasPrefix(p, mountPoint) {
		return len(p) > len(mountPoint) && (p[len(mountPoint)] == '/' || mountPoint[len(mountPoint)-1] == '/')
	}
	return false
}

func unescapeMountInfoPath(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+3 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		oct := s[i+1 : i+4]
		if oct[0] < '0' || oct[0] > '7' || oct[1] < '0' || oct[1] > '7' || oct[2] < '0' || oct[2] > '7' {
			b.WriteByte(s[i])
			continue
		}
		v, err := strconv.ParseInt(oct, 8, 32)
		if err != nil {
			b.WriteByte(s[i])
			continue
		}
		b.WriteByte(byte(v))
		i += 3
	}
	return b.String()
}

func isNvmeBySysfs(block string) bool {
	if ok, nonRot := readNonRotational(block); ok {
		return nonRot
	}
	parent := parentBlockDevice(block)
	if parent != "" && parent != block {
		if ok, nonRot := readNonRotational(parent); ok {
			return nonRot
		}
	}
	return false
}

func readNonRotational(block string) (bool, bool) {
	rotPath := filepath.Join("/sys/class/block", block, "queue", "rotational")
	b, err := os.ReadFile(rotPath)
	if err != nil {
		rotPath = filepath.Join("/sys/block", block, "queue", "rotational")
		b, err = os.ReadFile(rotPath)
		if err != nil {
			return false, false
		}
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return false, false
	}
	return true, v == 0
}

func parentBlockDevice(block string) string {
	if _, err := os.Stat(filepath.Join("/sys/class/block", block, "partition")); err != nil {
		return block
	}
	rp, err := filepath.EvalSymlinks(filepath.Join("/sys/class/block", block))
	if err != nil || rp == "" {
		return block
	}
	parent := filepath.Base(filepath.Dir(rp))
	if parent == "" || parent == "." || parent == "/" {
		return block
	}
	return parent
}
