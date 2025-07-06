# VSQL - PostgreSQL互換スキーマレスデータベース

VSQLは、PostgreSQL互換のプロトコルを持つスキーマレスのインメモリデータベースです。PostgreSQLの公式パーサー（pg_query_go）を使用しているため、完全なPostgreSQL構文をサポートします。

## 特徴

- PostgreSQLワイヤープロトコル互換（psqlから接続可能）
- PostgreSQL公式パーサー使用による完全な構文サポート
- スキーマレス（カラムの自由な追加、存在しないカラムはNULL扱い）
- インメモリストレージ
- 複雑なWHERE句のサポート（AND、OR、NOT）

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

-- テーブル作成（PostgreSQL構文でカラム定義が必要）
CREATE TABLE users (id int, name text, email text);

-- データ挿入（スキーマレスなので新しいカラムも追加可能）
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- 検索
SELECT * FROM users;
SELECT name, email FROM users WHERE id = 1;
SELECT * FROM users WHERE age > 25;

-- 更新
UPDATE users SET email = 'bob@example.com' WHERE name = 'Bob';

-- 削除
DELETE FROM users WHERE id = 1;

-- 複雑な条件
SELECT * FROM users WHERE age >= 30 AND name = 'Bob';
```

## 実装済み機能

- CREATE TABLE（PostgreSQL構文）
- INSERT INTO
- SELECT（*、特定カラム）
- UPDATE
- DELETE
- DROP TABLE
- WHERE句（=, !=, <>, >, <, >=, <=、AND、OR、NOT）

## 技術的特徴

- `github.com/pganalyze/pg_query_go/v5`を使用した本格的なSQL解析
- PostgreSQLの実際のパーサーを使用しているため、将来的な拡張が容易
- スキーマレス設計により、NoSQLのような柔軟性とSQLの表現力を両立

## 今後の拡張可能性

- ORDER BY、LIMIT句
- JOINのサポート
- サブクエリ
- 集約関数（COUNT、SUM、AVG等）
- 永続化機能
- インデックス
- トランザクション