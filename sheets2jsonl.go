package sheets2jsonl

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"os"
)

type VerboseRow struct {
	Book  string   `json:"book"`
	Sheet string   `json:"sheet"`
	Row   uint32   `json:"row"`
	Cols  []string `json:"cols"`
}

type Rows iter.Seq2[VerboseRow, error]

type BookName = string

type BookToRows func(context.Context, BookName) Rows

type ReaderToRows func(context.Context, io.Reader) Rows

type Writer func(context.Context, VerboseRow) error

type WriterJSON struct{ *json.Encoder }

func (j WriterJSON) ToWriter() Writer {
	return func(_ context.Context, row VerboseRow) error {
		return j.Encoder.Encode(row)
	}
}

func (w Writer) WriteAll(ctx context.Context, rows Rows) error {
	for row := range rows {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := w(ctx, row)
		if nil != err {
			return err
		}
	}

	return nil
}

type IoWriter struct{ io.Writer }

func (i IoWriter) ToWriterJSON() WriterJSON {
	return WriterJSON{Encoder: json.NewEncoder(i.Writer)}
}

type Converter struct {
	ReaderToRows
	Writer
}

func (c Converter) Convert(ctx context.Context, rdr io.Reader) error {
	var rows Rows = c.ReaderToRows(ctx, rdr)
	return c.Writer.WriteAll(ctx, rows)
}

func (r ReaderToRows) StdinToStdout(ctx context.Context) error {
	var rdr io.Reader = os.Stdin
	var wtr *bufio.Writer = bufio.NewWriter(os.Stdout)
	conv := Converter{
		ReaderToRows: r,
		Writer: IoWriter{Writer: wtr}.
			ToWriterJSON().
			ToWriter(),
	}
	return errors.Join(conv.Convert(ctx, rdr), wtr.Flush())
}
