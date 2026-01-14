package main

/*
line2md will query profile with certain lines.
The cum and flat will be returned and printed in markdown table and csv files.
The output files are in directory specified by user through -dir option.

The input is lines to get cum/flat data. For example:

runtime.mallocgcSmallScanNoHeader
src/runtime/malloc.go:1389,publicationBarrier
src/runtime/malloc.go:1399,span.freeIndexForScan = span.freeindex

runtime.mallocgcSmallNoscan
src/runtime/malloc.go:1298,publicationBarrier
src/runtime/malloc.go:1308,span.freeIndexForScan = span.freeindex
src/runtime/malloc.go:1327,c.nextSample -= int64(size)
src/runtime/malloc.go:1328,c.nextSample cond

There are many sections in the input separated by empty lines. For each section, the first line is the function name.
The following lines are formated in csv style. The first column is the query file:line. Notice that "file" can be suffix.
The query engine should handle this. The 2nd column is the code, as a comment of the line for faster lookup.

The output is markdown tables in collect.md and many csv files named as <function name>.csv.

In collect.md, for each section, the function name is printed. Then is the markdown table with the following format:

| line | code | flat | cum |
...

<function name>.csv has similar orginization of data.
*/
