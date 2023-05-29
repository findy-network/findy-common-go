package db

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type bucket struct {
	sync.Mutex
	id   string
	data map[string][]byte
}

type memDB struct {
	buckets map[string]*bucket

	name string // for logs mainly
	on   OnFn
}

// NewMemDB creates new memory database. The memory DB has same interface (Handle)
// as the normal Bolt DB, but instead of writing data to file it leaves into the
// memory. The DB is meant for the tests and performance measurements.
func NewMemDB(bucketNames [][]byte, a ...string) Handle {
	buckets := make(map[string]*bucket, len(bucketNames))

	for _, id := range bucketNames {
		name := string(id)
		b := &bucket{id: name, data: make(map[string][]byte, 12)}
		buckets[name] = b
	}
	name := ""
	if len(a) > 0 {
		name = a[0]
	}
	return &memDB{buckets: buckets, name: name}

}

func (db *memDB) AddKeyValueToBucket(bucket []byte, keyValue, index *Data) (err error) {
	defer err2.Handle(&err)
	try.To(db.checkIsOn())

	b := db.buckets[string(bucket)]
	b.Lock()
	defer b.Unlock()

	key := string(index.get())
	b.data[key] = keyValue.get()
	return nil
}

func (db *memDB) RmKeyValueFromBucket(bucket []byte, index *Data) (err error) {
	defer err2.Handle(&err)
	try.To(db.checkIsOn())

	b := db.buckets[string(bucket)]
	b.Lock()
	defer b.Unlock()

	key := string(index.get())
	delete(b.data, key)
	return nil
}

func (db *memDB) GetKeyValueFromBucket(
	bucket []byte,
	index, keyValue *Data,
) (
	found bool,
	err error,
) {
	defer err2.Handle(&err)
	try.To(db.checkIsOn())

	b := db.buckets[string(bucket)]
	b.Lock()
	defer b.Unlock()

	var d []byte
	d, found = b.data[string(index.get())]
	if !found {
		return
	}
	keyValue.set(d)
	return
}

func (db *memDB) GetAllValuesFromBucket(
	bucket []byte,
	transforms ...Filter,
) (
	values [][]byte,
	err error,
) {
	defer err2.Handle(&err)
	try.To(db.checkIsOn())

	b := db.buckets[string(bucket)]
	b.Lock()
	defer b.Unlock()

	values = make([][]byte, 0, len(b.data))

	for _, res := range b.data {
		for _, transform := range transforms {
			res = transform(res)
		}
		values = append(values, res)
	}
	return
}

func (db *memDB) BackupTicker(interval time.Duration) (done chan<- struct{}) {
	return nil
}

func (db *memDB) Backup() (did bool, err error) {
	defer err2.Handle(&err)
	try.To(db.checkIsOn())
	return false, nil
}

func (db *memDB) Wipe() (err error) {
	defer err2.Handle(&err)
	try.To(db.checkIsOn())
	return nil
}

func (db *memDB) Close() (err error) {
	glog.V(1).Infoln("closing mem db:", db.name)
	return nil
}

func (db *memDB) SetStatusFn(f OnFn) {
	db.on = f
}

func (db *memDB) isOn() bool {
	if db.on == nil {
		return true
	}
	return db.on()
}

// checkIsOn checks if db is on. If not returns an error.
// TODO: we could use x.Whom function here?
func (db *memDB) checkIsOn() (err error) {
	if !db.isOn() {
		return ErrDisabledDB // TODO: use err2 value future
	}
	return nil
}
