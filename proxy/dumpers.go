// +build ignore

package proxy

import (
	"io"
	"net/http"
	"os"
)

var dumpDst = os.Stdout

func TeeReadCloser(r io.ReadCloser, w io.Writer) io.ReadCloser {
	return &teeReadCloser{r, w}
}

type teeReadCloser struct {
	r io.ReadCloser
	w io.Writer
}

func (t *teeReadCloser) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

func (t *teeReadCloser) Close() error {
	return t.r.Close()
}

func dumpResponse(res *http.Response) error {
	res.Body = TeeReadCloser(res.Body, dumpDst)
	return nil
}
