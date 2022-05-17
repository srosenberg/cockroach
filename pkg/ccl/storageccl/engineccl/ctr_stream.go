package engineccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl/engineccl/enginepbccl"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
)

type FileCipherStreamCreator struct {
	envType    enginepb.EnvType
	keyManager PebbleKeyManager
}

const (
	ctrBlockSize = 16
	ctrNonceSize = 12
)

func (c *FileCipherStreamCreator) CreateNew(
	ctx context.Context,
) (*enginepbccl.EncryptionSettings, FileStream, error) {
	__antithesis_instrumentation__.Notify(23454)
	key, err := c.keyManager.ActiveKey(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(23460)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(23461)
	}
	__antithesis_instrumentation__.Notify(23455)
	settings := &enginepbccl.EncryptionSettings{}
	if key == nil || func() bool {
		__antithesis_instrumentation__.Notify(23462)
		return key.Info.EncryptionType == enginepbccl.EncryptionType_Plaintext == true
	}() == true {
		__antithesis_instrumentation__.Notify(23463)
		settings.EncryptionType = enginepbccl.EncryptionType_Plaintext
		stream := &filePlainStream{}
		return settings, stream, nil
	} else {
		__antithesis_instrumentation__.Notify(23464)
	}
	__antithesis_instrumentation__.Notify(23456)
	settings.EncryptionType = key.Info.EncryptionType
	settings.KeyId = key.Info.KeyId
	settings.Nonce = make([]byte, ctrNonceSize)
	_, err = rand.Read(settings.Nonce)
	if err != nil {
		__antithesis_instrumentation__.Notify(23465)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(23466)
	}
	__antithesis_instrumentation__.Notify(23457)
	counterBytes := make([]byte, 4)
	if _, err = rand.Read(counterBytes); err != nil {
		__antithesis_instrumentation__.Notify(23467)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(23468)
	}
	__antithesis_instrumentation__.Notify(23458)

	settings.Counter = binary.LittleEndian.Uint32(counterBytes)
	ctrCS, err := newCTRBlockCipherStream(key, settings.Nonce, settings.Counter)
	if err != nil {
		__antithesis_instrumentation__.Notify(23469)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(23470)
	}
	__antithesis_instrumentation__.Notify(23459)
	return settings, &fileCipherStream{bcs: ctrCS}, nil
}

func (c *FileCipherStreamCreator) CreateExisting(
	settings *enginepbccl.EncryptionSettings,
) (FileStream, error) {
	__antithesis_instrumentation__.Notify(23471)
	if settings == nil || func() bool {
		__antithesis_instrumentation__.Notify(23475)
		return settings.EncryptionType == enginepbccl.EncryptionType_Plaintext == true
	}() == true {
		__antithesis_instrumentation__.Notify(23476)
		return &filePlainStream{}, nil
	} else {
		__antithesis_instrumentation__.Notify(23477)
	}
	__antithesis_instrumentation__.Notify(23472)
	key, err := c.keyManager.GetKey(settings.KeyId)
	if err != nil {
		__antithesis_instrumentation__.Notify(23478)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23479)
	}
	__antithesis_instrumentation__.Notify(23473)
	ctrCS, err := newCTRBlockCipherStream(key, settings.Nonce, settings.Counter)
	if err != nil {
		__antithesis_instrumentation__.Notify(23480)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23481)
	}
	__antithesis_instrumentation__.Notify(23474)
	return &fileCipherStream{bcs: ctrCS}, nil
}

type FileStream interface {
	Encrypt(fileOffset int64, data []byte)

	Decrypt(fileOffset int64, data []byte)
}

type filePlainStream struct{}

func (s *filePlainStream) Encrypt(fileOffset int64, data []byte) {
	__antithesis_instrumentation__.Notify(23482)
}
func (s *filePlainStream) Decrypt(fileOffset int64, data []byte) {
	__antithesis_instrumentation__.Notify(23483)
}

type fileCipherStream struct {
	bcs *cTRBlockCipherStream
}

func (s *fileCipherStream) Encrypt(fileOffset int64, data []byte) {
	__antithesis_instrumentation__.Notify(23484)
	if len(data) == 0 {
		__antithesis_instrumentation__.Notify(23486)
		return
	} else {
		__antithesis_instrumentation__.Notify(23487)
	}
	__antithesis_instrumentation__.Notify(23485)
	blockIndex := uint64(fileOffset / int64(ctrBlockSize))
	blockOffset := int(fileOffset % int64(ctrBlockSize))

	var buf struct {
		dataScratch [ctrBlockSize]byte
		ivScratch   [ctrBlockSize]byte
	}
	for len(data) > 0 {
		__antithesis_instrumentation__.Notify(23488)

		byteCount := ctrBlockSize - blockOffset
		if byteCount > len(data) {
			__antithesis_instrumentation__.Notify(23491)

			byteCount = len(data)
		} else {
			__antithesis_instrumentation__.Notify(23492)
		}
		__antithesis_instrumentation__.Notify(23489)
		if byteCount < int(ctrBlockSize) {
			__antithesis_instrumentation__.Notify(23493)

			copy(buf.dataScratch[blockOffset:blockOffset+byteCount], data[:byteCount])
			s.bcs.transform(blockIndex, buf.dataScratch[:], buf.ivScratch[:])

			copy(data[:byteCount], buf.dataScratch[blockOffset:blockOffset+byteCount])
			blockOffset = 0
		} else {
			__antithesis_instrumentation__.Notify(23494)
			s.bcs.transform(blockIndex, data[:byteCount], buf.ivScratch[:])
		}
		__antithesis_instrumentation__.Notify(23490)
		blockIndex++
		data = data[byteCount:]
	}
}

func (s *fileCipherStream) Decrypt(fileOffset int64, data []byte) {
	__antithesis_instrumentation__.Notify(23495)
	s.Encrypt(fileOffset, data)
}

type cTRBlockCipherStream struct {
	key     *enginepbccl.SecretKey
	nonce   [ctrNonceSize]byte
	counter uint32

	cBlock cipher.Block
}

func newCTRBlockCipherStream(
	key *enginepbccl.SecretKey, nonce []byte, counter uint32,
) (*cTRBlockCipherStream, error) {
	__antithesis_instrumentation__.Notify(23496)
	switch key.Info.EncryptionType {
	case enginepbccl.EncryptionType_AES128_CTR:
		__antithesis_instrumentation__.Notify(23500)
	case enginepbccl.EncryptionType_AES192_CTR:
		__antithesis_instrumentation__.Notify(23501)
	case enginepbccl.EncryptionType_AES256_CTR:
		__antithesis_instrumentation__.Notify(23502)
	default:
		__antithesis_instrumentation__.Notify(23503)
		return nil, fmt.Errorf("unknown EncryptionType: %d", key.Info.EncryptionType)
	}
	__antithesis_instrumentation__.Notify(23497)
	stream := &cTRBlockCipherStream{key: key, counter: counter}

	copy(stream.nonce[:], nonce)
	var err error
	if stream.cBlock, err = aes.NewCipher(key.Key); err != nil {
		__antithesis_instrumentation__.Notify(23504)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23505)
	}
	__antithesis_instrumentation__.Notify(23498)
	if stream.cBlock.BlockSize() != ctrBlockSize {
		__antithesis_instrumentation__.Notify(23506)
		return nil, fmt.Errorf("unexpected block size: %d", stream.cBlock.BlockSize())
	} else {
		__antithesis_instrumentation__.Notify(23507)
	}
	__antithesis_instrumentation__.Notify(23499)
	return stream, nil
}

func (s *cTRBlockCipherStream) transform(blockIndex uint64, data []byte, scratch []byte) {
	__antithesis_instrumentation__.Notify(23508)
	iv := append(scratch[:0], s.nonce[:]...)
	var blockCounter = uint32(uint64(s.counter) + blockIndex)
	binary.BigEndian.PutUint32(iv[len(iv):len(iv)+4], blockCounter)
	iv = iv[0 : len(iv)+4]
	s.cBlock.Encrypt(iv, iv)
	for i := 0; i < ctrBlockSize; i++ {
		__antithesis_instrumentation__.Notify(23509)
		data[i] = data[i] ^ iv[i]
	}
}
