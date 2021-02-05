package db

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/lainio/err2"
	bolt "go.etcd.io/bbolt"
)

type Cfg struct {
	Filename   string
	BackupName string
	Buckets    [][]byte
}

type Mgd struct {
	Cfg
	db *bolt.DB
	l  sync.RWMutex
}

func (m *Mgd) operate(f func(db *bolt.DB) error) (err error) {
	defer err2.Annotate("db operate", &err)

	m.l.Lock()
	defer m.l.Unlock()

	if m.db == nil {
		err2.Check(m.open())
	}
	return f(MgdDB.db)
}

var (
	MgdDB Mgd
)

func Init(cfg Cfg) (err error) {
	MgdDB = Mgd{
		Cfg: cfg,
	}
	return nil
}

func (m *Mgd) open() (err error) {
	defer err2.Return(&err)

	glog.V(1).Infoln("open DB", m.Filename)
	m.db, err = bolt.Open(m.Filename, 0600, nil)
	err2.Check(err)

	err2.Check(m.db.Update(func(tx *bolt.Tx) (err error) {
		defer err2.Annotate("create buckets", &err)

		for _, bucket := range m.Buckets {
			err2.Try(tx.CreateBucketIfNotExists(bucket))
		}
		return nil
	}))
	return err
}

type Filter func(value []byte) (k []byte)
type Use func(value []byte) interface{}

type Data struct {
	Data  []byte
	Read  Filter
	Write Filter
	Use
	Result interface{}
}

func (d *Data) get() []byte {
	if d.Read == nil {
		return append(d.Data[:0:0], d.Data...)
	}
	return d.Read(d.Data)
}

func (d *Data) set(b []byte) {
	if d.Write == nil {
		if d.Use != nil {
			d.Result = d.Use(b)
		} else {
			copy(d.Data, b)
		}
	} else {
		d.Data = d.Write(b)
		if d.Use != nil {
			d.Result = d.Use(d.Data)
		}
	}
}

// close closes the sealed box of the enclave. It can be open again with
// InitSealedBox.
func (m *Mgd) close() (err error) {
	defer err2.Return(&err)

	glog.V(1).Infoln("close DB", m.Filename)
	err2.Check(m.db.Close())
	m.db = nil
	return nil
}

func (m *Mgd) backupName() string {
	tsStr := time.Now().Format(time.RFC3339)
	backupName := tsStr + "_" + m.BackupName
	glog.V(3).Infoln("backup name:", backupName)
	return backupName
}

func AddKeyValueToBucket(bucket []byte, keyValue, index *Data) (err error) {
	return MgdDB.operate(func(DB *bolt.DB) error {
		defer err2.Annotate("add key", &err)

		err2.Check(DB.Update(func(tx *bolt.Tx) (err error) {
			defer err2.Return(&err)

			b := tx.Bucket(bucket)
			err2.Check(b.Put(index.get(), keyValue.get()))
			return nil
		}))
		return nil
	})
}

func GetKeyValueFromBucket(bucket []byte, index, keyValue *Data) (found bool, err error) {
	defer err2.Annotate("get value", &err)

	err2.Check(MgdDB.operate(func(DB *bolt.DB) error {
		err2.Check(DB.View(func(tx *bolt.Tx) (err error) {
			defer err2.Return(&err)

			b := tx.Bucket(bucket)
			d := b.Get(index.get())
			if d == nil {
				found = false
				return nil
			}
			keyValue.set(d)
			found = true
			return nil
		}))
		return nil
	}))
	return found, nil
}

func BackupTicker(interval time.Duration) (done chan<- struct{}) {
	ticker := time.NewTicker(interval)
	doneCh := make(chan struct{})
	go func() {
		defer err2.CatchTrace(func(err error) {
			glog.Error(err)
		})
		for {
			select {
			case <-doneCh:
				return
			case <-ticker.C:
				err := Backup()
				if err != nil {
					glog.Errorln("backup ticker:", err)
				}
			}
		}
	}()
	return doneCh
}

func Backup() (err error) {
	defer err2.Annotate("backup", &err)

	MgdDB.l.Lock()
	defer MgdDB.l.Unlock()

	if MgdDB.db != nil {
		glog.V(1).Infoln("close DB")
		err2.Check(MgdDB.close())
	}

	// we keep locks on during the whole copy, but try to do it as fast as
	// possible. If this would be critical we could first read the source file
	// when locks are on and then write the target file in a new gorountine.
	backupName := MgdDB.backupName()
	err2.Check(fileCopy(MgdDB.Filename, backupName))
	glog.V(1).Infoln("successful backup to file:", backupName)

	return nil
}

func Wipe() (err error) {
	defer err2.Annotate("wipe", &err)

	MgdDB.l.Lock()
	defer MgdDB.l.Unlock()

	if MgdDB.db != nil {
		err2.Check(MgdDB.close())
	}

	return os.RemoveAll(MgdDB.Filename)
}

func fileCopy(src, dst string) (err error) {
	defer err2.Returnf(&err, "copy %s %s", src, dst)

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
