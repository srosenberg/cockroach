// cr2pg is a program that reads CockroachDB-formatted SQL files on stdin,
// modifies them to be Postgres compatible, and outputs them to stdout.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"

	"github.com/cockroachdb/cockroach/pkg/cmd/cr2pg/sqlstream"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

func main() {
	__antithesis_instrumentation__.Notify(38170)
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	done := ctx.Done()

	readStmts := make(chan tree.Statement, 100)
	writeStmts := make(chan tree.Statement, 100)

	g.Go(func() error {
		__antithesis_instrumentation__.Notify(38174)
		defer close(readStmts)
		stream := sqlstream.NewStream(os.Stdin)
		for {
			__antithesis_instrumentation__.Notify(38176)
			stmt, err := stream.Next()
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(38179)
				break
			} else {
				__antithesis_instrumentation__.Notify(38180)
			}
			__antithesis_instrumentation__.Notify(38177)
			if err != nil {
				__antithesis_instrumentation__.Notify(38181)
				return err
			} else {
				__antithesis_instrumentation__.Notify(38182)
			}
			__antithesis_instrumentation__.Notify(38178)
			select {
			case readStmts <- stmt:
				__antithesis_instrumentation__.Notify(38183)
			case <-done:
				__antithesis_instrumentation__.Notify(38184)
				return nil
			}
		}
		__antithesis_instrumentation__.Notify(38175)
		return nil
	})
	__antithesis_instrumentation__.Notify(38171)
	g.Go(func() error {
		__antithesis_instrumentation__.Notify(38185)
		defer close(writeStmts)
		newstmts := make([]tree.Statement, 8)
		for stmt := range readStmts {
			__antithesis_instrumentation__.Notify(38187)
			newstmts = newstmts[:1]
			newstmts[0] = stmt
			switch stmt := stmt.(type) {
			case *tree.CreateTable:
				__antithesis_instrumentation__.Notify(38189)
				stmt.PartitionByTable = nil
				var newdefs tree.TableDefs
				for _, def := range stmt.Defs {
					__antithesis_instrumentation__.Notify(38191)
					switch def := def.(type) {
					case *tree.FamilyTableDef:
						__antithesis_instrumentation__.Notify(38192)

					case *tree.IndexTableDef:
						__antithesis_instrumentation__.Notify(38193)

						newstmts = append(newstmts, &tree.CreateIndex{
							Name:     def.Name,
							Table:    stmt.Table,
							Inverted: def.Inverted,
							Columns:  def.Columns,
							Storing:  def.Storing,
						})
					case *tree.UniqueConstraintTableDef:
						__antithesis_instrumentation__.Notify(38194)
						if def.PrimaryKey {
							__antithesis_instrumentation__.Notify(38197)

							for i, col := range def.Columns {
								__antithesis_instrumentation__.Notify(38199)
								if col.Direction != tree.Ascending {
									__antithesis_instrumentation__.Notify(38201)
									return errors.New("PK directions not supported by postgres")
								} else {
									__antithesis_instrumentation__.Notify(38202)
								}
								__antithesis_instrumentation__.Notify(38200)
								def.Columns[i].Direction = tree.DefaultDirection
							}
							__antithesis_instrumentation__.Notify(38198)

							def.Name = ""
							newdefs = append(newdefs, def)
							break
						} else {
							__antithesis_instrumentation__.Notify(38203)
						}
						__antithesis_instrumentation__.Notify(38195)
						newstmts = append(newstmts, &tree.CreateIndex{
							Name:     def.Name,
							Table:    stmt.Table,
							Unique:   true,
							Inverted: def.Inverted,
							Columns:  def.Columns,
							Storing:  def.Storing,
						})
					default:
						__antithesis_instrumentation__.Notify(38196)
						newdefs = append(newdefs, def)
					}
				}
				__antithesis_instrumentation__.Notify(38190)
				stmt.Defs = newdefs
			}
			__antithesis_instrumentation__.Notify(38188)
			for _, stmt := range newstmts {
				__antithesis_instrumentation__.Notify(38204)
				select {
				case writeStmts <- stmt:
					__antithesis_instrumentation__.Notify(38205)
				case <-done:
					__antithesis_instrumentation__.Notify(38206)
					return nil
				}
			}
		}
		__antithesis_instrumentation__.Notify(38186)
		return nil
	})
	__antithesis_instrumentation__.Notify(38172)
	g.Go(func() error {
		__antithesis_instrumentation__.Notify(38207)
		w := bufio.NewWriterSize(os.Stdout, 1024*1024)
		fmtctx := tree.NewFmtCtx(tree.FmtSimple)
		for stmt := range writeStmts {
			__antithesis_instrumentation__.Notify(38209)
			stmt.Format(fmtctx)
			_, _ = w.WriteString(fmtctx.CloseAndGetString())
			_, _ = w.WriteString(";\n\n")
		}
		__antithesis_instrumentation__.Notify(38208)
		w.Flush()
		return nil
	})
	__antithesis_instrumentation__.Notify(38173)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(38210)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(38211)
	}
}
