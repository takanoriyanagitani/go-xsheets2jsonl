package xrdr

import (
	"context"
	"io"
	"iter"
	"log"

	xj "github.com/takanoriyanagitani/go-xsheets2jsonl"
	xr "github.com/xuri/excelize/v2"
)

type Xfile struct{ *xr.File }

func (x Xfile) Close() error { return x.File.Close() }

func (x Xfile) Sheets() []string { return x.File.GetSheetList() }

func (x Xfile) RowsBySheetName(sheet string) (*xr.Rows, error) {
	return x.File.Rows(sheet)
}

type Xrows struct{ *xr.Rows }

func (r Xrows) Close() error { return r.Rows.Close() }

func (r Xrows) ToVerboseRows(
	bookName string,
	sheetName string,
) iter.Seq2[xj.VerboseRow, error] {
	return func(yield func(xj.VerboseRow, error) bool) {
		defer func() {
			err := r.Close()
			if nil != err {
				log.Printf("%v\n", err)
			}
		}()

		var rowNo uint32 = 0

		for r.Rows.Next() {
			rowNo += 1
			row, err := r.Rows.Columns()
			vrow := xj.VerboseRow{
				Book:  bookName,
				Sheet: sheetName,
				Row:   rowNo,
				Cols:  row,
			}

			if !yield(vrow, err) {
				return
			}
		}
	}
}

func (x Xfile) ToVerboseRows(
	bookName string,
) iter.Seq2[xj.VerboseRow, error] {
	return func(yield func(xj.VerboseRow, error) bool) {
		defer func() {
			err := x.Close()
			if nil != err {
				log.Printf("%v\n", err)
			}
		}()

		var sheets []string = x.Sheets()

		for _, sheet := range sheets {
			rows, err := x.RowsBySheetName(sheet)
			if nil != err {
				yield(xj.VerboseRow{}, err)
				return
			}

			xrows := Xrows{Rows: rows}
			var vrows iter.Seq2[xj.VerboseRow, error] = xrows.ToVerboseRows(
				bookName,
				sheet,
			)

			for vrow, err := range vrows {
				if !yield(vrow, err) {
					return
				}
			}
		}
	}
}

type BookName string

func (n BookName) ToReaderToRows() xj.ReaderToRows {
	return func(ctx context.Context, rdr io.Reader) xj.Rows {
		return func(yield func(xj.VerboseRow, error) bool) {
			file, err := xr.OpenReader(rdr)
			if nil != err {
				yield(xj.VerboseRow{}, err)
				return
			}

			xfile := Xfile{File: file}

			var vrows iter.Seq2[xj.VerboseRow, error] = xfile.ToVerboseRows(
				string(n),
			)

			for row, err := range vrows {
				select {
				case <-ctx.Done():
					yield(xj.VerboseRow{}, ctx.Err())
					return
				default:
				}

				if !yield(row, err) {
					return
				}
			}
		}
	}
}

//nolint:gochecknoglobals
var BookNameDefault BookName = "__UNNAMED_BOOK__"
