package fsm

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func GenerateURL(subPath string, m *Machine) (URL *url.URL, err error) {
	defer err2.Returnf(&err, "generate plantuml URL")
	rawPlantModel := m.String()
	deflated := tryDeflate([]byte(rawPlantModel))
	b64 := base64.RawStdEncoding.EncodeToString(deflated[2 : len(deflated)-4])
	codedModel := translate(b64)
	return url.Parse(fmt.Sprintf("http://www.plantuml.com/plantuml/%s/%s", subPath, codedModel))
}

// translate translates standard base64 string to plantuml's version:
//
//	normal base64:n
//	ABCDEFGHIJ KLMNOPQRSTUVWXYZ abcdefghij klmnopqrstuvwxyz 0123456789 +/
//	plant's 'close' to base64:
//	0123456789 ABCDEFGHIJKLMNOP QRSTUVWXYZ abcdefghijklmnop qrstuvwxyz -_
func translate(b64 string) string {
	trans := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'J':
			return '0' + (r - 'A')
		case r >= 'K' && r <= 'Z':
			return 'A' + (r - 'K')
		case r >= 'a' && r <= 'j':
			return 'Q' + (r - 'a')
		case r >= 'k' && r <= 'z':
			return 'a' + (r - 'k')
		case r >= '0' && r <= '9':
			return 'q' + (r - '0')
		case r == '+':
			return '-'
		case r == '/':
			return '_'
		}
		return r
	}
	return string(bytes.Map(trans, []byte(b64)))
}

func tryDeflate(d []byte) []byte {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	try.To1(w.Write(d))
	try.To(w.Close())
	return b.Bytes()
}
