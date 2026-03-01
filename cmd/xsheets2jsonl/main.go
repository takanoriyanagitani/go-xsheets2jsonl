package main

import (
	"context"
	"flag"
	"log"

	xj "github.com/takanoriyanagitani/go-xsheets2jsonl"
	xrdr "github.com/takanoriyanagitani/go-xsheets2jsonl/xfile/reader"
)

func sub(ctx context.Context, bookName string) error {
	var bknm xrdr.BookName = xrdr.BookName(bookName)
	var rdr2rows xj.ReaderToRows = bknm.ToReaderToRows()

	return rdr2rows.StdinToStdout(ctx)
}

func main() {
	var bookName string
	flag.StringVar(&bookName, "name-of-the-book", "/path/to/dummy-book-name.xlsx", "book name")
	flag.Parse()
	err := sub(context.Background(), bookName)
	if nil != err {
		log.Printf("%v\n", err)
	}
}
