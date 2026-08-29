# llm-memory 改善Issue

実装順に並べる。状態は `[ ]` TODO / `[-]` WIP / `[~]` implemented / `[*]` tested / `[x]` done。

- [~] ISSUE-001: UPDATE時のtransaction-time整合性を修正する
  - `SUPERSEDED`化した旧memoryへ `tx_invalidated_at` を設定する。
  - valid time と transaction time の両方を検証するテストを追加する。
- [~] ISSUE-002: グラフ・セクション取り込みの部分失敗を可視化する
  - ノード／エッジ失敗を握り潰さず、結果またはエラーとして呼び出し元へ返す。
  - セクション単位の失敗数と対象を失わない。
- [~] ISSUE-003: compactとWORM方針を整合させる
  - 履歴を壊す通常UPDATEを許すのか、追記型compactionにするのかを実装と文書で統一する。
  - DB権限またはtrigger等、選択した境界をテスト可能にする。
- [~] ISSUE-004: ADD/UPDATE/NOOPの競合窓を閉じる
  - 現運用ではカテゴリ＋正規化タイトルをACTIVE identityとして一意制約を適用済み。source/task-specific titleを要求する。
  - 同一内容の重複とversion飛びを再現テストする。
  - 完了済み: 既存の完全一致重複（rule / walkthrough）はDB側で本文マージ後、1件へ統合した。
- [~] ISSUE-005: タイトル完全一致依存を弱める
  - 正規化本文・タグ・Memory Object等を候補判定に加える。
  - 誤UPDATEと重複ADDの境界をテストする。
- [~] ISSUE-006: 検索を外部脳向けに拡張する
  - PostgreSQL内のtrigger管理tsvector全文検索＋現行substring検索を併用する。意味検索（embedding）は別Issueとして残す。
  - 導入しない環境でも従来CLIが動く。
- [~] ISSUE-007: 看板・ライセンス・設計文書を実装へ合わせる
  - RAG / JITMIND / Graphiti / WORM の表現を実体に合わせる。
  - MITとProprietaryの権利表記を一つの法的方針へ統一する（README/LICENSEをMITへ統一）。

## 残作業

- embeddingによる意味検索は未実装。pgvector導入は、モデル・次元数・再embedding方針を決めてから別Issue化する。
- PostgreSQL実機でのDDL適用、競合投入、WORM境界の検証は未実施。
