package db

import (
	"os"
	"sync"
	"time"

	"github.com/findy-network/findy-grpc/backup"
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
	db    *bolt.DB
	dirty bool
	l     sync.Mutex
}

func (m *Mgd) operate(f func(db *bolt.DB) error) (err error) {
	defer err2.Annotate("db operate", &err)

	m.l.Lock()
	defer m.l.Unlock()

	m.dirty = true
	if m.db == nil {
		err2.Check(m.open())
	}
	return f(mgedDB.db)
}

var (
	mgedDB Mgd
)

// Init initializes managed version of the encrypted database. Database is ready
// to use after this call. See more information of Cfg struct.
func Init(cfg Cfg) (err error) {
	mgedDB = Mgd{
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

// Data is general data element for encrypted database. It offers placeholders
// for read, write, and use operators to over write.
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

// close closes managed encrypted db. Note! Instance must be locked!
func (m *Mgd) close() (err error) {
	defer err2.Return(&err)

	glog.V(1).Infoln("close DB", m.Filename)
	err2.Check(m.db.Close())
	m.db = nil
	return nil
}

func (m *Mgd) backupName() string {
	timeStr := time.Now().Format(time.RFC3339)
	return backup.PrefixName(timeStr, m.BackupName)
}

// AddKeyValueToBucket add value to bucket pointed by the index. keyValue and
// index use Data types operators to encrypt and hash data on the fly.
func AddKeyValueToBucket(bucket []byte, keyValue, index *Data) (err error) {
	return mgedDB.operate(func(DB *bolt.DB) error {
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

// GetKeyValueFromBucket writes keyValue data by the index from a bucket. It
// returns found if key value exists. Returns are only if it cannot perform the
// transaction successfully.
func GetKeyValueFromBucket(bucket []byte, index, keyValue *Data) (found bool, err error) {
	defer err2.Annotate("get value", &err)

	err2.Check(mgedDB.operate(func(DB *bolt.DB) error {
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

// BackupTicker creates a backup ticker which takes backup copy of the database
// file specified by the interval. Ticker can be stopped with returned done
// channel.
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
				_, err := Backup()
				if err != nil {
					glog.Errorln("backup ticker:", err)
				}
			}
		}
	}()
	return doneCh
}

// Backup takes backup copy of the database. Before backup the database is
// closed.
func Backup() (did bool, err error) {
	defer err2.Annotate("backup", &err)

	mgedDB.l.Lock()
	defer mgedDB.l.Unlock()

	if !mgedDB.dirty {
		glog.V(1).Infoln("db isn't dirty, skipping backup")
		return false, nil
	}
	if mgedDB.db != nil {
		err2.Check(mgedDB.close())
	}

	// we keep locks on during the whole copy, but try to do it as fast as
	// possible. If this would be critical we could first read the source file
	// when locks are on and then write the target file in a new gorountine.
	backupName := mgedDB.backupName()
	err2.Check(backup.FileCopy(mgedDB.Filename, backupName))
	glog.V(1).Infoln("successful backup to file:", backupName)

	mgedDB.dirty = false
	return true, nil
}

// Wipe removes the whole database and its master file.
func Wipe() (err error) {
	defer err2.Annotate("wipe", &err)

	mgedDB.l.Lock()
	defer mgedDB.l.Unlock()

	if mgedDB.db != nil {
		err2.Check(mgedDB.close())
	}

	return os.RemoveAll(mgedDB.Filename)
}

// Close closes the database. It can be used after that if wanted. Transactions
// opens the database when needed.
func Close() (err error) {
	mgedDB.l.Lock()
	defer mgedDB.l.Unlock()

	if mgedDB.db != nil {
		return mgedDB.close()
	}

	return nil
}
