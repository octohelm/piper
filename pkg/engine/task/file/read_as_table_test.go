package file

import (
	"bytes"
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func Test_readAsTable(t *testing.T) {
	rows := (&tableReader{}).ReadAsTable(bytes.NewBufferString(`
# /etc/fstab: static file system information.
#
# Use 'blkid' to print the universally unique identifier for a
# device; this may be used with UUID= as a more robust way to name devices
# that works even if disks are added and removed. See fstab(5).
#
# <file system> <mount point>   <type>  <options>       <dump>  <pass>
# / was on /dev/vda2 during installation
UUID=d13f95d5-4ac4-478b-bdb8-0efe1192e2c6 /               ext4    errors=remount-ro 0       1
# /boot/efi was on /dev/vda1 during installation
UUID=7306-10E6  /boot/efi       vfat    umask=0077      0       1
`))

	testingx.Expect(t, rows, testingx.Equal([][]string{
		{"UUID=d13f95d5-4ac4-478b-bdb8-0efe1192e2c6", "/", "ext4", "errors=remount-ro", "0", "1"},
		{"UUID=7306-10E6", "/boot/efi", "vfat", "umask=0077", "0", "1"},
	}))
}
