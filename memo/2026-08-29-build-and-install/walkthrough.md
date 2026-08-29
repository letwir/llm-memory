# Walkthrough: build and install batch

## 実行計画

1. Existing build and skill destinationsを確認する。
2. Build成功後に全skill pathへ安全に配布するbatchを追加する。
3. Missing configurationの停止、unit test、構文経路を確認する。
4. 結果をllm-memへ登録する。

## 実施内容

- `BUILD_AND_INSTALL.bat`を追加。
- `INSTALL.bat`には依存せず、6配置先へSKILLとexeを配布。
- `pwsh.exe`とGoの存在、DB URL設定、build出力を検証。
- `USAGE.md`に利用方法を追記。

## 検証結果

- `go test ./...`: PASS
- DB URL未設定時のbatch guard: PASS
- `pwsh build.ps1`による一時exe build: PASS
- 実配布: 未実行
