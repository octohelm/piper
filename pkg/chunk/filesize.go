package chunk

import (
	"github.com/octohelm/piper/pkg/units"
)

type FileSize int64

const (
	KiB FileSize = 1024
	MiB FileSize = 1024 * KiB
	GiB FileSize = 1024 * MiB
	TiB FileSize = 1024 * GiB
	PiB FileSize = 1024 * TiB
)

func (f FileSize) String() string {
	return units.CustomSize("%.4g%s", float64(f), 1024.0, binaryAbbrs)
}

var binaryAbbrs = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
