protoc-gen-json2bin
================
[![license](https://img.shields.io/badge/license-MIT-4183c4.svg)](https://github.com/peto-tn/protoc-gen-json2bin/blob/master/LICENSE)

`protoc-gen-json2bin` は `json` から `protobufバイナリ` を生成する `protocプラグイン` です。  
`protocプラグイン` は本来、任意の言語に向けてのクラス等を生成するのに用いられますが、 `protoc-gen-json2bin` では定義ファイルである `.proto` の他に `json` も入力として与え、ソースコードではなく `protobuf` のバイナリデータファイルを出力します。

----
## インストール
```
go get github.com/peto-tn/protoc-gen-json2bin
```

## 使い方
```
protoc --plugin=$GOPATH/bin/protoc-gen-json2bin --json2bin_out=json=<JSON_FILE_PATH>,message_type=<MESSAGE_TYPE>,output=<OUTPUT_FILE_NAME>:<OUTPUT_DIR> <PROTO_FILE>
```
