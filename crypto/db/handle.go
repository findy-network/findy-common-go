package db

import "time"

type Handle interface {
	AddKeyValueToBucket(bucket []byte, keyValue, index *Data) (err error)
	RmKeyValueFromBucket(bucket []byte, index *Data) (err error)
	GetKeyValueFromBucket(bucket []byte, index, keyValue *Data) (found bool, err error)
	GetAllValuesFromBucket(bucket []byte, transforms ...Filter) (values [][]byte, err error)
	BackupTicker(interval time.Duration) (done chan<- struct{})
	Backup() (did bool, err error)
	Wipe() (err error)
	Close() (err error)
}
