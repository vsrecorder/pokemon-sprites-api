# pokemon-sprites-api

ポケモンのスプライト（画像）情報を提供する REST API サーバです。

対戦記録サービス [vsrecorder](https://vsrecorder.mobi) のフロントエンドから、
デッキ登録時などにポケモンのアイコン画像を表示するために利用されます。

## 概要

PostgreSQL に格納されたポケモンの一覧（図鑑番号・日本語名）を返す API を提供します。

レスポンスには、各ポケモンのスプライト画像を配信している CDN の画像 URL が含まれます。

- メガシンカ（`_mega`）やキョダイマックス（`_gmax`）などのフォルム違いにも対応
- 収録件数は約 1,380 件（`db/pokemon-sprites.sql` 参照）

## API

ベースパス: `/api/v1beta`

### `GET /api/v1beta/pokemon-sprites`

登録されているポケモンのスプライト情報を、図鑑 ID の昇順ですべて返します。

#### レスポンス例

```json
[
  {
    "id": "0001",
    "name": "フシギダネ",
    "image_url": "https://xx8nnpgt.user.webaccel.jp/images/pokemon-sprites/1.png"
  },
  {
    "id": "0003_mega",
    "name": "メガフシギバナ",
    "image_url": "https://xx8nnpgt.user.webaccel.jp/images/pokemon-sprites/3_mega.png"
  }
]
```

| フィールド  | 説明                                                        |
| ----------- | ----------------------------------------------------------- |
| `id`        | 図鑑 ID（4 桁ゼロ埋め、フォルム違いはサフィックス付き）     |
| `name`      | ポケモンの日本語名                                          |
| `image_url` | スプライト画像の URL（`id` の先頭のゼロを除いたものを使用） |

## アーキテクチャ

クリーンアーキテクチャ（レイヤードアーキテクチャ）で構成しています。

```
cmd/main.go                              エントリポイント / サーバ起動・DI・CORS設定
internal/
├── controller/                          HTTPハンドラ（Gin）
│   ├── pokemon_sprite.go                 ルーティング・リクエスト処理
│   ├── dto/                              レスポンス構造体
│   └── presenter/                        エンティティ → レスポンス変換
├── usecase/                             ユースケース（アプリケーションロジック）
├── domain/
│   ├── entity/                           ドメインエンティティ
│   └── repository/                       リポジトリのインターフェース定義
└── infrastructure/                      永続化の実装
    ├── pokemon_sprite.go                 リポジトリ実装（GORM）
    ├── model/                            DBモデル
    └── postgres/                         PostgreSQL接続
```

依存の方向は `controller → usecase → repository(interface) ← infrastructure` となっており、
ドメイン層は外側の実装に依存しません。

## 技術スタック

- 言語: Go 1.27
- Web フレームワーク: [Gin](https://github.com/gin-gonic/gin)
- ORM: [GORM](https://gorm.io/)（PostgreSQL ドライバ）
- データベース: PostgreSQL
- 実行環境: Docker / Docker Compose（distroless イメージ）

## セットアップ

### 1. 環境変数の設定

`.env.sample` をもとに `.env` を作成し、DB 接続情報を設定します。

```
DB_HOSTNAME=
DB_PORT=
DB_USER_NAME=
DB_USER_PASSWORD=
DB_NAME=
```

### 2. データベースの準備

`db/pokemon-sprites.sql` を PostgreSQL に流し込み、`pokemon_sprites` テーブルと初期データを作成します。

```sql
CREATE TABLE pokemon_sprites (
    id      VARCHAR(128) PRIMARY KEY,
    name    VARCHAR(256) NOT NULL
);
```

### 3. 起動

ローカルで直接起動する場合:

```bash
make run
```

サーバは `127.0.0.1:8998`(ループバックのみ)で起動します。本番も同じで、外からは nginx 経由でしか届きません。

## Makefile タスク

| コマンド       | 説明                               |
| -------------- | ---------------------------------- |
| `make run`     | ローカルで直接起動（`go run`）     |
| `make build`   | ビルドの確認                       |
| `make image`   | Docker イメージのビルド & プッシュ |
| `make up`      | Docker Compose で起動              |
| `make down`    | Docker Compose を停止              |
| `make restart` | 停止して再起動                     |
| `make deploy`  | 最新イメージを pull してデプロイ   |
| `make log`     | コンテナのログを追跡表示           |

## CORS

以下のオリジンからのアクセスを許可しています。

- `http://localhost:3000`
- `https://vsrecorder.mobi`
- `https://local.vsrecorder.mobi`
  </content>
  </invoke>
