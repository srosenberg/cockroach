package cloudtestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/blobs"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

var EConnRefused = &net.OpError{Err: &os.SyscallError{
	Syscall: "test",
	Err:     sysutil.ECONNREFUSED,
}}

var econnreset = &net.OpError{Err: &os.SyscallError{
	Syscall: "test",
	Err:     sysutil.ECONNRESET,
}}

type antagonisticDialer struct {
	net.Dialer
	rnd               *rand.Rand
	numRepeatFailures *int
}

type antagonisticConn struct {
	net.Conn
	rnd               *rand.Rand
	numRepeatFailures *int
}

func (d *antagonisticDialer) DialContext(
	ctx context.Context, network, addr string,
) (net.Conn, error) {
	__antithesis_instrumentation__.Notify(36079)
	if network == "tcp" {
		__antithesis_instrumentation__.Notify(36081)

		if *d.numRepeatFailures < cloud.MaxDelayedRetryAttempts-1 && func() bool {
			__antithesis_instrumentation__.Notify(36084)
			return d.rnd.Int()%2 == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(36085)
			*(d.numRepeatFailures)++
			return nil, EConnRefused
		} else {
			__antithesis_instrumentation__.Notify(36086)
		}
		__antithesis_instrumentation__.Notify(36082)
		c, err := d.Dialer.DialContext(ctx, network, addr)
		if err != nil {
			__antithesis_instrumentation__.Notify(36087)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(36088)
		}
		__antithesis_instrumentation__.Notify(36083)

		return &antagonisticConn{Conn: c, rnd: d.rnd, numRepeatFailures: d.numRepeatFailures}, nil
	} else {
		__antithesis_instrumentation__.Notify(36089)
	}
	__antithesis_instrumentation__.Notify(36080)
	return d.Dialer.DialContext(ctx, network, addr)
}

func (c *antagonisticConn) Read(b []byte) (int, error) {
	__antithesis_instrumentation__.Notify(36090)

	if *c.numRepeatFailures < cloud.MaxDelayedRetryAttempts-1 && func() bool {
		__antithesis_instrumentation__.Notify(36092)
		return c.rnd.Int()%2 == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(36093)
		*(c.numRepeatFailures)++
		return 0, econnreset
	} else {
		__antithesis_instrumentation__.Notify(36094)
	}
	__antithesis_instrumentation__.Notify(36091)
	return c.Conn.Read(b[:64])
}

func appendPath(t *testing.T, s, add string) string {
	__antithesis_instrumentation__.Notify(36095)
	u, err := url.Parse(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(36097)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(36098)
	}
	__antithesis_instrumentation__.Notify(36096)
	u.Path = filepath.Join(u.Path, add)
	return u.String()
}

func storeFromURI(
	ctx context.Context,
	t *testing.T,
	uri string,
	clientFactory blobs.BlobClientFactory,
	user security.SQLUsername,
	ie sqlutil.InternalExecutor,
	kvDB *kv.DB,
	testSettings *cluster.Settings,
) cloud.ExternalStorage {
	__antithesis_instrumentation__.Notify(36099)
	conf, err := cloud.ExternalStorageConfFromURI(uri, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(36102)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(36103)
	}
	__antithesis_instrumentation__.Notify(36100)

	s, err := cloud.MakeExternalStorage(ctx, conf, base.ExternalIODirConfig{}, testSettings,
		clientFactory, ie, kvDB, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(36104)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(36105)
	}
	__antithesis_instrumentation__.Notify(36101)
	return s
}

func CheckExportStore(
	t *testing.T,
	storeURI string,
	skipSingleFile bool,
	user security.SQLUsername,
	ie sqlutil.InternalExecutor,
	kvDB *kv.DB,
	testSettings *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(36106)
	ioConf := base.ExternalIODirConfig{}
	ctx := context.Background()

	conf, err := cloud.ExternalStorageConfFromURI(storeURI, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(36115)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(36116)
	}
	__antithesis_instrumentation__.Notify(36107)

	clientFactory := blobs.TestBlobServiceClient(testSettings.ExternalIODir)
	s, err := cloud.MakeExternalStorage(ctx, conf, ioConf, testSettings, clientFactory, ie, kvDB, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(36117)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(36118)
	}
	__antithesis_instrumentation__.Notify(36108)
	defer s.Close()

	if readConf := s.Conf(); readConf != conf {
		__antithesis_instrumentation__.Notify(36119)
		t.Fatalf("conf does not roundtrip: started with %+v, got back %+v", conf, readConf)
	} else {
		__antithesis_instrumentation__.Notify(36120)
	}
	__antithesis_instrumentation__.Notify(36109)

	rng, _ := randutil.NewTestRand()

	t.Run("simple round trip", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(36121)
		sampleName := "somebytes"
		sampleBytes := "hello world"

		for i := 0; i < 10; i++ {
			__antithesis_instrumentation__.Notify(36122)
			name := fmt.Sprintf("%s-%d", sampleName, i)
			payload := []byte(strings.Repeat(sampleBytes, i))
			if err := cloud.WriteFile(ctx, s, name, bytes.NewReader(payload)); err != nil {
				__antithesis_instrumentation__.Notify(36128)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(36129)
			}
			__antithesis_instrumentation__.Notify(36123)

			if sz, err := s.Size(ctx, name); err != nil {
				__antithesis_instrumentation__.Notify(36130)
				t.Error(err)
			} else {
				__antithesis_instrumentation__.Notify(36131)
				if sz != int64(len(payload)) {
					__antithesis_instrumentation__.Notify(36132)
					t.Errorf("size mismatch, got %d, expected %d", sz, len(payload))
				} else {
					__antithesis_instrumentation__.Notify(36133)
				}
			}
			__antithesis_instrumentation__.Notify(36124)

			r, err := s.ReadFile(ctx, name)
			if err != nil {
				__antithesis_instrumentation__.Notify(36134)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(36135)
			}
			__antithesis_instrumentation__.Notify(36125)
			defer r.Close(ctx)

			res, err := ioctx.ReadAll(ctx, r)
			if err != nil {
				__antithesis_instrumentation__.Notify(36136)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(36137)
			}
			__antithesis_instrumentation__.Notify(36126)
			if !bytes.Equal(res, payload) {
				__antithesis_instrumentation__.Notify(36138)
				t.Fatalf("got %v expected %v", res, payload)
			} else {
				__antithesis_instrumentation__.Notify(36139)
			}
			__antithesis_instrumentation__.Notify(36127)

			require.NoError(t, s.Delete(ctx, name))
		}
	})
	__antithesis_instrumentation__.Notify(36110)

	t.Run("exceeds-4mb-chunk", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(36140)
		const size = 1024 * 1024 * 5
		testingContent := randutil.RandBytes(rng, size)
		testingFilename := "testing-123"

		if err := cloud.WriteFile(ctx, s, testingFilename, bytes.NewReader(testingContent)); err != nil {
			__antithesis_instrumentation__.Notify(36146)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36147)
		}
		__antithesis_instrumentation__.Notify(36141)

		res, err := s.ReadFile(ctx, testingFilename)
		if err != nil {
			__antithesis_instrumentation__.Notify(36148)
			t.Fatalf("Could not get reader for %s: %+v", testingFilename, err)
		} else {
			__antithesis_instrumentation__.Notify(36149)
		}
		__antithesis_instrumentation__.Notify(36142)
		defer res.Close(ctx)
		content, err := ioctx.ReadAll(ctx, res)
		if err != nil {
			__antithesis_instrumentation__.Notify(36150)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36151)
		}
		__antithesis_instrumentation__.Notify(36143)

		if !bytes.Equal(content, testingContent) {
			__antithesis_instrumentation__.Notify(36152)
			t.Fatalf("wrong content")
		} else {
			__antithesis_instrumentation__.Notify(36153)
		}
		__antithesis_instrumentation__.Notify(36144)

		t.Run("rand-readats", func(t *testing.T) {
			__antithesis_instrumentation__.Notify(36154)
			for i := 0; i < 10; i++ {
				__antithesis_instrumentation__.Notify(36155)
				t.Run("", func(t *testing.T) {
					__antithesis_instrumentation__.Notify(36156)
					byteReader := bytes.NewReader(testingContent)
					offset, length := rng.Int63n(size), rng.Intn(32*1024)
					t.Logf("read %d of file at %d", length, offset)
					reader, size, err := s.ReadFileAt(ctx, testingFilename, offset)
					require.NoError(t, err)
					defer reader.Close(ctx)
					require.Equal(t, int64(len(testingContent)), size)
					expected, got := make([]byte, length), make([]byte, length)
					_, err = byteReader.Seek(offset, io.SeekStart)
					require.NoError(t, err)

					expectedN, expectedErr := io.ReadFull(byteReader, expected)
					gotN, gotErr := io.ReadFull(ioctx.ReaderCtxAdapter(ctx, reader), got)
					require.Equal(t, expectedErr != nil, gotErr != nil, "%+v vs %+v", expectedErr, gotErr)
					require.Equal(t, expectedN, gotN)
					require.Equal(t, expected, got)
				})
			}
		})
		__antithesis_instrumentation__.Notify(36145)

		require.NoError(t, s.Delete(ctx, testingFilename))
	})
	__antithesis_instrumentation__.Notify(36111)
	if skipSingleFile {
		__antithesis_instrumentation__.Notify(36157)
		return
	} else {
		__antithesis_instrumentation__.Notify(36158)
	}
	__antithesis_instrumentation__.Notify(36112)
	t.Run("read-single-file-by-uri", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(36159)
		const testingFilename = "A"
		if err := cloud.WriteFile(ctx, s, testingFilename, bytes.NewReader([]byte("aaa"))); err != nil {
			__antithesis_instrumentation__.Notify(36164)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36165)
		}
		__antithesis_instrumentation__.Notify(36160)
		singleFile := storeFromURI(ctx, t, appendPath(t, storeURI, testingFilename), clientFactory,
			user, ie, kvDB, testSettings)
		defer singleFile.Close()

		res, err := singleFile.ReadFile(ctx, "")
		if err != nil {
			__antithesis_instrumentation__.Notify(36166)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36167)
		}
		__antithesis_instrumentation__.Notify(36161)
		defer res.Close(ctx)
		content, err := ioctx.ReadAll(ctx, res)
		if err != nil {
			__antithesis_instrumentation__.Notify(36168)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36169)
		}
		__antithesis_instrumentation__.Notify(36162)

		if !bytes.Equal(content, []byte("aaa")) {
			__antithesis_instrumentation__.Notify(36170)
			t.Fatalf("wrong content")
		} else {
			__antithesis_instrumentation__.Notify(36171)
		}
		__antithesis_instrumentation__.Notify(36163)
		require.NoError(t, s.Delete(ctx, testingFilename))
	})
	__antithesis_instrumentation__.Notify(36113)
	t.Run("write-single-file-by-uri", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(36172)
		const testingFilename = "B"
		singleFile := storeFromURI(ctx, t, appendPath(t, storeURI, testingFilename), clientFactory,
			user, ie, kvDB, testSettings)
		defer singleFile.Close()

		if err := cloud.WriteFile(ctx, singleFile, "", bytes.NewReader([]byte("bbb"))); err != nil {
			__antithesis_instrumentation__.Notify(36177)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36178)
		}
		__antithesis_instrumentation__.Notify(36173)

		res, err := s.ReadFile(ctx, testingFilename)
		if err != nil {
			__antithesis_instrumentation__.Notify(36179)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36180)
		}
		__antithesis_instrumentation__.Notify(36174)
		defer res.Close(ctx)
		content, err := ioctx.ReadAll(ctx, res)
		if err != nil {
			__antithesis_instrumentation__.Notify(36181)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36182)
		}
		__antithesis_instrumentation__.Notify(36175)

		if !bytes.Equal(content, []byte("bbb")) {
			__antithesis_instrumentation__.Notify(36183)
			t.Fatalf("wrong content")
		} else {
			__antithesis_instrumentation__.Notify(36184)
		}
		__antithesis_instrumentation__.Notify(36176)

		require.NoError(t, s.Delete(ctx, testingFilename))
	})
	__antithesis_instrumentation__.Notify(36114)

	t.Run("file-does-not-exist", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(36185)
		const testingFilename = "A"
		if err := cloud.WriteFile(ctx, s, testingFilename, bytes.NewReader([]byte("aaa"))); err != nil {
			__antithesis_instrumentation__.Notify(36190)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36191)
		}
		__antithesis_instrumentation__.Notify(36186)
		singleFile := storeFromURI(ctx, t, storeURI, clientFactory, user, ie, kvDB, testSettings)
		defer singleFile.Close()

		res, err := singleFile.ReadFile(ctx, testingFilename)
		if err != nil {
			__antithesis_instrumentation__.Notify(36192)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36193)
		}
		__antithesis_instrumentation__.Notify(36187)
		defer res.Close(ctx)
		content, err := ioctx.ReadAll(ctx, res)
		if err != nil {
			__antithesis_instrumentation__.Notify(36194)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36195)
		}
		__antithesis_instrumentation__.Notify(36188)

		if !bytes.Equal(content, []byte("aaa")) {
			__antithesis_instrumentation__.Notify(36196)
			t.Fatalf("wrong content")
		} else {
			__antithesis_instrumentation__.Notify(36197)
		}
		__antithesis_instrumentation__.Notify(36189)

		_, err = singleFile.ReadFile(ctx, "file_does_not_exist")
		require.Error(t, err)
		require.True(t, errors.Is(err, cloud.ErrFileDoesNotExist), "Expected a file does not exist error but returned %s")

		_, _, err = singleFile.ReadFileAt(ctx, "file_does_not_exist", 0)
		require.Error(t, err)
		require.True(t, errors.Is(err, cloud.ErrFileDoesNotExist), "Expected a file does not exist error but returned %s")

		_, _, err = singleFile.ReadFileAt(ctx, "file_does_not_exist", 24)
		require.Error(t, err)
		require.True(t, errors.Is(err, cloud.ErrFileDoesNotExist), "Expected a file does not exist error but returned %s")

		require.NoError(t, s.Delete(ctx, testingFilename))
	})
}

func CheckListFiles(
	t *testing.T,
	storeURI string,
	user security.SQLUsername,
	ie sqlutil.InternalExecutor,
	kvDB *kv.DB,
	testSettings *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(36198)
	CheckListFilesCanonical(t, storeURI, "", user, ie, kvDB, testSettings)
}

func CheckListFilesCanonical(
	t *testing.T,
	storeURI string,
	canonical string,
	user security.SQLUsername,
	ie sqlutil.InternalExecutor,
	kvDB *kv.DB,
	testSettings *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(36199)
	ctx := context.Background()
	dataLetterFiles := []string{"file/letters/dataA.csv", "file/letters/dataB.csv", "file/letters/dataC.csv"}
	dataNumberFiles := []string{"file/numbers/data1.csv", "file/numbers/data2.csv", "file/numbers/data3.csv"}
	letterFiles := []string{"file/abc/A.csv", "file/abc/B.csv", "file/abc/C.csv"}
	fileNames := append(dataLetterFiles, dataNumberFiles...)
	fileNames = append(fileNames, letterFiles...)
	sort.Strings(fileNames)

	clientFactory := blobs.TestBlobServiceClient(testSettings.ExternalIODir)
	for _, fileName := range fileNames {
		__antithesis_instrumentation__.Notify(36203)
		file := storeFromURI(ctx, t, storeURI, clientFactory, user, ie, kvDB, testSettings)
		if err := cloud.WriteFile(ctx, file, fileName, bytes.NewReader([]byte("bbb"))); err != nil {
			__antithesis_instrumentation__.Notify(36205)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36206)
		}
		__antithesis_instrumentation__.Notify(36204)
		_ = file.Close()
	}
	__antithesis_instrumentation__.Notify(36200)

	foreach := func(in []string, fn func(string) string) []string {
		__antithesis_instrumentation__.Notify(36207)
		out := make([]string, len(in))
		for i := range in {
			__antithesis_instrumentation__.Notify(36209)
			out[i] = fn(in[i])
		}
		__antithesis_instrumentation__.Notify(36208)
		return out
	}
	__antithesis_instrumentation__.Notify(36201)

	t.Run("List", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(36210)
		for _, tc := range []struct {
			name      string
			uri       string
			prefix    string
			delimiter string
			expected  []string
		}{
			{
				"root",
				storeURI,
				"",
				"",
				foreach(fileNames, func(s string) string { __antithesis_instrumentation__.Notify(36211); return "/" + s }),
			},
			{
				"file-slash-numbers-slash",
				storeURI,
				"file/numbers/",
				"",
				[]string{"data1.csv", "data2.csv", "data3.csv"},
			},
			{
				"root-slash",
				storeURI,
				"/",
				"",
				foreach(fileNames, func(s string) string { __antithesis_instrumentation__.Notify(36212); return s }),
			},
			{
				"file",
				storeURI,
				"file",
				"",
				foreach(fileNames, func(s string) string {
					__antithesis_instrumentation__.Notify(36213)
					return strings.TrimPrefix(s, "file")
				}),
			},
			{
				"file-slash",
				storeURI,
				"file/",
				"",
				foreach(fileNames, func(s string) string {
					__antithesis_instrumentation__.Notify(36214)
					return strings.TrimPrefix(s, "file/")
				}),
			},
			{
				"slash-f",
				storeURI,
				"/f",
				"",
				foreach(fileNames, func(s string) string { __antithesis_instrumentation__.Notify(36215); return strings.TrimPrefix(s, "f") }),
			},
			{
				"nothing",
				storeURI,
				"nothing",
				"",
				nil,
			},
			{
				"delim-slash-file-slash",
				storeURI,
				"file/",
				"/",
				[]string{"abc/", "letters/", "numbers/"},
			},
			{
				"delim-data",
				storeURI,
				"",
				"data",
				[]string{"/file/abc/A.csv", "/file/abc/B.csv", "/file/abc/C.csv", "/file/letters/data", "/file/numbers/data"},
			},
		} {
			__antithesis_instrumentation__.Notify(36216)
			t.Run(tc.name, func(t *testing.T) {
				__antithesis_instrumentation__.Notify(36217)
				s := storeFromURI(ctx, t, tc.uri, clientFactory, user, ie, kvDB, testSettings)
				var actual []string
				require.NoError(t, s.List(ctx, tc.prefix, tc.delimiter, func(f string) error {
					__antithesis_instrumentation__.Notify(36219)
					actual = append(actual, f)
					return nil
				}))
				__antithesis_instrumentation__.Notify(36218)
				sort.Strings(actual)
				require.Equal(t, tc.expected, actual)
			})
		}
	})
	__antithesis_instrumentation__.Notify(36202)

	for _, fileName := range fileNames {
		__antithesis_instrumentation__.Notify(36220)
		file := storeFromURI(ctx, t, storeURI, clientFactory, user, ie, kvDB, testSettings)
		if err := file.Delete(ctx, fileName); err != nil {
			__antithesis_instrumentation__.Notify(36222)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(36223)
		}
		__antithesis_instrumentation__.Notify(36221)
		_ = file.Close()
	}
}

func uploadData(
	t *testing.T,
	testSettings *cluster.Settings,
	rnd *rand.Rand,
	dest roachpb.ExternalStorage,
	basename string,
) ([]byte, func()) {
	__antithesis_instrumentation__.Notify(36224)
	data := randutil.RandBytes(rnd, 16<<20)
	ctx := context.Background()

	s, err := cloud.MakeExternalStorage(
		ctx, dest, base.ExternalIODirConfig{}, testSettings,
		nil, nil, nil, nil)
	require.NoError(t, err)
	require.NoError(t, cloud.WriteFile(ctx, s, basename, bytes.NewReader(data)))
	return data, func() {
		__antithesis_instrumentation__.Notify(36225)
		defer s.Close()
		_ = s.Delete(ctx, basename)
	}
}

func CheckAntagonisticRead(
	t *testing.T, conf roachpb.ExternalStorage, testSettings *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(36226)
	rnd, _ := randutil.NewTestRand()

	const basename = "test-antagonistic-read"
	data, cleanup := uploadData(t, testSettings, rnd, conf, basename)
	defer cleanup()

	failures := 0

	dialer := &antagonisticDialer{rnd: rnd, numRepeatFailures: &failures}

	transport := http.DefaultTransport.(*http.Transport)
	transport.DialContext =
		func(ctx context.Context, network, addr string) (net.Conn, error) {
			__antithesis_instrumentation__.Notify(36229)
			return dialer.DialContext(ctx, network, addr)
		}
	__antithesis_instrumentation__.Notify(36227)
	transport.DisableKeepAlives = true

	defer func() {
		__antithesis_instrumentation__.Notify(36230)
		transport.DialContext = nil
		transport.DisableKeepAlives = false
	}()
	__antithesis_instrumentation__.Notify(36228)

	ctx := context.Background()
	s, err := cloud.MakeExternalStorage(
		ctx, conf, base.ExternalIODirConfig{}, testSettings,
		nil, nil, nil, nil)
	require.NoError(t, err)
	defer s.Close()

	stream, err := s.ReadFile(ctx, basename)
	require.NoError(t, err)
	defer stream.Close(ctx)
	read, err := ioctx.ReadAll(ctx, stream)
	require.NoError(t, err)
	require.Equal(t, data, read)
}
