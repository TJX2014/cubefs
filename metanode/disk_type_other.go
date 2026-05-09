//go:build !linux

package metanode

func isNvmeDisk(dirPath string) (bool, string, error) {
	return true, "", nil
}
