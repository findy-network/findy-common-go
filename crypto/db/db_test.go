package db

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/stretchr/testify/assert"
)

const dbFilename = "fido-enclave.bolt"

var buckets = [][]byte{{01, 01}}

func TestMain(m *testing.M) {
	err2.Check(flag.Set("logtostderr", "true"))
	err2.Check(flag.Set("v", "3"))

	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	_ = os.RemoveAll(dbFilename)
	glog.V(1).Infoln("init enclave", dbFilename)
	sealedBoxFilename := dbFilename
	backupName := "backup-" + sealedBoxFilename
	err2.Check(Init(Cfg{
		Filename:   sealedBoxFilename,
		BackupName: backupName,
		Buckets:    buckets,
	}))

	// insert data to DB to that it's open when first tests are started
	err2.Check(AddKeyValueToBucket(buckets[0],
		&Data{
			Data: []byte{0, 0, 1, 1, 1, 1},
			Read: encrypt,
		},
		&Data{
			Data: []byte{0, 0, 1, 1, 1, 1},
			Read: hash,
		},
	))

}

func tearDown() {
	err2.Check(Wipe())
	removeFiles(".", "*"+dbFilename)
}

func removeFiles(home, nameFilter string) {
	filter := filepath.Join(home, nameFilter)
	files, _ := filepath.Glob(filter)
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			panic(err)
		}
	}
}

func TestGetKeyValueFromBucket(t *testing.T) {
	value := &Data{
		Write: decrypt,
	}

	already, err := GetKeyValueFromBucket(buckets[0],
		&Data{
			Data: []byte{0, 0, 1, 1, 1, 1},
			Read: hash,
		},
		value,
	)
	assert.NoError(t, err)
	assert.True(t, already)
	assert.Equal(t, value.Data, []byte{0, 0, 1, 1, 1, 1})
}

func TestRm(t *testing.T) {
	err := AddKeyValueToBucket(buckets[0],
		&Data{
			Data: []byte{1, 0, 1, 1, 1, 1},
			Read: encrypt,
		},
		&Data{
			Data: []byte{1, 0, 1, 1, 1, 1},
			Read: hash,
		},
	)
	assert.NoError(t, err)
	err = RmKeyValueFromBucket(buckets[0],
		&Data{
			Data: []byte{1, 0, 1, 1, 1, 1},
			Read: encrypt,
		},
	)
	assert.NoError(t, err)

	// let's check that we actually removed the key/value pair
	value := &Data{
		Write: decrypt,
	}
	already, err := GetKeyValueFromBucket(buckets[0],
		&Data{
			Data: []byte{1, 0, 1, 1, 1, 1},
			Read: hash,
		},
		value,
	)
	assert.NoError(t, err)
	assert.False(t, already)
}

func TestBackup(t *testing.T) {
	tests := []struct {
		name       string
		dirtyAfter bool
		wantDid    bool
		wantErr    bool
	}{
		{name: "db already dirty", dirtyAfter: true, wantDid: true, wantErr: false},
		{name: "we made it dirty", dirtyAfter: false, wantDid: true, wantErr: false},

		{name: "db is clean", dirtyAfter: false, wantDid: false, wantErr: false},
		{name: "db is still clean", dirtyAfter: true, wantDid: false, wantErr: false},

		{name: "db dirty again", dirtyAfter: true, wantDid: true, wantErr: false},
		{name: "we made it dirty", dirtyAfter: false, wantDid: true, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			did, err := Backup()
			if (err != nil) != tt.wantErr {
				t.Errorf("Backup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if did != tt.wantDid {
				t.Errorf("Backup() did = %v, wantDid %v", did, tt.wantDid)
			}

			if tt.dirtyAfter {
				err2.Check(AddKeyValueToBucket(buckets[0],
					&Data{
						Data: []byte{0, 0, 1, 1, 1, 1},
						Read: encrypt,
					},
					&Data{
						Data: []byte{0, 0, 1, 1, 1, 1},
						Read: hash,
					},
				))
			}
		})
	}
}

// hash makes the cryptographic hash of the map key value. This prevents us to
// store key value index (email, DID) to the DB aka sealed box as plain text.
// Please use salt when implementing this.
func hash(key []byte) (k []byte) {
	return append(key[:0:0], key...)
}

// encrypt encrypts the actual wallet key value. This is used when data is
// stored do the DB aka sealed box.
func encrypt(value []byte) (k []byte) {
	return append(value[:0:0], value...)
}

// decrypt decrypts the actual wallet key value. This is used when data is
// retrieved from the DB aka sealed box.
func decrypt(value []byte) (k []byte) {
	return append(value[:0:0], value...)
}

// noop function if need e.g. tests
func _(value []byte) (k []byte) {
	return value
}
