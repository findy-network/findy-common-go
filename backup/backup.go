/*
Package backup implements dedicated helpers for the current backup system.
*/
package backup

import (
	"io"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/lainio/err2"
)

// PrefixName builds a new prefixed file name. The file name is in format:
//  prefix_file.name
// If prefix is empty the returned string starts with _.
func PrefixName(prefix, name string) string {
	dir, file := filepath.Split(name)
	file = prefix + "_" + file
	backupName := filepath.Join(dir, file)
	glog.V(3).Infoln("backup name:", backupName)
	return backupName
}

// FileCopy copies a source file to destination file.
func FileCopy(src, dst string) (err error) {
	defer err2.Returnf(&err, "copy %s -> %s", src, dst)

	r := err2.File.Try(os.Open(src))
	defer r.Close()

	w := err2.File.Try(os.Create(dst))
	defer err2.Handle(&err, func() {
		os.Remove(dst)
	})
	defer w.Close()
	err2.Empty.Try(io.Copy(w, r))
	return nil
}
