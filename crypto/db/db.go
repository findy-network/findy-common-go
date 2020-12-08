package db

import (
	"errors"
	"fmt"

	"github.com/lainio/err2"
	bolt "go.etcd.io/bbolt"
)

var DB *bolt.DB

// ErrSealBoxAlreadyExists is an error for enclave sealed box already exists.
var ErrSealBoxAlreadyExists = errors.New("enclave sealed box exists")

func assertDB() {
	if DB == nil {
		panic("don't forget init the seal box")
	}
}

func Open(filename string, buckets [][]byte) (err error) {
	if DB != nil {
		return ErrSealBoxAlreadyExists
	}
	defer err2.Return(&err)

	DB, err = bolt.Open(filename, 0600, nil)
	err2.Check(err)

	err2.Check(DB.Update(func(tx *bolt.Tx) (err error) {
		defer err2.Annotate("create buckets", &err)

		for _, bucket := range buckets {
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

// Close closes the sealed box of the enclave. It can be open again with
// InitSealedBox.
func Close() {
	defer err2.CatchTrace(func(err error) {
		fmt.Println(err)
	})
	assertDB()

	err2.Check(DB.Close())
	DB = nil
}

func AddKeyValueToBucket(bucket []byte, keyValue, index *Data) (err error) {
	assertDB()

	defer err2.Annotate("add key", &err)

	err2.Check(DB.Update(func(tx *bolt.Tx) (err error) {
		defer err2.Return(&err)

		b := tx.Bucket(bucket)
		err2.Check(b.Put(index.get(), keyValue.get()))
		return nil
	}))
	return nil
}

func GetKeyValueFromBucket(bucket []byte, index, keyValue *Data) (found bool, err error) {
	assertDB()

	defer err2.Return(&err)

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
	return found, nil
}
