package crypto

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/lainio/err2/try"
)

const hexKey = "15308490f1e4026284594dd08d31291bc8ef2aeac730d0daf6ff87bb92d4336c"

var (
	k = try.To1(hex.DecodeString(hexKey))
)

func TestNewCipher(t *testing.T) {
	type args struct {
		k []byte
	}
	tests := []struct {
		name string
		args args
		want *Cipher
	}{
		{"simple", args{k}, NewCipher(k)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCipher(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCipher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCipher_TryDecrypt(t *testing.T) {
	type args struct {
		in []byte
	}
	tests := []struct {
		name    string
		args    args
		wantOut []byte
	}{
		{"simple", args{NewCipher(k).Encrypt(k)}, k},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCipher(k)
			if gotOut := c.TryDecrypt(tt.args.in); !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("Cipher.TryEncrypt() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
