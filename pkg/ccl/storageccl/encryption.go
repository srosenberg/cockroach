package storageccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	crypto_rand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"

	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble/sstable"
	"github.com/cockroachdb/pebble/vfs"
	"golang.org/x/crypto/pbkdf2"
)

var encryptionPreamble = []byte("encrypt")

const encryptionSaltSize = 16

const encryptionVersionIVPrefix = 1

const encryptionVersionChunk = 2

var encryptionChunkSizeV2 = 64 << 10

const nonceSize = 12
const headerSize = 7 + 1 + nonceSize
const tagSize = 16

func GenerateSalt() ([]byte, error) {
	__antithesis_instrumentation__.Notify(23342)

	salt := make([]byte, encryptionSaltSize)
	if _, err := crypto_rand.Read(salt); err != nil {
		__antithesis_instrumentation__.Notify(23344)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23345)
	}
	__antithesis_instrumentation__.Notify(23343)
	return salt, nil
}

func GenerateKey(passphrase, salt []byte) []byte {
	__antithesis_instrumentation__.Notify(23346)
	return pbkdf2.Key(passphrase, salt, 64000, 32, sha256.New)
}

func AppearsEncrypted(text []byte) bool {
	__antithesis_instrumentation__.Notify(23347)
	return bytes.HasPrefix(text, encryptionPreamble)
}

type encWriter struct {
	gcm        cipher.AEAD
	iv         []byte
	ciphertext io.WriteCloser
	buf        []byte
	bufPos     int
}

type NopCloser struct {
	io.Writer
}

func (NopCloser) Close() error {
	__antithesis_instrumentation__.Notify(23348)
	return nil
}

func EncryptFile(plaintext, key []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(23349)
	b := &bytes.Buffer{}
	w, err := EncryptingWriter(NopCloser{b}, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(23353)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23354)
	}
	__antithesis_instrumentation__.Notify(23350)
	_, err = io.Copy(w, bytes.NewReader(plaintext))
	if err != nil {
		__antithesis_instrumentation__.Notify(23355)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23356)
	}
	__antithesis_instrumentation__.Notify(23351)
	if err := w.Close(); err != nil {
		__antithesis_instrumentation__.Notify(23357)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23358)
	}
	__antithesis_instrumentation__.Notify(23352)
	return b.Bytes(), nil
}

func EncryptingWriter(ciphertext io.WriteCloser, key []byte) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(23359)
	gcm, err := aesgcm(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(23363)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23364)
	}
	__antithesis_instrumentation__.Notify(23360)

	header := make([]byte, len(encryptionPreamble)+1+nonceSize)
	copy(header, encryptionPreamble)
	header[len(encryptionPreamble)] = encryptionVersionChunk

	ivStart := len(encryptionPreamble) + 1
	iv := make([]byte, nonceSize)
	if _, err := crypto_rand.Read(iv); err != nil {
		__antithesis_instrumentation__.Notify(23365)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23366)
	}
	__antithesis_instrumentation__.Notify(23361)
	copy(header[ivStart:], iv)

	if n, err := ciphertext.Write(header); err != nil {
		__antithesis_instrumentation__.Notify(23367)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23368)
		if n != len(header) {
			__antithesis_instrumentation__.Notify(23369)
			return nil, io.ErrShortWrite
		} else {
			__antithesis_instrumentation__.Notify(23370)
		}
	}
	__antithesis_instrumentation__.Notify(23362)

	return &encWriter{gcm: gcm, iv: iv, ciphertext: ciphertext, buf: make([]byte, encryptionChunkSizeV2+tagSize)}, nil
}

func (e *encWriter) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(23371)
	var wrote int
	for wrote < len(p) {
		__antithesis_instrumentation__.Notify(23373)
		copied := copy(e.buf[e.bufPos:encryptionChunkSizeV2], p[wrote:])
		e.bufPos += copied
		if e.bufPos == encryptionChunkSizeV2 {
			__antithesis_instrumentation__.Notify(23375)
			if err := e.flush(); err != nil {
				__antithesis_instrumentation__.Notify(23376)
				return wrote, err
			} else {
				__antithesis_instrumentation__.Notify(23377)
			}
		} else {
			__antithesis_instrumentation__.Notify(23378)
		}
		__antithesis_instrumentation__.Notify(23374)
		wrote += copied
	}
	__antithesis_instrumentation__.Notify(23372)
	return wrote, nil
}

func (e *encWriter) Close() error {
	__antithesis_instrumentation__.Notify(23379)

	err := e.flush()
	return errors.CombineErrors(err, e.ciphertext.Close())
}

func (e *encWriter) flush() error {
	__antithesis_instrumentation__.Notify(23380)
	e.buf = e.gcm.Seal(e.buf[:0], e.iv, e.buf[:e.bufPos], nil)
	for flushed := 0; flushed < len(e.buf); {
		__antithesis_instrumentation__.Notify(23382)
		n, err := e.ciphertext.Write(e.buf[flushed:])
		flushed += n
		if err != nil {
			__antithesis_instrumentation__.Notify(23383)
			return err
		} else {
			__antithesis_instrumentation__.Notify(23384)
		}
	}
	__antithesis_instrumentation__.Notify(23381)
	binary.BigEndian.PutUint64(e.iv[4:], binary.BigEndian.Uint64(e.iv[4:])+1)
	e.bufPos = 0
	return nil
}

func DecryptFile(
	ctx context.Context, ciphertext, key []byte, mm *mon.BoundAccount,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(23385)
	r, err := decryptingReader(bytes.NewReader(ciphertext), key)
	if err != nil {
		__antithesis_instrumentation__.Notify(23387)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23388)
	}
	__antithesis_instrumentation__.Notify(23386)
	return mon.ReadAll(ctx, ioctx.ReaderAdapter(r.(io.Reader)), mm)
}

type decryptReader struct {
	ciphertext io.ReaderAt
	g          cipher.AEAD
	fileIV     []byte

	ivScratch []byte
	buf       []byte
	pos       int64
	chunk     int64
}

type readerAndReaderAt interface {
	io.Reader
	io.ReaderAt
}

func decryptingReader(ciphertext readerAndReaderAt, key []byte) (sstable.ReadableFile, error) {
	__antithesis_instrumentation__.Notify(23389)
	gcm, err := aesgcm(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(23395)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23396)
	}
	__antithesis_instrumentation__.Notify(23390)

	header := make([]byte, headerSize)
	_, readHeaderErr := io.ReadFull(ciphertext, header)

	if !AppearsEncrypted(header) {
		__antithesis_instrumentation__.Notify(23397)
		return nil, errors.New("file does not appear to be encrypted")
	} else {
		__antithesis_instrumentation__.Notify(23398)
	}
	__antithesis_instrumentation__.Notify(23391)
	if readHeaderErr != nil {
		__antithesis_instrumentation__.Notify(23399)
		return nil, errors.Wrap(readHeaderErr, "invalid encryption header")
	} else {
		__antithesis_instrumentation__.Notify(23400)
	}
	__antithesis_instrumentation__.Notify(23392)

	version := header[len(encryptionPreamble)]
	if version < encryptionVersionIVPrefix || func() bool {
		__antithesis_instrumentation__.Notify(23401)
		return version > encryptionVersionChunk == true
	}() == true {
		__antithesis_instrumentation__.Notify(23402)
		return nil, errors.Errorf("unexpected encryption scheme/config version %d", version)
	} else {
		__antithesis_instrumentation__.Notify(23403)
	}
	__antithesis_instrumentation__.Notify(23393)
	iv := header[len(encryptionPreamble)+1:]

	if version == encryptionVersionIVPrefix {
		__antithesis_instrumentation__.Notify(23404)
		buf, err := ioutil.ReadAll(ciphertext)
		if err != nil {
			__antithesis_instrumentation__.Notify(23406)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23407)
		}
		__antithesis_instrumentation__.Notify(23405)
		buf, err = gcm.Open(buf[:0], iv, buf, nil)
		return vfs.NewMemFile(buf), errors.Wrap(err, "failed to decrypt — maybe incorrect key")
	} else {
		__antithesis_instrumentation__.Notify(23408)
	}
	__antithesis_instrumentation__.Notify(23394)
	buf := make([]byte, nonceSize, encryptionChunkSizeV2+tagSize+nonceSize)
	ivScratch := buf[:nonceSize]
	buf = buf[nonceSize:]
	r := &decryptReader{g: gcm, fileIV: iv, ivScratch: ivScratch, ciphertext: ciphertext, buf: buf, chunk: -1}
	return r, err
}

func (r *decryptReader) fill(chunk int64) error {
	__antithesis_instrumentation__.Notify(23409)
	if chunk == r.chunk {
		__antithesis_instrumentation__.Notify(23413)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(23414)
	}
	__antithesis_instrumentation__.Notify(23410)

	r.chunk = -1
	ciphertextChunkSize := int64(encryptionChunkSizeV2) + tagSize

	n, err := r.ciphertext.ReadAt(r.buf[:cap(r.buf)], headerSize+chunk*ciphertextChunkSize)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(23415)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(23416)
		return err
	} else {
		__antithesis_instrumentation__.Notify(23417)
	}
	__antithesis_instrumentation__.Notify(23411)
	r.buf = r.buf[:n]

	buf, err := r.g.Open(r.buf[:0], r.chunkIV(chunk), r.buf, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(23418)
		return errors.Wrap(err, "failed to decrypt — maybe incorrect key")
	} else {
		__antithesis_instrumentation__.Notify(23419)
	}
	__antithesis_instrumentation__.Notify(23412)
	r.buf = buf
	r.chunk = chunk
	return err
}

func (r *decryptReader) chunkIV(num int64) []byte {
	__antithesis_instrumentation__.Notify(23420)
	r.ivScratch = append(r.ivScratch[:0], r.fileIV...)
	binary.BigEndian.PutUint64(r.ivScratch[4:], binary.BigEndian.Uint64(r.ivScratch[4:])+uint64(num))
	return r.ivScratch
}

func (r *decryptReader) ReadAt(p []byte, offset int64) (int, error) {
	__antithesis_instrumentation__.Notify(23421)
	if offset < 0 {
		__antithesis_instrumentation__.Notify(23423)
		return 0, errors.New("bad offset")
	} else {
		__antithesis_instrumentation__.Notify(23424)
	}
	__antithesis_instrumentation__.Notify(23422)

	var read int
	for {
		__antithesis_instrumentation__.Notify(23425)
		chunk := offset / int64(encryptionChunkSizeV2)
		offsetInChunk := offset % int64(encryptionChunkSizeV2)

		if err := r.fill(chunk); err != nil {
			__antithesis_instrumentation__.Notify(23430)
			return read, err
		} else {
			__antithesis_instrumentation__.Notify(23431)
		}
		__antithesis_instrumentation__.Notify(23426)

		if offsetInChunk >= int64(len(r.buf)) {
			__antithesis_instrumentation__.Notify(23432)
			return read, io.EOF
		} else {
			__antithesis_instrumentation__.Notify(23433)
		}
		__antithesis_instrumentation__.Notify(23427)

		n := copy(p[read:], r.buf[offsetInChunk:])
		read += n

		if read == len(p) {
			__antithesis_instrumentation__.Notify(23434)
			return read, nil
		} else {
			__antithesis_instrumentation__.Notify(23435)
		}
		__antithesis_instrumentation__.Notify(23428)

		if len(r.buf) < encryptionChunkSizeV2 {
			__antithesis_instrumentation__.Notify(23436)
			return read, io.EOF
		} else {
			__antithesis_instrumentation__.Notify(23437)
		}
		__antithesis_instrumentation__.Notify(23429)

		offset += int64(n)
	}
}

func (r *decryptReader) Read(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(23438)
	n, err := r.ReadAt(p, r.pos)
	r.pos += int64(n)
	return n, err
}

func (r *decryptReader) Close() error {
	__antithesis_instrumentation__.Notify(23439)
	if closer, ok := r.ciphertext.(io.Closer); ok {
		__antithesis_instrumentation__.Notify(23441)
		return closer.Close()
	} else {
		__antithesis_instrumentation__.Notify(23442)
	}
	__antithesis_instrumentation__.Notify(23440)
	return nil
}

func (r *decryptReader) Stat() (os.FileInfo, error) {
	__antithesis_instrumentation__.Notify(23443)
	stater, ok := r.ciphertext.(interface{ Stat() (os.FileInfo, error) })
	if !ok {
		__antithesis_instrumentation__.Notify(23446)
		return nil, errors.Newf("%T does not support stat", r.ciphertext)
	} else {
		__antithesis_instrumentation__.Notify(23447)
	}
	__antithesis_instrumentation__.Notify(23444)
	stat, err := stater.Stat()
	if err != nil {
		__antithesis_instrumentation__.Notify(23448)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23449)
	}
	__antithesis_instrumentation__.Notify(23445)

	size := stat.Size()
	size -= headerSize
	size -= tagSize * ((size / (int64(encryptionChunkSizeV2) + tagSize)) + 1)
	return sizeStat(size), nil
}

func aesgcm(key []byte) (cipher.AEAD, error) {
	__antithesis_instrumentation__.Notify(23450)
	block, err := aes.NewCipher(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(23452)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23453)
	}
	__antithesis_instrumentation__.Notify(23451)
	return cipher.NewGCM(block)
}
