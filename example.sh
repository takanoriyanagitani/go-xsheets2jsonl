#!/bin/sh

ixlsx="./input.xlsx"

geninput() {
	echo generating the input xls file...

	jq -c -n '[
		{time:"2026-02-25T03:22:34.012Z", status:200, body:"apt update done"},
		{},
		{time:"2026-02-25T03:22:34.012Z", status:200, body:""}
	]' |
		python3 \
			-c 'import pandas; import functools; import json; import sys; import operator; functools.reduce(
				lambda state, f: f(state),
				[
					json.load,
					pandas.DataFrame,
					operator.methodcaller(
						"to_excel",
						"input.xlsx",
						index=False,
					),
				],
				sys.stdin.buffer,
			)'

}

test -f "${ixlsx}" || geninput

test -f "${ixlsx}" || exec sh -c '
	echo unable to create the input xlsx file.
	exit 1
'

cat "${ixlsx}" |
	wasmtime \
		run \
		./xsheets2jsonl.wasm \
		--name-of-the-book "my-test-book.xlsx" |
	jq -c
