package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/errors"
	"github.com/xdg-go/scram"
	"github.com/xdg-go/stringprep"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

var BcryptCost = settings.RegisterIntSetting(
	settings.TenantWritable,
	BcryptCostSettingName,
	fmt.Sprintf(
		"the hashing cost to use when storing passwords supplied as cleartext by SQL clients "+
			"with the hashing method crdb-bcrypt (allowed range: %d-%d)",
		bcrypt.MinCost, bcrypt.MaxCost),

	10,
	func(i int64) error {
		__antithesis_instrumentation__.Notify(186724)
		if i < int64(bcrypt.MinCost) || func() bool {
			__antithesis_instrumentation__.Notify(186726)
			return i > int64(bcrypt.MaxCost) == true
		}() == true {
			__antithesis_instrumentation__.Notify(186727)
			return bcrypt.InvalidCostError(int(i))
		} else {
			__antithesis_instrumentation__.Notify(186728)
		}
		__antithesis_instrumentation__.Notify(186725)
		return nil
	}).WithPublic()

const BcryptCostSettingName = "server.user_login.password_hashes.default_cost.crdb_bcrypt"

var SCRAMCost = settings.RegisterIntSetting(
	settings.TenantWritable,
	SCRAMCostSettingName,
	fmt.Sprintf(
		"the hashing cost to use when storing passwords supplied as cleartext by SQL clients "+
			"with the hashing method scram-sha-256 (allowed range: %d-%d)",
		scramMinCost, scramMaxCost),

	119680,
	func(i int64) error {
		__antithesis_instrumentation__.Notify(186729)
		if i < scramMinCost || func() bool {
			__antithesis_instrumentation__.Notify(186731)
			return i > scramMaxCost == true
		}() == true {
			__antithesis_instrumentation__.Notify(186732)
			return errors.Newf("cost not in allowed range (%d,%d)", scramMinCost, scramMaxCost)
		} else {
			__antithesis_instrumentation__.Notify(186733)
		}
		__antithesis_instrumentation__.Notify(186730)
		return nil
	}).WithPublic()

const scramMinCost = 4096
const scramMaxCost = 240000000000

const SCRAMCostSettingName = "server.user_login.password_hashes.default_cost.scram_sha_256"

var ErrEmptyPassword = errors.New("empty passwords are not permitted")

var ErrPasswordTooShort = errors.New("password too short")

var ErrUnknownHashMethod = errors.New("unknown hash method")

type HashMethod int8

const (
	HashInvalidMethod HashMethod = 0

	HashMissingPassword HashMethod = 1

	HashBCrypt HashMethod = 2

	HashSCRAMSHA256 HashMethod = 3
)

func (h HashMethod) String() string {
	__antithesis_instrumentation__.Notify(186734)
	switch h {
	case HashInvalidMethod:
		__antithesis_instrumentation__.Notify(186735)
		return "<invalid>"
	case HashMissingPassword:
		__antithesis_instrumentation__.Notify(186736)
		return "<missing password>"
	case HashBCrypt:
		__antithesis_instrumentation__.Notify(186737)
		return "crdb-bcrypt"
	case HashSCRAMSHA256:
		__antithesis_instrumentation__.Notify(186738)
		return "scram-sha-256"
	default:
		__antithesis_instrumentation__.Notify(186739)
		panic(errors.AssertionFailedf("programming errof: unknown hash method %d", int(h)))
	}
}

type PasswordHash interface {
	fmt.Stringer

	Method() HashMethod

	Size() int

	compareWithCleartextPassword(ctx context.Context, cleartext string) (ok bool, err error)
}

var _ PasswordHash = emptyPassword{}
var _ PasswordHash = invalidHash(nil)
var _ PasswordHash = bcryptHash(nil)
var _ PasswordHash = (*scramHash)(nil)

type emptyPassword struct{}

func (e emptyPassword) String() string {
	__antithesis_instrumentation__.Notify(186740)
	return "<missing>"
}

func (e emptyPassword) Method() HashMethod {
	__antithesis_instrumentation__.Notify(186741)
	return HashMissingPassword
}

func (e emptyPassword) Size() int { __antithesis_instrumentation__.Notify(186742); return 0 }

func (e emptyPassword) compareWithCleartextPassword(
	ctx context.Context, cleartext string,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(186743)
	return false, nil
}

var MissingPasswordHash PasswordHash = emptyPassword{}

type invalidHash []byte

func (n invalidHash) String() string { __antithesis_instrumentation__.Notify(186744); return string(n) }

func (n invalidHash) Method() HashMethod {
	__antithesis_instrumentation__.Notify(186745)
	return HashInvalidMethod
}

func (n invalidHash) Size() int { __antithesis_instrumentation__.Notify(186746); return len(n) }

func (n invalidHash) compareWithCleartextPassword(
	ctx context.Context, cleartext string,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(186747)
	return false, nil
}

type bcryptHash []byte

func (b bcryptHash) String() string { __antithesis_instrumentation__.Notify(186748); return string(b) }

func (b bcryptHash) Method() HashMethod {
	__antithesis_instrumentation__.Notify(186749)
	return HashBCrypt
}

func (b bcryptHash) Size() int { __antithesis_instrumentation__.Notify(186750); return len(b) }

type scramHash struct {
	bytes   []byte
	decoded scram.StoredCredentials
}

func (s *scramHash) String() string {
	__antithesis_instrumentation__.Notify(186751)
	return string(s.bytes)
}

func (s *scramHash) Method() HashMethod {
	__antithesis_instrumentation__.Notify(186752)
	return HashSCRAMSHA256
}

func (s *scramHash) Size() int {
	__antithesis_instrumentation__.Notify(186753)
	return int(unsafe.Sizeof(*s)) + len(s.bytes) + len(s.decoded.Salt) + len(s.decoded.StoredKey) + len(s.decoded.ServerKey)
}

func GetSCRAMStoredCredentials(hash PasswordHash) (ok bool, creds scram.StoredCredentials) {
	__antithesis_instrumentation__.Notify(186754)
	h, ok := hash.(*scramHash)
	if ok {
		__antithesis_instrumentation__.Notify(186756)
		return ok, h.decoded
	} else {
		__antithesis_instrumentation__.Notify(186757)
	}
	__antithesis_instrumentation__.Notify(186755)
	return false, creds
}

func LoadPasswordHash(ctx context.Context, storedHash []byte) (res PasswordHash) {
	__antithesis_instrumentation__.Notify(186758)
	res = invalidHash(storedHash)
	if len(storedHash) == 0 {
		__antithesis_instrumentation__.Notify(186762)
		return emptyPassword{}
	} else {
		__antithesis_instrumentation__.Notify(186763)
	}
	__antithesis_instrumentation__.Notify(186759)
	if isBcryptHash(storedHash, false) {
		__antithesis_instrumentation__.Notify(186764)
		return bcryptHash(storedHash)
	} else {
		__antithesis_instrumentation__.Notify(186765)
	}
	__antithesis_instrumentation__.Notify(186760)
	if ok, parts := isSCRAMHash(storedHash); ok {
		__antithesis_instrumentation__.Notify(186766)
		return makeSCRAMHash(storedHash, parts, res)
	} else {
		__antithesis_instrumentation__.Notify(186767)
	}
	__antithesis_instrumentation__.Notify(186761)

	return res
}

var sha256NewSum = sha256.New().Sum(nil)

func appendEmptySha256(password string) []byte {
	__antithesis_instrumentation__.Notify(186768)

	return append([]byte(password), sha256NewSum...)
}

func CompareHashAndCleartextPassword(
	ctx context.Context, hashedPassword PasswordHash, password string,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(186769)
	return hashedPassword.compareWithCleartextPassword(ctx, password)
}

func (b bcryptHash) compareWithCleartextPassword(
	ctx context.Context, cleartext string,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(186770)
	sem := getExpensiveHashComputeSem(ctx)
	alloc, err := sem.Acquire(ctx, 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(186773)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(186774)
	}
	__antithesis_instrumentation__.Notify(186771)
	defer alloc.Release()

	err = bcrypt.CompareHashAndPassword([]byte(b), appendEmptySha256(cleartext))
	if err != nil {
		__antithesis_instrumentation__.Notify(186775)
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			__antithesis_instrumentation__.Notify(186777)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(186778)
		}
		__antithesis_instrumentation__.Notify(186776)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(186779)
	}
	__antithesis_instrumentation__.Notify(186772)
	return true, nil
}

func (s *scramHash) compareWithCleartextPassword(
	ctx context.Context, cleartext string,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(186780)
	sem := getExpensiveHashComputeSem(ctx)
	alloc, err := sem.Acquire(ctx, 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(186783)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(186784)
	}
	__antithesis_instrumentation__.Notify(186781)
	defer alloc.Release()

	prepared, err := stringprep.SASLprep.Prepare(cleartext)
	if err != nil {
		__antithesis_instrumentation__.Notify(186785)

		prepared = cleartext
	} else {
		__antithesis_instrumentation__.Notify(186786)
	}
	__antithesis_instrumentation__.Notify(186782)

	saltedPassword := pbkdf2.Key([]byte(prepared), []byte(s.decoded.Salt), s.decoded.Iters, sha256.Size, sha256.New)

	serverKey := computeHMAC(scram.SHA256, saltedPassword, []byte("Server Key"))
	return bytes.Equal(serverKey, s.decoded.ServerKey), nil
}

func computeHMAC(hg scram.HashGeneratorFcn, key, data []byte) []byte {
	__antithesis_instrumentation__.Notify(186787)
	mac := hmac.New(hg, key)
	mac.Write(data)
	return mac.Sum(nil)
}

var PasswordHashMethod = settings.RegisterEnumSetting(
	settings.TenantWritable,
	"server.user_login.password_encryption",
	"which hash method to use to encode cleartext passwords passed via ALTER/CREATE USER/ROLE WITH PASSWORD",

	HashBCrypt.String(),
	map[int64]string{
		int64(HashBCrypt):      HashBCrypt.String(),
		int64(HashSCRAMSHA256): HashSCRAMSHA256.String(),
	},
).WithPublic()

func hasClusterVersion(
	ctx context.Context, values *settings.Values, versionkey clusterversion.Key,
) bool {
	__antithesis_instrumentation__.Notify(186788)
	var vh clusterversion.Handle
	if values != nil {
		__antithesis_instrumentation__.Notify(186790)
		vh = values.Opaque().(clusterversion.Handle)
	} else {
		__antithesis_instrumentation__.Notify(186791)
	}
	__antithesis_instrumentation__.Notify(186789)
	return vh != nil && func() bool {
		__antithesis_instrumentation__.Notify(186792)
		return vh.IsActive(ctx, versionkey) == true
	}() == true
}

func GetConfiguredPasswordHashMethod(ctx context.Context, sv *settings.Values) (method HashMethod) {
	__antithesis_instrumentation__.Notify(186793)
	method = HashMethod(PasswordHashMethod.Get(sv))
	if method == HashSCRAMSHA256 && func() bool {
		__antithesis_instrumentation__.Notify(186795)
		return !hasClusterVersion(ctx, sv, clusterversion.SCRAMAuthentication) == true
	}() == true {
		__antithesis_instrumentation__.Notify(186796)

		method = HashBCrypt
	} else {
		__antithesis_instrumentation__.Notify(186797)
	}
	__antithesis_instrumentation__.Notify(186794)
	return method
}

func HashPassword(ctx context.Context, sv *settings.Values, password string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(186798)
	method := GetConfiguredPasswordHashMethod(ctx, sv)
	switch method {
	case HashBCrypt:
		__antithesis_instrumentation__.Notify(186799)
		sem := getExpensiveHashComputeSem(ctx)
		alloc, err := sem.Acquire(ctx, 1)
		if err != nil {
			__antithesis_instrumentation__.Notify(186803)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(186804)
		}
		__antithesis_instrumentation__.Notify(186800)
		defer alloc.Release()
		return bcrypt.GenerateFromPassword(appendEmptySha256(password), int(BcryptCost.Get(sv)))

	case HashSCRAMSHA256:
		__antithesis_instrumentation__.Notify(186801)
		cost := int(SCRAMCost.Get(sv))
		return hashPasswordUsingSCRAM(ctx, cost, password)

	default:
		__antithesis_instrumentation__.Notify(186802)
		return nil, errors.Newf("unsupported hash method: %v", method)
	}
}

func hashPasswordUsingSCRAM(ctx context.Context, cost int, cleartext string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(186805)
	prepared, err := stringprep.SASLprep.Prepare(cleartext)
	if err != nil {
		__antithesis_instrumentation__.Notify(186810)

		prepared = cleartext
	} else {
		__antithesis_instrumentation__.Notify(186811)
	}
	__antithesis_instrumentation__.Notify(186806)

	client, err := scram.SHA256.NewClientUnprepped("", prepared, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(186812)
		return nil, errors.AssertionFailedf("programming error: client construction should never fail")
	} else {
		__antithesis_instrumentation__.Notify(186813)
	}
	__antithesis_instrumentation__.Notify(186807)

	const scramSaltSize = 16
	salt := make([]byte, scramSaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		__antithesis_instrumentation__.Notify(186814)
		return nil, errors.Wrap(err, "generating random salt")
	} else {
		__antithesis_instrumentation__.Notify(186815)
	}
	__antithesis_instrumentation__.Notify(186808)

	sem := getExpensiveHashComputeSem(ctx)
	alloc, err := sem.Acquire(ctx, 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(186816)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186817)
	}
	__antithesis_instrumentation__.Notify(186809)
	defer alloc.Release()

	creds := client.GetStoredCredentials(scram.KeyFactors{Iters: cost, Salt: string(salt)})

	return encodeScramHash(salt, creds), nil
}

func encodeScramHash(saltBytes []byte, sc scram.StoredCredentials) []byte {
	__antithesis_instrumentation__.Notify(186818)
	b64enc := base64.StdEncoding
	saltLen := b64enc.EncodedLen(len(saltBytes))
	storedKeyLen := b64enc.EncodedLen(len(sc.StoredKey))
	serverKeyLen := b64enc.EncodedLen(len(sc.ServerKey))

	res := make([]byte, 0, len(scramPrefix)+1+4+1+saltLen+1+storedKeyLen+1+serverKeyLen)
	res = append(res, scramPrefix...)
	res = append(res, '$')
	res = strconv.AppendInt(res, int64(sc.Iters), 10)
	res = append(res, ':')
	res = append(res, make([]byte, saltLen)...)
	b64enc.Encode(res[len(res)-saltLen:], saltBytes)
	res = append(res, '$')
	res = append(res, make([]byte, storedKeyLen)...)
	b64enc.Encode(res[len(res)-storedKeyLen:], sc.StoredKey)
	res = append(res, ':')
	res = append(res, make([]byte, serverKeyLen)...)
	b64enc.Encode(res[len(res)-serverKeyLen:], sc.ServerKey)
	return res
}

var AutoDetectPasswordHashes = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"server.user_login.store_client_pre_hashed_passwords.enabled",
	"whether the server accepts to store passwords pre-hashed by clients",
	true,
)

const crdbBcryptPrefix = "CRDB-BCRYPT"

var bcryptHashRe = regexp.MustCompile(`^(` + crdbBcryptPrefix + `)?\$\d[a-z]?\$\d\d\$[0-9A-Za-z\./]{22}[0-9A-Za-z\./]+$`)

func isBcryptHash(inputPassword []byte, strict bool) bool {
	__antithesis_instrumentation__.Notify(186819)
	if !bcryptHashRe.Match(inputPassword) {
		__antithesis_instrumentation__.Notify(186822)
		return false
	} else {
		__antithesis_instrumentation__.Notify(186823)
	}
	__antithesis_instrumentation__.Notify(186820)
	if strict && func() bool {
		__antithesis_instrumentation__.Notify(186824)
		return !bytes.HasPrefix(inputPassword, []byte(crdbBcryptPrefix+`$`)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(186825)
		return false
	} else {
		__antithesis_instrumentation__.Notify(186826)
	}
	__antithesis_instrumentation__.Notify(186821)
	return true
}

func checkBcryptHash(inputPassword []byte) (ok bool, hashedPassword []byte, err error) {
	__antithesis_instrumentation__.Notify(186827)
	if !isBcryptHash(inputPassword, true) {
		__antithesis_instrumentation__.Notify(186829)
		return false, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(186830)
	}
	__antithesis_instrumentation__.Notify(186828)

	hashedPassword = inputPassword[len(crdbBcryptPrefix):]

	_, err = bcrypt.Cost(hashedPassword)
	return true, hashedPassword, err
}

const scramPrefix = "SCRAM-SHA-256"

var scramHashRe = regexp.MustCompile(`^` + scramPrefix + `\$(\d+):([A-Za-z0-9+/]+=*)\$([A-Za-z0-9+/]{43}=):([A-Za-z0-9+/]{43}=)$`)

type scramParts [][]byte

func (sp scramParts) getIters() (int, error) {
	__antithesis_instrumentation__.Notify(186831)
	return strconv.Atoi(string(sp[1]))
}

func (sp scramParts) getSalt() ([]byte, error) {
	__antithesis_instrumentation__.Notify(186832)
	return base64.StdEncoding.DecodeString(string(sp[2]))
}

func (sp scramParts) getStoredKey() ([]byte, error) {
	__antithesis_instrumentation__.Notify(186833)
	return base64.StdEncoding.DecodeString(string(sp[3]))
}

func (sp scramParts) getServerKey() ([]byte, error) {
	__antithesis_instrumentation__.Notify(186834)
	return base64.StdEncoding.DecodeString(string(sp[4]))
}

func isSCRAMHash(inputPassword []byte) (bool, scramParts) {
	__antithesis_instrumentation__.Notify(186835)
	parts := scramHashRe.FindSubmatch(inputPassword)
	if parts == nil {
		__antithesis_instrumentation__.Notify(186838)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(186839)
	}
	__antithesis_instrumentation__.Notify(186836)
	if len(parts) != 5 {
		__antithesis_instrumentation__.Notify(186840)
		panic(errors.AssertionFailedf("programming error: scramParts type must have same length as regexp groups"))
	} else {
		__antithesis_instrumentation__.Notify(186841)
	}
	__antithesis_instrumentation__.Notify(186837)
	return true, scramParts(parts)
}

func checkSCRAMHash(inputPassword []byte) (ok bool, hashedPassword []byte, err error) {
	__antithesis_instrumentation__.Notify(186842)
	ok, parts := isSCRAMHash(inputPassword)
	if !ok {
		__antithesis_instrumentation__.Notify(186846)
		return false, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(186847)
	}
	__antithesis_instrumentation__.Notify(186843)
	iters, err := strconv.ParseInt(string(parts[1]), 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(186848)
		return true, nil, errors.Wrap(err, "invalid scram-sha-256 iteration count")
	} else {
		__antithesis_instrumentation__.Notify(186849)
	}
	__antithesis_instrumentation__.Notify(186844)

	if iters < scramMinCost || func() bool {
		__antithesis_instrumentation__.Notify(186850)
		return iters > scramMaxCost == true
	}() == true {
		__antithesis_instrumentation__.Notify(186851)
		return true, nil, errors.Newf("scram-sha-256 iteration count not in allowed range (%d,%d)", scramMinCost, scramMaxCost)
	} else {
		__antithesis_instrumentation__.Notify(186852)
	}
	__antithesis_instrumentation__.Notify(186845)
	return true, inputPassword, nil
}

func makeSCRAMHash(storedHash []byte, parts scramParts, invalidHash PasswordHash) PasswordHash {
	__antithesis_instrumentation__.Notify(186853)
	iters, err := parts.getIters()
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(186858)
		return iters < scramMinCost == true
	}() == true {
		__antithesis_instrumentation__.Notify(186859)
		return invalidHash
	} else {
		__antithesis_instrumentation__.Notify(186860)
	}
	__antithesis_instrumentation__.Notify(186854)
	salt, err := parts.getSalt()
	if err != nil {
		__antithesis_instrumentation__.Notify(186861)
		return invalidHash
	} else {
		__antithesis_instrumentation__.Notify(186862)
	}
	__antithesis_instrumentation__.Notify(186855)
	storedKey, err := parts.getStoredKey()
	if err != nil {
		__antithesis_instrumentation__.Notify(186863)
		return invalidHash
	} else {
		__antithesis_instrumentation__.Notify(186864)
	}
	__antithesis_instrumentation__.Notify(186856)
	serverKey, err := parts.getServerKey()
	if err != nil {
		__antithesis_instrumentation__.Notify(186865)
		return invalidHash
	} else {
		__antithesis_instrumentation__.Notify(186866)
	}
	__antithesis_instrumentation__.Notify(186857)
	return &scramHash{
		bytes: storedHash,
		decoded: scram.StoredCredentials{
			KeyFactors: scram.KeyFactors{
				Salt:  string(salt),
				Iters: iters,
			},
			StoredKey: storedKey,
			ServerKey: serverKey,
		},
	}
}

func isMD5Hash(hashedPassword []byte) bool {
	__antithesis_instrumentation__.Notify(186867)

	return bytes.HasPrefix(hashedPassword, []byte("md5")) && func() bool {
		__antithesis_instrumentation__.Notify(186868)
		return len(hashedPassword) == 35 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(186869)
		return len(bytes.Trim(hashedPassword[3:], "0123456789abcdef")) == 0 == true
	}() == true
}

func CheckPasswordHashValidity(
	ctx context.Context, inputPassword []byte,
) (
	isPreHashed, supportedScheme bool,
	issueNum int,
	schemeName string,
	hashedPassword []byte,
	err error,
) {
	__antithesis_instrumentation__.Notify(186870)
	if ok, hashedPassword, err := checkBcryptHash(inputPassword); ok {
		__antithesis_instrumentation__.Notify(186874)
		return true, true, 0, "crdb-bcrypt", hashedPassword, err
	} else {
		__antithesis_instrumentation__.Notify(186875)
	}
	__antithesis_instrumentation__.Notify(186871)
	if ok, hashedPassword, err := checkSCRAMHash(inputPassword); ok {
		__antithesis_instrumentation__.Notify(186876)
		return true, true, 0, "scram-sha-256", hashedPassword, err
	} else {
		__antithesis_instrumentation__.Notify(186877)
	}
	__antithesis_instrumentation__.Notify(186872)
	if isMD5Hash(inputPassword) {
		__antithesis_instrumentation__.Notify(186878)

		return true, false, 73337, "md5", inputPassword, nil
	} else {
		__antithesis_instrumentation__.Notify(186879)
	}
	__antithesis_instrumentation__.Notify(186873)

	return false, false, 0, "", inputPassword, nil
}

var MinPasswordLength = settings.RegisterIntSetting(
	settings.TenantWritable,
	"server.user_login.min_password_length",
	"the minimum length accepted for passwords set in cleartext via SQL. "+
		"Note that a value lower than 1 is ignored: passwords cannot be empty in any case.",
	1,
	settings.NonNegativeInt,
).WithPublic()

var expensiveHashComputeSemOnce struct {
	sem  *quotapool.IntPool
	once sync.Once
}

var envMaxHashComputeConcurrency = envutil.EnvOrDefaultInt("COCKROACH_MAX_PW_HASH_COMPUTE_CONCURRENCY", 0)

func getExpensiveHashComputeSem(ctx context.Context) *quotapool.IntPool {
	__antithesis_instrumentation__.Notify(186880)
	expensiveHashComputeSemOnce.once.Do(func() {
		__antithesis_instrumentation__.Notify(186882)
		var n int
		if envMaxHashComputeConcurrency >= 1 {
			__antithesis_instrumentation__.Notify(186885)

			n = envMaxHashComputeConcurrency
		} else {
			__antithesis_instrumentation__.Notify(186886)

			n = runtime.GOMAXPROCS(-1) / 8
		}
		__antithesis_instrumentation__.Notify(186883)
		if n < 1 {
			__antithesis_instrumentation__.Notify(186887)
			n = 1
		} else {
			__antithesis_instrumentation__.Notify(186888)
		}
		__antithesis_instrumentation__.Notify(186884)
		log.VInfof(ctx, 1, "configured maximum hashing concurrency: %d", n)
		expensiveHashComputeSemOnce.sem = quotapool.NewIntPool("password_hashes", uint64(n))
	})
	__antithesis_instrumentation__.Notify(186881)
	return expensiveHashComputeSemOnce.sem
}

var bcryptCostToSCRAMIterCount = []int64{
	0,
	0,
	0,
	0,
	4096,
	9420,
	12977,
	20090,
	34318,
	62772,
	119680,
	233497,
	461131,
	916398,
	1826932,
	3648001,
	7290139,
	14574415,
	29142967,
	58280072,
	116554280,
	233102696,
	466199529,
	932393195,
	1864780528,
	3729555192,
	7459104520,
	14918203177,
	29836400491,
	59672795119,
	119345584374,
	238691162884,
}

var autoUpgradePasswordHashes = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"server.user_login.upgrade_bcrypt_stored_passwords_to_scram.enabled",
	"whether to automatically re-encode stored passwords using crdb-bcrypt to scram-sha-256",

	false,
).WithPublic()

func MaybeUpgradePasswordHash(
	ctx context.Context, sv *settings.Values, cleartext string, hashed PasswordHash,
) (converted bool, prevHashBytes, newHashBytes []byte, newMethod string, err error) {
	__antithesis_instrumentation__.Notify(186889)
	bh, isBcrypt := hashed.(bcryptHash)

	if !isBcrypt || func() bool {
		__antithesis_instrumentation__.Notify(186897)
		return !autoUpgradePasswordHashes.Get(sv) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(186898)
		return GetConfiguredPasswordHashMethod(ctx, sv) != HashSCRAMSHA256 == true
	}() == true {
		__antithesis_instrumentation__.Notify(186899)

		return false, nil, nil, "", nil
	} else {
		__antithesis_instrumentation__.Notify(186900)
	}
	__antithesis_instrumentation__.Notify(186890)

	bcryptCost, err := bcrypt.Cost([]byte(bh))
	if err != nil {
		__antithesis_instrumentation__.Notify(186901)

		return false, nil, nil, "", errors.NewAssertionErrorWithWrappedErrf(err, "programming error: authn succeeded but invalid bcrypt hash")
	} else {
		__antithesis_instrumentation__.Notify(186902)
	}
	__antithesis_instrumentation__.Notify(186891)
	if bcryptCost < 0 || func() bool {
		__antithesis_instrumentation__.Notify(186903)
		return bcryptCost >= len(bcryptCostToSCRAMIterCount) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(186904)
		return bcryptCostToSCRAMIterCount[bcryptCost] == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(186905)

		return false, nil, nil, "", errors.AssertionFailedf("unexpected: bcrypt cost %d is out of bounds or has no mapping", bcryptCost)
	} else {
		__antithesis_instrumentation__.Notify(186906)
	}
	__antithesis_instrumentation__.Notify(186892)

	scramIterCount := bcryptCostToSCRAMIterCount[bcryptCost]
	if scramIterCount > math.MaxInt {
		__antithesis_instrumentation__.Notify(186907)

		return false, nil, nil, "", nil
	} else {
		__antithesis_instrumentation__.Notify(186908)
	}
	__antithesis_instrumentation__.Notify(186893)

	if bcryptCost > 10 {
		__antithesis_instrumentation__.Notify(186909)

		log.Infof(ctx, "hash conversion: computing a SCRAM hash with iteration count %d (from bcrypt cost %d)", scramIterCount, bcryptCost)
	} else {
		__antithesis_instrumentation__.Notify(186910)
	}
	__antithesis_instrumentation__.Notify(186894)

	rawHash, err := hashPasswordUsingSCRAM(ctx, int(scramIterCount), cleartext)
	if err != nil {
		__antithesis_instrumentation__.Notify(186911)

		return false, nil, nil, "", err
	} else {
		__antithesis_instrumentation__.Notify(186912)
	}
	__antithesis_instrumentation__.Notify(186895)

	expectedMethod := HashSCRAMSHA256
	newHash := LoadPasswordHash(ctx, rawHash)
	if newHash.Method() != expectedMethod {
		__antithesis_instrumentation__.Notify(186913)

		log.Errorf(ctx, "unexpected hash contents during bcrypt->scram conversion: %T %+v -> %T %+v", hashed, hashed, newHash, newHash)

		return false, nil, nil, "", errors.AssertionFailedf("programming error: re-hash failed to produce SCRAM hash, produced %T instead", newHash)
	} else {
		__antithesis_instrumentation__.Notify(186914)
	}
	__antithesis_instrumentation__.Notify(186896)
	return true, []byte(bh), rawHash, expectedMethod.String(), nil
}
