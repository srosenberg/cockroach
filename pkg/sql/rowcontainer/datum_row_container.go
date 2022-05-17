package rowcontainer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/bits"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type RowContainer struct {
	_ util.NoCopy

	numCols int

	rowsPerChunk      int
	rowsPerChunkShift uint
	chunks            [][]tree.Datum
	firstChunk        [1][]tree.Datum
	numRows           int

	chunkMemSize int64

	fixedColsSize int64

	varSizedColumns []int

	deletedRows int

	memAcc mon.BoundAccount
}

func NewRowContainer(acc mon.BoundAccount, ti colinfo.ColTypeInfo) *RowContainer {
	__antithesis_instrumentation__.Notify(568937)
	return NewRowContainerWithCapacity(acc, ti, 0)
}

func NewRowContainerWithCapacity(
	acc mon.BoundAccount, ti colinfo.ColTypeInfo, rowCapacity int,
) *RowContainer {
	__antithesis_instrumentation__.Notify(568938)
	c := &RowContainer{}
	c.Init(acc, ti, rowCapacity)
	return c
}

var rowsPerChunkShift = uint(util.ConstantWithMetamorphicTestValue(
	"row-container-rows-per-chunk-shift",
	6,
	1,
))

func (c *RowContainer) Init(acc mon.BoundAccount, ti colinfo.ColTypeInfo, rowCapacity int) {
	__antithesis_instrumentation__.Notify(568939)
	nCols := ti.NumColumns()

	c.numCols = nCols
	c.memAcc = acc

	if rowCapacity != 0 {
		__antithesis_instrumentation__.Notify(568942)

		c.rowsPerChunkShift = 64 - uint(bits.LeadingZeros64(uint64(rowCapacity-1)))
	} else {
		__antithesis_instrumentation__.Notify(568943)
		if nCols != 0 {
			__antithesis_instrumentation__.Notify(568944)

			c.rowsPerChunkShift = rowsPerChunkShift
		} else {
			__antithesis_instrumentation__.Notify(568945)

			c.rowsPerChunkShift = 32
		}
	}
	__antithesis_instrumentation__.Notify(568940)
	c.rowsPerChunk = 1 << c.rowsPerChunkShift

	for i := 0; i < nCols; i++ {
		__antithesis_instrumentation__.Notify(568946)
		sz, variable := tree.DatumTypeSize(ti.Type(i))
		if variable {
			__antithesis_instrumentation__.Notify(568947)
			if c.varSizedColumns == nil {
				__antithesis_instrumentation__.Notify(568949)

				c.varSizedColumns = make([]int, 0, nCols)
			} else {
				__antithesis_instrumentation__.Notify(568950)
			}
			__antithesis_instrumentation__.Notify(568948)
			c.varSizedColumns = append(c.varSizedColumns, i)
		} else {
			__antithesis_instrumentation__.Notify(568951)
			c.fixedColsSize += int64(sz)
		}
	}
	__antithesis_instrumentation__.Notify(568941)

	if nCols > 0 {
		__antithesis_instrumentation__.Notify(568952)

		c.chunkMemSize = memsize.DatumOverhead * int64(c.rowsPerChunk*c.numCols)
		c.chunkMemSize += memsize.DatumsOverhead
	} else {
		__antithesis_instrumentation__.Notify(568953)
	}
}

func (c *RowContainer) Clear(ctx context.Context) {
	__antithesis_instrumentation__.Notify(568954)
	c.chunks = nil
	c.numRows = 0
	c.deletedRows = 0
	c.memAcc.Clear(ctx)
}

func (c *RowContainer) UnsafeReset(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(568955)
	c.numRows = 0
	c.deletedRows = 0
	return c.memAcc.ResizeTo(ctx, int64(len(c.chunks))*c.chunkMemSize)
}

func (c *RowContainer) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(568956)
	if c == nil {
		__antithesis_instrumentation__.Notify(568958)

		return
	} else {
		__antithesis_instrumentation__.Notify(568959)
	}
	__antithesis_instrumentation__.Notify(568957)
	c.chunks = nil
	c.varSizedColumns = nil
	c.memAcc.Close(ctx)
}

func (c *RowContainer) allocChunks(ctx context.Context, numChunks int) error {
	__antithesis_instrumentation__.Notify(568960)
	datumsPerChunk := c.rowsPerChunk * c.numCols

	if err := c.memAcc.Grow(ctx, c.chunkMemSize*int64(numChunks)); err != nil {
		__antithesis_instrumentation__.Notify(568964)
		return err
	} else {
		__antithesis_instrumentation__.Notify(568965)
	}
	__antithesis_instrumentation__.Notify(568961)

	if c.chunks == nil {
		__antithesis_instrumentation__.Notify(568966)
		if numChunks == 1 {
			__antithesis_instrumentation__.Notify(568967)
			c.chunks = c.firstChunk[:0:1]
		} else {
			__antithesis_instrumentation__.Notify(568968)
			c.chunks = make([][]tree.Datum, 0, numChunks)
		}
	} else {
		__antithesis_instrumentation__.Notify(568969)
	}
	__antithesis_instrumentation__.Notify(568962)

	datums := make([]tree.Datum, numChunks*datumsPerChunk)
	for i, pos := 0, 0; i < numChunks; i++ {
		__antithesis_instrumentation__.Notify(568970)
		c.chunks = append(c.chunks, datums[pos:pos+datumsPerChunk])
		pos += datumsPerChunk
	}
	__antithesis_instrumentation__.Notify(568963)
	return nil
}

func (c *RowContainer) rowSize(row tree.Datums) int64 {
	__antithesis_instrumentation__.Notify(568971)
	rsz := c.fixedColsSize
	for _, i := range c.varSizedColumns {
		__antithesis_instrumentation__.Notify(568973)
		rsz += int64(row[i].Size())
	}
	__antithesis_instrumentation__.Notify(568972)
	return rsz
}

func (c *RowContainer) getChunkAndPos(rowIdx int) (chunk int, pos int) {
	__antithesis_instrumentation__.Notify(568974)

	row := rowIdx + c.deletedRows
	chunk = row >> c.rowsPerChunkShift
	return chunk, (row - (chunk << c.rowsPerChunkShift)) * (c.numCols)
}

func (c *RowContainer) AddRow(ctx context.Context, row tree.Datums) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(568975)
	if len(row) != c.numCols {
		__antithesis_instrumentation__.Notify(568980)
		panic(errors.AssertionFailedf("invalid row length %d, expected %d", len(row), c.numCols))
	} else {
		__antithesis_instrumentation__.Notify(568981)
	}
	__antithesis_instrumentation__.Notify(568976)
	if c.numCols == 0 {
		__antithesis_instrumentation__.Notify(568982)
		if c.chunks == nil {
			__antithesis_instrumentation__.Notify(568984)
			c.chunks = [][]tree.Datum{{}}
		} else {
			__antithesis_instrumentation__.Notify(568985)
		}
		__antithesis_instrumentation__.Notify(568983)
		c.numRows++
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(568986)
	}
	__antithesis_instrumentation__.Notify(568977)

	if err := c.memAcc.Grow(ctx, c.rowSize(row)); err != nil {
		__antithesis_instrumentation__.Notify(568987)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(568988)
	}
	__antithesis_instrumentation__.Notify(568978)
	chunk, pos := c.getChunkAndPos(c.numRows)
	if chunk == len(c.chunks) {
		__antithesis_instrumentation__.Notify(568989)

		numChunks := 1 + len(c.chunks)/8
		if err := c.allocChunks(ctx, numChunks); err != nil {
			__antithesis_instrumentation__.Notify(568990)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568991)
		}
	} else {
		__antithesis_instrumentation__.Notify(568992)
	}
	__antithesis_instrumentation__.Notify(568979)
	copy(c.chunks[chunk][pos:pos+c.numCols], row)
	c.numRows++
	return c.chunks[chunk][pos : pos+c.numCols : pos+c.numCols], nil
}

func (c *RowContainer) Len() int {
	__antithesis_instrumentation__.Notify(568993)
	return c.numRows
}

func (c *RowContainer) NumCols() int {
	__antithesis_instrumentation__.Notify(568994)
	return c.numCols
}

func (c *RowContainer) At(i int) tree.Datums {
	__antithesis_instrumentation__.Notify(568995)

	chunk, pos := c.getChunkAndPos(i)
	return c.chunks[chunk][pos : pos+c.numCols : pos+c.numCols]
}

func (c *RowContainer) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(568996)
	r1 := c.At(i)
	r2 := c.At(j)
	for idx := 0; idx < c.numCols; idx++ {
		__antithesis_instrumentation__.Notify(568997)
		r1[idx], r2[idx] = r2[idx], r1[idx]
	}
}

func (c *RowContainer) PopFirst(ctx context.Context) {
	__antithesis_instrumentation__.Notify(568998)
	if c.numRows == 0 {
		__antithesis_instrumentation__.Notify(569000)
		panic("no rows added to container, nothing to pop")
	} else {
		__antithesis_instrumentation__.Notify(569001)
	}
	__antithesis_instrumentation__.Notify(568999)
	c.numRows--
	if c.numCols != 0 {
		__antithesis_instrumentation__.Notify(569002)
		c.deletedRows++
		if c.deletedRows == c.rowsPerChunk {
			__antithesis_instrumentation__.Notify(569003)

			size := c.chunkMemSize
			for i, pos := 0, 0; i < c.rowsPerChunk; i, pos = i+1, pos+c.numCols {
				__antithesis_instrumentation__.Notify(569005)
				size += c.rowSize(c.chunks[0][pos : pos+c.numCols])
			}
			__antithesis_instrumentation__.Notify(569004)

			c.chunks[0] = nil
			c.deletedRows = 0
			c.chunks = c.chunks[1:]
			c.memAcc.Shrink(ctx, size)
		} else {
			__antithesis_instrumentation__.Notify(569006)
		}
	} else {
		__antithesis_instrumentation__.Notify(569007)
	}
}

func (c *RowContainer) Replace(ctx context.Context, i int, newRow tree.Datums) error {
	__antithesis_instrumentation__.Notify(569008)
	newSz := c.rowSize(newRow)
	row := c.At(i)
	oldSz := c.rowSize(row)
	if newSz != oldSz {
		__antithesis_instrumentation__.Notify(569010)
		if err := c.memAcc.Resize(ctx, oldSz, newSz); err != nil {
			__antithesis_instrumentation__.Notify(569011)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569012)
		}
	} else {
		__antithesis_instrumentation__.Notify(569013)
	}
	__antithesis_instrumentation__.Notify(569009)
	copy(row, newRow)
	return nil
}
