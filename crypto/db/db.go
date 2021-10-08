// package db implements Bolt based database which can be encrypted on the fly
// and which supports automatic backups. It offers very simple API and hides all
// the complex stuff behind it. It's thread safe. More information see the Cfg
// struct.
package db

import (
	"os"
	"sync"
	"time"

	"github.com/findy-network/findy-common-go/backup"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	bolt "go.etcd.io/bbolt"
)

// Cfg is configuration needed to create and open managed database.
type Cfg struct {
	Filename   string   // Filename is full path file name of the DB file
	BackupName string   // Base part of the backup file names. Date and time is added.
	Buckets    [][]byte // Buckets is list of the buckets needed
}

// Mgd is a managed and encrypted (option, can be pre-procession as well) DB.
type Mgd struct {
	Cfg
	db    *bolt.DB
	dirty bool
	l     sync.Mutex
}

// operate is a key element of the managed Bolt DB. It keeps track of closing
// and opening of the DB which is needed that DB can operate and backups can be
// taken without explicitly closing the database.
func (m *Mgd) operate(f func(db *bolt.DB) error) (err error) {
	defer err2.Annotate("db operate", &err)

	m.l.Lock()
	defer m.l.Unlock()

	m.dirty = true
	if m.db == nil {
		err2.Check(m.open())
	}
	return f(m.db)
}

var (
	mgedDB Mgd
)

// New creates a new managed and encrypted database. This is a preferred way to
// use the managed database package. There is also the alternated Init function when you
// don't need to store the Mgd instance by yourself. It's for the cases when
// only one managed database is needed per a process or an application. Database
// is ready to use after this call. You don't need to open it and backup can be
// taken during the run. See more information of Cfg struct.
func New(cfg Cfg) *Mgd {
	return &Mgd{
		Cfg: cfg,
	}
}

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
// index use Data type's operators to encrypt and hash data on the fly.
func (db *Mgd) AddKeyValueToBucket(bucket []byte, keyValue, index *Data) (err error) {
	return db.operate(func(DB *bolt.DB) error {
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

// AddKeyValueToBucket add value to bucket pointed by the index. keyValue and
// index use Data type's operators to encrypt and hash data on the fly.
func AddKeyValueToBucket(bucket []byte, keyValue, index *Data) (err error) {
	return mgedDB.AddKeyValueToBucket(bucket, keyValue, index)
}

// RmKeyValueFromBucket removes value pointed by the index from the bucket.
// The index uses Data type's operators to encrypt and hash data on the fly.
func (db *Mgd) RmKeyValueFromBucket(bucket []byte, index *Data) (err error) {
	return db.operate(func(DB *bolt.DB) error {
		defer err2.Annotate("rm key", &err)

		err2.Check(DB.Update(func(tx *bolt.Tx) (err error) {
			defer err2.Return(&err)

			b := tx.Bucket(bucket)
			err2.Check(b.Delete(index.get()))
			return nil
		}))
		return nil
	})
}

// RmKeyValueFromBucket removes value pointed by the index from the bucket.
// The index uses Data type's operators to encrypt and hash data on the fly.
func RmKeyValueFromBucket(bucket []byte, index *Data) (err error) {
	return mgedDB.RmKeyValueFromBucket(bucket, index)
}

// GetKeyValueFromBucket writes keyValue data by the index from a bucket. It
// returns `found` if key value exists. Errors will return only if it cannot
// perform the transaction successfully.
func GetKeyValueFromBucket(
	bucket []byte,
	index, keyValue *Data,
) (
	found bool,
	err error,
) {
	return mgedDB.GetKeyValueFromBucket(bucket, index, keyValue)
}

// GetKeyValueFromBucket writes keyValue data by the index from a bucket. It
// returns `found` if key value exists. Errors will return only if it cannot
// perform the transaction successfully.
func (db *Mgd) GetKeyValueFromBucket(
	bucket []byte,
	index, keyValue *Data,
) (
	found bool,
	err error,
) {
	defer err2.Annotate("get value", &err)

	err2.Check(db.operate(func(DB *bolt.DB) error {
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
	return mgedDB.BackupTicker(interval)
}

// BackupTicker creates a backup ticker which takes backup copy of the database
// file specified by the interval. Ticker can be stopped with returned done
// channel.
func (db *Mgd) BackupTicker(interval time.Duration) (done chan<- struct{}) {
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
				_, err := db.Backup()
				if err != nil {
					glog.Errorln("backup ticker:", err)
				}
			}
		}
	}()
	return doneCh
}

// Backup takes backup copy of the database. Before backup the database is
// closed automatically and only dirty databases are backed up.
func Backup() (did bool, err error) {
	return mgedDB.Backup()
}

// Backup takes backup copy of the database. Before backup the database is
// closed automatically and only dirty databases are backed up.
func (db *Mgd) Backup() (did bool, err error) {
	defer err2.Annotate("backup", &err)

	db.l.Lock()
	defer db.l.Unlock()

	if !db.dirty {
		glog.V(1).Infoln("db isn't dirty, skipping backup")
		return false, nil
	}
	if db.db != nil {
		err2.Check(db.close())
	}

	// we keep locks on during the whole copy, but try to do it as fast as
	// possible. If this would be critical we could first read the source file
	// when locks are on and then write the target file in a new gorountine.
	backupName := db.backupName()
	err2.Check(backup.FileCopy(db.Filename, backupName))
	glog.V(1).Infoln("successful backup to file:", backupName)

	db.dirty = false
	return true, nil
}

// Wipe removes the whole database and its master file.
func Wipe() (err error) {
	return mgedDB.Wipe()
}

// Wipe removes the whole database and its master file.
func (db *Mgd) Wipe() (err error) {
	defer err2.Annotate("wipe", &err)

	db.l.Lock()
	defer db.l.Unlock()

	if db.db != nil {
		err2.Check(db.close())
	}

	return os.RemoveAll(db.Filename)
}

// Close closes the database. It can be used after that if wanted. Transactions
// opens the database when needed.
func Close() (err error) {
	return mgedDB.close()
}

// Close closes the database. It can be used after that if wanted. Transactions
// opens the database when needed.
func (db *Mgd) Close() (err error) {
	db.l.Lock()
	defer db.l.Unlock()

	if db.db != nil {
		return db.close()
	}

	return nil
}
