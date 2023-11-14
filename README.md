verifyarc
=========

English/[Japanese](./README_ja.md)

A tool that verifies that when there is a ZIP file and a directory in which it seems to have been extracted, both match and it is safe to delete one.

```
$ verifyarc {-C (DIR)} FOO.ZIP
```

- When `FOO.ZIP` contains `A.txt`, `B.bin` and `C.exe`, `verifies` compares them and `(DIR)/A.txt`, `(DIR)/B.bin` and `(DIR)/C.exe`.
    - When it finds a different file, it stops immediately.
- When `(DIR)/D.obj` exists, but `FOO.ZIP` does not contain `D.obj`, `verifies` reports it.
    - It continues until displaying all files that are not found in the archive.
- When `-C (DIR)` is omitted, `(DIR)` is set the current working directory.
