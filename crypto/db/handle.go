package db

import (
	"errors"
	"time"
)

type Handle interface {
	AddKeyValueToBucket(bucket []byte, keyValue, index *Data) (err error)
	RmKeyValueFromBucket(bucket []byte, index *Data) (err error)
	GetKeyValueFromBucket(bucket []byte, index, keyValue *Data) (found bool, err error)
	GetAllValuesFromBucket(bucket []byte, transforms ...Filter) (values [][]byte, err error)
	BackupTicker(interval time.Duration) (done chan<- struct{})
	Backup() (did bool, err error)
	Wipe() (err error)
	Close() (err error)
	SetStatusFn(f OnFn)
}

// OnFn is call back to status of the database: on or off.
type OnFn func() bool

var ErrDisabledDB = errors.New("database is turned off")
