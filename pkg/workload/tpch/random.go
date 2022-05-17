package tpch

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"strconv"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/workload/faker"
	"golang.org/x/exp/rand"
)

const alphanumericLen64 = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890, `

func randInt(rng *rand.Rand, x, y int) int {
	__antithesis_instrumentation__.Notify(698783)
	return rng.Intn(y-x+1) + x
}

func randFloat(rng *rand.Rand, x, y, shift int) float32 {
	__antithesis_instrumentation__.Notify(698784)
	return float32(randInt(rng, x, y)) / float32(shift)
}

type textPool interface {
	randString(rng *rand.Rand, minLen, maxLen int) []byte
}

type fakeTextPool struct {
	seed uint64
	once struct {
		sync.Once
		buf []byte
	}
}

func (p *fakeTextPool) randString(rng *rand.Rand, minLen, maxLen int) []byte {
	__antithesis_instrumentation__.Notify(698785)
	const fakeTextPoolSize = 1 << 20
	p.once.Do(func() {
		__antithesis_instrumentation__.Notify(698787)
		bufRng := rand.New(rand.NewSource(p.seed))
		f := faker.NewFaker()

		buf := bytes.NewBuffer(make([]byte, 0, fakeTextPoolSize+1024))
		for buf.Len() < fakeTextPoolSize {
			__antithesis_instrumentation__.Notify(698789)
			buf.WriteString(f.Paragraph(bufRng))
			buf.WriteString(` `)
		}
		__antithesis_instrumentation__.Notify(698788)
		p.once.buf = buf.Bytes()[:fakeTextPoolSize:fakeTextPoolSize]
	})
	__antithesis_instrumentation__.Notify(698786)
	start := rng.Intn(len(p.once.buf) - maxLen)
	end := start + rng.Intn(maxLen-minLen) + minLen
	return p.once.buf[start:end]
}

func randVString(rng *rand.Rand, a *bufalloc.ByteAllocator, minLen, maxLen int) []byte {
	__antithesis_instrumentation__.Notify(698790)
	var buf []byte
	*a, buf = a.Alloc(randInt(rng, minLen, maxLen), 0)
	for i := range buf {
		__antithesis_instrumentation__.Notify(698792)
		buf[i] = alphanumericLen64[rng.Intn(len(alphanumericLen64))]
	}
	__antithesis_instrumentation__.Notify(698791)
	return buf
}

func randPhone(rng *rand.Rand, a *bufalloc.ByteAllocator, nationKey int16) []byte {
	__antithesis_instrumentation__.Notify(698793)
	var buf []byte
	*a, buf = a.Alloc(15, 0)
	buf = buf[:0]

	countryCode := nationKey + 10
	localNumber1 := randInt(rng, 100, 999)
	localNumber2 := randInt(rng, 100, 999)
	localNumber3 := randInt(rng, 1000, 9999)
	buf = strconv.AppendInt(buf, int64(countryCode), 10)
	buf = append(buf, '-')
	buf = strconv.AppendInt(buf, int64(localNumber1), 10)
	buf = append(buf, '-')
	buf = strconv.AppendInt(buf, int64(localNumber2), 10)
	buf = append(buf, '-')
	buf = strconv.AppendInt(buf, int64(localNumber3), 10)
	return buf
}

var randPartNames = [...]string{
	"almond", "antique", "aquamarine", "azure", "beige", "bisque", "black", "blanched", "blue",
	"blush", "brown", "burlywood", "burnished", "chartreuse", "chiffon", "chocolate", "coral",
	"cornflower", "cornsilk", "cream", "cyan", "dark", "deep", "dim", "dodger", "drab", "firebrick",
	"floral", "forest", "frosted", "gainsboro", "ghost", "goldenrod", "green", "grey", "honeydew",
	"hot", "indian", "ivory", "khaki", "lace", "lavender", "lawn", "lemon", "light", "lime", "linen",
	"magenta", "maroon", "medium", "metallic", "midnight", "mint", "misty", "moccasin", "navajo",
	"navy", "olive", "orange", "orchid", "pale", "papaya", "peach", "peru", "pink", "plum", "powder",
	"puff", "purple", "red", "rose", "rosy", "royal", "saddle", "salmon", "sandy", "seashell",
	"sienna", "sky", "slate", "smoke", "snow", "spring", "steel", "tan", "thistle", "tomato",
	"turquoise", "violet", "wheat", "white", "yellow",
}

const maxPartNameLen = 10
const nPartNames = 5

func randPartName(rng *rand.Rand, namePerm []int, a *bufalloc.ByteAllocator) []byte {
	__antithesis_instrumentation__.Notify(698794)

	for i := 0; i < nPartNames; i++ {
		__antithesis_instrumentation__.Notify(698797)
		j := rng.Intn(i + 1)
		namePerm[i] = namePerm[j]
		namePerm[j] = i
	}
	__antithesis_instrumentation__.Notify(698795)
	var buf []byte
	*a, buf = a.Alloc(maxPartNameLen*nPartNames+nPartNames, 0)
	buf = buf[:0]
	for i := 0; i < nPartNames; i++ {
		__antithesis_instrumentation__.Notify(698798)
		if i != 0 {
			__antithesis_instrumentation__.Notify(698800)
			buf = append(buf, byte(' '))
		} else {
			__antithesis_instrumentation__.Notify(698801)
		}
		__antithesis_instrumentation__.Notify(698799)
		buf = append(buf, randPartNames[namePerm[i]]...)
	}
	__antithesis_instrumentation__.Notify(698796)
	return buf
}

const manufacturerString = "Manufacturer#"

func randMfgr(rng *rand.Rand, a *bufalloc.ByteAllocator) (byte, []byte) {
	__antithesis_instrumentation__.Notify(698802)
	var buf []byte
	*a, buf = a.Alloc(len(manufacturerString)+1, 0)

	copy(buf, manufacturerString)
	m := byte(rng.Intn(5) + '1')
	buf[len(buf)-1] = m
	return m, buf
}

const brandString = "Brand#"

func randBrand(rng *rand.Rand, a *bufalloc.ByteAllocator, m byte) []byte {
	__antithesis_instrumentation__.Notify(698803)
	var buf []byte
	*a, buf = a.Alloc(len(brandString)+2, 0)

	copy(buf, brandString)
	n := byte(rng.Intn(5) + '1')
	buf[len(buf)-2] = m
	buf[len(buf)-1] = n
	return buf
}

const clerkString = "Clerk#"

func randClerk(rng *rand.Rand, a *bufalloc.ByteAllocator, scaleFactor int) []byte {
	__antithesis_instrumentation__.Notify(698804)
	var buf []byte
	*a, buf = a.Alloc(len(clerkString)+9, 0)
	copy(buf, clerkString)
	ninePaddedInt(buf[len(clerkString):], int64(randInt(rng, 1, scaleFactor*1000)))
	return buf
}

const supplierString = "Supplier#"

func supplierName(a *bufalloc.ByteAllocator, suppKey int64) []byte {
	__antithesis_instrumentation__.Notify(698805)
	var buf []byte
	*a, buf = a.Alloc(len(supplierString)+9, 0)
	copy(buf, supplierString)
	ninePaddedInt(buf[len(supplierString):], suppKey)
	return buf
}

const customerString = "Customer#"

func customerName(a *bufalloc.ByteAllocator, custKey int64) []byte {
	__antithesis_instrumentation__.Notify(698806)
	var buf []byte
	*a, buf = a.Alloc(len(customerString)+9, 0)
	copy(buf, customerString)
	ninePaddedInt(buf[len(customerString):], custKey)
	return buf
}

const ninePadding = `000000000`

func ninePaddedInt(buf []byte, x int64) {
	__antithesis_instrumentation__.Notify(698807)
	buf = buf[:len(ninePadding)]
	intLen := len(strconv.AppendInt(buf[:0], x, 10))
	numZeros := len(ninePadding) - intLen
	copy(buf[numZeros:], buf[:intLen])
	copy(buf[:numZeros], ninePadding[:numZeros])
}

func randSyllables(
	rng *rand.Rand, a *bufalloc.ByteAllocator, maxLen int, syllables [][]string,
) []byte {
	__antithesis_instrumentation__.Notify(698808)
	var buf []byte
	*a, buf = a.Alloc(maxLen, 0)
	buf = buf[:0]

	for i, syl := range syllables {
		__antithesis_instrumentation__.Notify(698810)
		if i != 0 {
			__antithesis_instrumentation__.Notify(698812)
			buf = append(buf, ' ')
		} else {
			__antithesis_instrumentation__.Notify(698813)
		}
		__antithesis_instrumentation__.Notify(698811)
		buf = append(buf, syl[rng.Intn(len(syl))]...)
	}
	__antithesis_instrumentation__.Notify(698809)
	return buf
}

var typeSyllables = [][]string{
	{"STANDARD", "SMALL", "MEDIUM", "LARGE", "ECONOMY", "PROMO"},
	{"ANODIZED", "BURNISHED", "PLATED", "POLISHED", "BRUSHED"},
	{"TIN", "NICKEL", "BRASS", "STEEL", "COPPER"},
}

const maxTypeLen = 25

func randType(rng *rand.Rand, a *bufalloc.ByteAllocator) []byte {
	__antithesis_instrumentation__.Notify(698814)
	return randSyllables(rng, a, maxTypeLen, typeSyllables)
}

var containerSyllables = [][]string{
	{"SM", "MED", "JUMBO", "WRAP"},
	{"BOX", "BAG", "JAR", "PKG", "PACK", "CAN", "DRUM"},
}

const maxContainerLen = 10

func randContainer(rng *rand.Rand, a *bufalloc.ByteAllocator) []byte {
	__antithesis_instrumentation__.Notify(698815)
	return randSyllables(rng, a, maxContainerLen, containerSyllables)
}

var segments = []string{
	"AUTOMOBILE", "BUILDING", "FURNITURE", "MACHINERY", "HOUSEHOLD",
}

func randSegment(rng *rand.Rand) []byte {
	__antithesis_instrumentation__.Notify(698816)
	return encoding.UnsafeConvertStringToBytes(segments[rng.Intn(len(segments))])
}

var priorities = []string{
	"1-URGENT", "2-HIGH", "3-MEDIUM", "4-NOT SPECIFIED",
}

func randPriority(rng *rand.Rand) []byte {
	__antithesis_instrumentation__.Notify(698817)
	return encoding.UnsafeConvertStringToBytes(priorities[rng.Intn(len(priorities))])
}

var instructions = []string{
	"DELIVER IN PERSON",
	"COLLECT COD", "NONE",
	"TAKE BACK RETURN",
}

func randInstruction(rng *rand.Rand) []byte {
	__antithesis_instrumentation__.Notify(698818)
	return encoding.UnsafeConvertStringToBytes(instructions[rng.Intn(len(instructions))])
}

var modes = []string{
	"REG AIR", "AIR", "RAIL", "SHIP", "TRUCK", "MAIL", "FOB",
}

func randMode(rng *rand.Rand) []byte {
	__antithesis_instrumentation__.Notify(698819)
	return []byte(modes[rng.Intn(len(modes))])
}
