# Knowledge Base

## Context

- task_id: `2026-08-29-flac-analyzer-task-record-contract`
- timestamp: `2026-08-29` (Asia/Tokyo)
- target environment: `A:\Users\letwir\repo\flac_analyzer` の添付Markdown資料、および `A:\Users\letwir\repo\llm-memory` のSKILL/CLI実装。
- source scope: `changeLOG_Implementation Plan.md`, `decisions.md`, `issues.md`, `memo.md`, `method.md`, `history.md`。
- status: 添付資料からの設計知識の正規化。flac_analyzer本体の現行コードをこのタスクでは変更していない。

## Findings

- 5950X/64GB環境での並列Producer-ConsumerとSharedMemory蓄積によるOOM・リソース枯渇を受け、PowerShellがFLACを列挙し、Pythonを1ファイルずつ同期起動する構成が採用された（添付資料記載。ライブ環境では未再検証）。
- `run_batch.ps1` は成功した絶対パスを `flac.done` に記録し、起動時にHashSet化してスキップする計画・実装記録がある（添付資料記載）。
- 解析経路は、デコード → 波形分離 → 特徴量抽出 → PostgreSQL UPSERT → FLACタグ更新の単一ファイル直列処理へ移行した（`history.md`記載）。
- 特徴量はFLACタグ用の丸め値とPostgreSQL JSONB用の生floatを分離し、Chroma、HPSS、Flux、Onset、Tempogram、Dynamic Range、MFCC等を拡張する方針が記録されている。
- CUEの複数格納場所、`audio_hash`、ステレオ入力のdownmix、Demucsのdrums/bass向けtempobeatなど、後続設計の論点が `memo.md` / `method.md` に記録されている。
- `issues.md` は現在進行中のイシューなしと記載している。ただし、これは添付文書の状態であり、現在のflac_analyzer実装を再テストした結果ではない。

## Morphism

- 添付された人間向け状態資料を、`llm-memory` の再利用可能な `knowledge` レコードへ変換した。
- 未確認のライブ状態は「添付資料記載」または「未再検証」と明示し、検証済み事実へ昇格していない。
- 再利用時は、実コード・実DB・実FLAC出力を別途確認してから現在の事実として扱う。
