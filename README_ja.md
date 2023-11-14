verifyarc
=========

[English](/README.md)/Japanese

ZIPファイルと、それを展開したと思われるディレクトリがあった時、両者が一致していて一方を削除してしまっても大丈夫かを検証するツール

```
$ verifyarc {-C (DIR)} FOO.ZIP
```

- `FOO.ZIP` が `A.txt`, `B.bin` と `C.exe` を保持している時, `verifyarc` はそれらと `(DIR)/A.txt`, `(DIR)/B.bin` と `(DIR)/C.exe` を比較します
    - 一つでも異なるファイルがあれば、ただちにエラー終了します
- `(DIR)/D.obj` が存在するが、`D.obj` が `FOO.ZIP` にない時はそれも報告します
    - 存在しないファイルを全部列挙するまで処理を継続します。
- `-C (DIR)` が省略された時、`(DIR)` はカレントディレクトリになります
