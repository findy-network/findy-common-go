package backup

import (
	"path/filepath"

	"github.com/golang/glog"
)

func PrefixName(prefix, name string) string {
	dir, file := filepath.Split(name)
	file = prefix + "_" + file
	backupName := filepath.Join(dir, file)
	glog.V(3).Infoln("backup name:", backupName)
	return backupName
}
