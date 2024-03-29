package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

// Cipher is type for our block AES cipher
type Cipher struct {
	block  cipher.Block
	aesGCM cipher.AEAD
}

// NewCipher creates a new cipher with key 32-byte data given
func NewCipher(k []byte) *Cipher {
	assert.SLen(k, 32)

	defer err2.Catch(err2.Err(func(err error) {
		glog.Error(err)
	}))

	newBlock := try.To1(aes.NewCipher(k))

	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	// https://golang.org/pkg/crypto/cipher/#NewGCM
	newAesGCM := try.To1(cipher.NewGCM(newBlock))

	return &Cipher{block: newBlock, aesGCM: newAesGCM}
}

// Encrypt is same as TryEncrypt but note used yet
func (c *Cipher) Encrypt(in []byte) (out []byte) {
	return c.TryEncrypt(in)
}

// TryEncrypt encrypts dat with the cipher. This will be merged with Encrypt
// function later.
func (c *Cipher) TryEncrypt(in []byte) (out []byte) {
	nonceSize := c.aesGCM.NonceSize()
	nonce := make([]byte, nonceSize)

	n, err := io.ReadFull(rand.Reader, nonce)
	assert.Equal(n, nonceSize)
	assert.NoError(err)

	// We add it as a prefix to the encrypted data. The first nonce argument in
	// Seal is the prefix.
	return c.aesGCM.Seal(nonce, nonce, in, nil)
}

// Decrypt is same as TryDecrypt but note used yet
func (c *Cipher) _(in []byte) (out []byte, err error) {
	defer err2.Handle(&err)
	return c.TryDecrypt(in), nil
}

func (c *Cipher) TryDecrypt(in []byte) (out []byte) {
	nonceSize := c.aesGCM.NonceSize()

	nonce, ciphertext := in[:nonceSize], in[nonceSize:]

	return try.To1(c.aesGCM.Open(nil, nonce, ciphertext, nil))
}
