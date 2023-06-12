// package db implements Bolt based database which can be encrypted on the fly
// and which supports automatic backups. It offers very simple API and hides all
// the complex stuff behind it. It's thread safe. More information see the Cfg
// struct.
package db

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/findy-network/findy-common-go/backup"
	"github.com/findy-network/findy-common-go/x"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
	bolt "go.etcd.io/bbolt"
)

const MEM_PREFIX = "MEMORY_"

// Cfg is configuration needed to create and open managed database that is
// implemented with Bolt DB or by memory maps for testing and profiling. See
// Filename for more information.
type Cfg struct {
	// Filename is full path file name of the DB file. Note, if the base
	// of the filename starts with MEM_PREFIX, the memory database is created.
	// That's useful e.g. testing and profiling.
	Filename string

	// Base part of the backup file names. Date and time is added.
	BackupName string

	// Buckets is slice of the bucket names that are in byte slice.
	Buckets [][]byte
}

// Mgd is a managed and encrypted (option, can be pre-procession as well) DB.
type Mgd struct {
	Cfg
	bdb   *bolt.DB
	dirty bool
	l     sync.Mutex

	// attributes to handle enabling/disabling the db
	// disabled bool
	on OnFn // tells if we are on or off, allows us to share one flag
}

// operate is a key element of the managed Bolt DB. It keeps track of closing
// and opening of the DB which is needed that DB can operate and backups can be
// taken without explicitly closing the database.
func (db *Mgd) operate(f func(db *bolt.DB) error) (err error) {
	defer err2.Handle(&err, "db operate")

	db.l.Lock()
	defer db.l.Unlock()

	try.To(db.checkIsOn())

	db.dirty = true
	if db.bdb == nil {
		try.To(db.open())
	}
	return f(db.bdb)
}

type sHandles = []Handle

var (
	mgedDB Handle

	instances = x.NewRWSlice[sHandles](0, 12)
)

// New creates a new managed and encrypted database. This is a preferred way to
// use the managed database package. There is also the alternated Init function
// when you don't need to store the Mgd instance by yourself. It's for the cases
// when only one managed database is needed per a process or an application.
// Database is ready to use after this call. You don't need to open it and
// backup can be taken during the run. See more information of Cfg struct.
func New(cfg Cfg) Handle {
	base := filepath.Base(cfg.Filename)
	if strings.HasPrefix(base, MEM_PREFIX) {
		glog.V(5).Infoln("MEMORY-DB open:", base)
		return addInstance(NewMemDB(cfg.Buckets, cfg.Filename))
	}
	glog.V(5).Infof("File system DB (%v)", base)
	return addInstance(&Mgd{
		Cfg: cfg,
	})
}

// GracefulStop closes all database instances immediately.
func GracefulStop() {
	instances.Rx(func(s sHandles) {
		for _, db := range s {
			if err := db.Close(); err != nil {
				glog.Warning(err)
			}
		}
	})
}

// Init initializes managed version of the encrypted database. Database is ready
// to use after this call. See more information of Cfg struct.
func Init(cfg Cfg) (err error) {
	mgedDB = New(cfg)
	return nil
}

func addInstance(h Handle) Handle {
	return instances.Add(h)
}

func (db *Mgd) open() (err error) {
	defer err2.Handle(&err)

	try.To(db.checkIsOn())

	glog.V(1).Infoln("open DB", db.Filename)
	db.bdb = try.To1(bolt.Open(db.Filename, 0600, nil))

	try.To(db.bdb.Update(func(tx *bolt.Tx) (err error) {
		defer err2.Handle(&err, "create buckets")

		for _, bucket := range db.Buckets {
			try.To1(tx.CreateBucketIfNotExists(bucket))
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
func (db *Mgd) close() (err error) {
	defer err2.Handle(&err)

	glog.V(1).Infoln("close DB", db.Filename)
	try.To(db.bdb.Close())
	db.bdb = nil
	return nil
}

func (db *Mgd) backupName() string {
	timeStr := time.Now().Format(time.RFC3339)
	return backup.PrefixName(timeStr, db.BackupName)
}

// AddKeyValueToBucket add value to bucket pointed by the index. keyValue and
// index use Data type's operators to encrypt and hash data on the fly.
func (db *Mgd) AddKeyValueToBucket(bucket []byte, keyValue, index *Data) (err error) {
	return db.operate(func(DB *bolt.DB) error {
		defer err2.Handle(&err, "add key")

		try.To(DB.Update(func(tx *bolt.Tx) (err error) {
			defer err2.Handle(&err)

			b := tx.Bucket(bucket)
			try.To(b.Put(index.get(), keyValue.get()))
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
		defer err2.Handle(&err, "rm key")

		try.To(DB.Update(func(tx *bolt.Tx) (err error) {
			defer err2.Handle(&err)

			b := tx.Bucket(bucket)
			try.To(b.Delete(index.get()))
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
	defer err2.Handle(&err, "get value")

	try.To(db.operate(func(DB *bolt.DB) error {
		try.To(DB.View(func(tx *bolt.Tx) (err error) {
			defer err2.Handle(&err)

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

// GetAllValuesFromBucket returns all entries from the bucket.
// Note:
// - Order is not guaranteed.
// - The returned slice contains only the values as byte arrays. Keys are excluded.
// Transform functions can be used e.g. to decrypt the data. They are applied in the provided order.
// Errors will return only if it cannot perform the transaction successfully.
func GetAllValuesFromBucket(
	bucket []byte,
	transforms ...Filter,
) (
	values [][]byte,
	err error,
) {
	return mgedDB.GetAllValuesFromBucket(bucket, transforms...)
}

// GetAllValuesFromBucket returns all entries from the bucket.
// Note:
// - Order is not guaranteed.
// - The returned slice contains only the values as byte arrays. Keys are excluded.
// Transform functions can be used e.g. to decrypt the data. They are applied in the provided order.
// Errors will return only if it cannot perform the transaction successfully.
func (db *Mgd) GetAllValuesFromBucket(
	bucket []byte,
	transforms ...Filter,
) (
	values [][]byte,
	err error,
) {
	defer err2.Handle(&err, "get all values")

	values = make([][]byte, 0)

	try.To(db.operate(func(DB *bolt.DB) error {
		try.To(DB.View(func(tx *bolt.Tx) (err error) {
			defer err2.Handle(&err)

			b := tx.Bucket(bucket)
			try.To(b.ForEach(func(_, v []byte) error {
				res := v
				for _, transform := range transforms {
					res = transform(res)
				}
				values = append(values, res)
				return nil
			}))
			return nil
		}))
		return nil
	}))
	return values, nil
}

// BackupTicker creates a backup ticker which takes backup copy of the database
// file specified by the interval. Ticker can be stopped with returned done
// channel.
func BackupTicker(interval time.Duration) (done chan<- struct{}) {
	assert.INotNil(mgedDB, "call Init before use pkg lvl db instance")
	return mgedDB.BackupTicker(interval)
}

// BackupTicker creates a backup ticker which takes backup copy of the database
// file specified by the interval. Ticker can be stopped with returned done
// channel.
func (db *Mgd) BackupTicker(interval time.Duration) (done chan<- struct{}) {
	ticker := time.NewTicker(interval)
	doneCh := make(chan struct{})
	go func() {
		defer func() {
			glog.V(1).Infoln("exiting backup tickers")
		}()
		defer err2.Catch(func(err error) {
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
	assert.INotNil(mgedDB, "call Init before use pkg lvl db instance")
	return mgedDB.Backup()
}

// Backup takes backup copy of the database. Before backup the database is
// closed automatically and only dirty databases are backed up.
func (db *Mgd) Backup() (did bool, err error) {
	defer err2.Handle(&err, "backup")

	db.l.Lock()
	defer db.l.Unlock()

	try.To(db.checkIsOn())

	if !db.dirty {
		glog.V(1).Infoln("db isn't dirty, skipping backup")
		return false, nil
	}
	if db.bdb != nil {
		try.To(db.close())
	}

	// we keep locks on during the whole copy, but try to do it as fast as
	// possible. If this would be critical we could first read the source file
	// when locks are on and then write the target file in a new gorountine.
	backupName := db.backupName()
	try.To(backup.FileCopy(db.Filename, backupName))
	glog.V(1).Infoln("successful backup to file:", backupName)

	db.dirty = false
	return true, nil
}

// Wipe removes the whole database and its master file.
func Wipe() (err error) {
	assert.INotNil(mgedDB, "call Init before use pkg lvl db instance")
	return mgedDB.Wipe()
}

// Wipe removes the whole database and its master file.
func (db *Mgd) Wipe() (err error) {
	defer err2.Handle(&err, "wipe")

	db.l.Lock()
	defer db.l.Unlock()

	try.To(db.checkIsOn())

	if db.bdb != nil {
		try.To(db.close())
	}

	return os.RemoveAll(db.Filename)
}

// Close closes the database. It can be used after that if wanted. Transactions
// opens the database when needed.
func Close() (err error) {
	assert.INotNil(mgedDB, "call Init before use pkg lvl db instance")
	return mgedDB.Close()
}

// Close closes the database. It can be used after that if wanted. Transactions
// opens the database when needed.
func (db *Mgd) Close() (err error) {
	defer err2.Handle(&err)

	db.l.Lock()
	defer db.l.Unlock()

	if db.bdb != nil {
		return db.close()
	}

	return nil
}

func (db *Mgd) SetStatusFn(f OnFn) {
	db.l.Lock()
	defer db.l.Unlock()
	db.on = f
}

func (db *Mgd) isOn() bool {
	if db.on == nil {
		return true
	}
	return db.on()
}

// checkIsOn checks if db is on. If not returns an error.
// TODO: we could use x.Whom function here?
func (db *Mgd) checkIsOn() (err error) {
	if !db.isOn() {
		return ErrDisabledDB // TODO: use err2 value future
	}
	return nil
}
