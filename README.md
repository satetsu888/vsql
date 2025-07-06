# VSQL - PostgreSQL互換スキーマレスデータベース

VSQLは、PostgreSQL互換のプロトコルを持つスキーマレスのインメモリデータベースです。

## 特徴

- PostgreSQLワイヤープロトコル互換（psqlから接続可能）
- スキーマレス（テーブル定義不要、カラムの自由な追加）
- インメモリストレージ
- 基本的なSQL操作（SELECT、INSERT、CREATE TABLE）
- WHERE句による条件フィルタリング

## ビルドと実行

```bash
# ビルド
go build -o vsql

# 実行（デフォルトはポート5432）
./vsql

# ポート指定
./vsql -port 5433
```

## 使用例

```sql
-- psqlから接続
psql -h localhost -p 5432 -U any_user -d any_database

-- テーブル作成
CREATE TABLE users;

-- データ挿入（カラムは自由）
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- 検索
SELECT * FROM users;
SELECT name, email FROM users WHERE id = 1;
SELECT * FROM users WHERE age > 25;
```

## 実装済み機能

- CREATE TABLE
- INSERT INTO
- SELECT（*, 特定カラム）
- WHERE句（=, !=, <>, >, <, >=, <=）

## 今後の拡張可能性

- UPDATE、DELETE文のサポート
- ORDER BY、LIMIT句
- JOINのサポート
- 永続化機能
- インデックス