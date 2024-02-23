verifyarc
=========

[English](/README.md)/Japanese

zip/tar ファイルと、それを展開したと思われるディレクトリがあった時、両者が一致していて一方を削除してしまっても大丈夫かを検証します

```
$ verifyarc {-C (DIR)} foo.zip
```

```
$ verifyarc {-C (DIR)} foo.tar
```

```
$ gzip -dc FOO.tar.gz | verifyarc {-C (DIR)} -
```


- 拡張子が .zip でない場合は、非圧縮の tar アーカイブとみなされます。
    - 標準入力は常に tar アーカイブ扱いとなります。
- `foo.zip` が `A.txt`, `B.bin` と `C.exe` を保持している時, `verifyarc` はそれらと `(DIR)/A.txt`, `(DIR)/B.bin` と `(DIR)/C.exe` を比較します
    - 一つでも異なるファイルがあれば、ただちにエラー終了します
- `(DIR)/D.obj` が存在するが、`D.obj` が `FOO.ZIP` にない時はそれも報告します
    - 存在しないファイルを全部列挙するまで処理を継続します。
- `-C (DIR)` が省略された時、`(DIR)` はカレントディレクトリになります

```
$ verifyarc {-C (DIR)} (SUBDIR)
```

- アーカイブファイルのかわりに、展開済みのファイルシステムを検証します
    - (SUBDIR) と (DIR)/(SUBDIR) が同じかをテストします
    - `tar cf - (SUBDIR) | verifyarc -C (DIR) -` と等価です。
