package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type oneliner struct {
	backend io.Writer
	buf     *bytes.Buffer
	encoder *json.Encoder
}

func newOneliner(backend io.Writer) *oneliner {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	return &oneliner{
		backend: backend,
		buf:     buf,
		encoder: encoder,
	}
}

func (j *oneliner) Write(p []byte) (n int, err error) {
	j.buf.Reset()
	j.encoder.Encode(string(p))
	return j.backend.Write(j.buf.Bytes())
}

var Discard = log.New(ioutil.Discard, "", 0)

func Open(plain bool) (infoLog, errLog *log.Logger) {
	if !plain {
		infoLog = log.New(newOneliner(os.Stdout), "INF: ", log.LUTC|log.LstdFlags|log.Lmicroseconds)
		errLog = log.New(newOneliner(os.Stderr), "ERR: ", log.LUTC|log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	} else {
		infoLog = log.New(os.Stdout, "", 0)
		errLog = log.New(os.Stderr, "ERR: ", log.Lshortfile)
	}
	return
}
