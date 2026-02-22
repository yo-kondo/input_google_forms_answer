// Googleフォーム自動入力URL生成
//
// ■ 概要
// TOML設定ファイルを読み込み、GoogleフォームのURLに
// entry.<質問番号>=<回答内容> を付加した事前入力URLを生成する。
//
// ■ 仕様
// - usp=sf_link があれば削除
// - usp=pp_url を付与
// - answer が "{today}" の場合はシステム日付 (yyyy-mm-dd) に置換
// - 最後に「エンコードURL」と「デコードURL」を出力する
//
// ■ エラー方針
// - エラー発生時は原因が特定できるメッセージを標準エラー出力へ表示する

package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Config は TOML 全体を表す構造体
type Config struct {
	FormURL string  `toml:"form_url"` // GoogleフォームのURL
	Entries []Entry `toml:"entries"`  // 入力項目の配列
}

// Entry は 1つの質問入力情報を表す
type Entry struct {
	QuestionID string `toml:"question_id"` // 質問番号
	Answer     string `toml:"answer"`      // 入力値
	Comment    string `toml:"comment"`     // メモ用（処理では使用しない）
}

// メイン処理
func main() {

	// デフォルトの設定ファイル名
	cfgPath := "config.toml"

	// コマンドライン引数があればそれを優先
	if len(os.Args) >= 2 && os.Args[1] != "" {
		cfgPath = os.Args[1]
	}

	// 設定ファイル読み込み
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "設定ファイルの読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	// URL生成
	encodedURL, decodedURL, err := buildPrefillURL(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "URL生成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 出力（要件通り）
	fmt.Println("=== Encoded URL ===")
	fmt.Println(encodedURL)
	fmt.Println()
	fmt.Println("=== Decoded URL ===")
	fmt.Println(decodedURL)
}

// 設定ファイル読み込み
func loadConfig(path string) (Config, error) {
	var cfg Config

	// ファイルの存在確認
	if _, err := os.Stat(path); err != nil {
		return cfg, fmt.Errorf("設定ファイルが見つかりません: %s (%w)", path, err)
	}

	// TOML を構造体へデコード
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("TOMLの解析に失敗しました: %w", err)
	}

	// 必須項目チェック
	if cfg.FormURL == "" {
		return cfg, errors.New("form_url が空です")
	}

	return cfg, nil
}

// Googleフォーム事前入力URL生成
func buildPrefillURL(cfg Config) (encoded string, decoded string, err error) {

	// URL文字列を構造体に変換
	u, err := url.Parse(cfg.FormURL)
	if err != nil {
		return "", "", fmt.Errorf("form_url のURL解析に失敗しました: %w", err)
	}

	// 既存クエリパラメータを取得
	q := u.Query()

	// usp=sf_link がある場合は削除
	if q.Get("usp") == "sf_link" {
		q.Del("usp")
	}

	// usp=pp_url を必ず設定（上書き）
	q.Set("usp", "pp_url")

	// entry.<質問番号>=<回答> を追加

	// 今日の日付（yyyy-mm-dd）
	today := time.Now().Format("2006-01-02")

	for i, e := range cfg.Entries {

		// 質問番号チェック
		if e.QuestionID == "" {
			return "", "", fmt.Errorf("entries[%d].question_id が空です", i)
		}

		ans := e.Answer

		// {today} の特殊置換処理
		if ans == "{today}" {
			ans = today
		}

		// Googleフォーム仕様：
		// entry.質問番号=回答内容
		questionID := "entry." + e.QuestionID

		// Set は既存値を上書きする
		q.Set(questionID, ans)
	}

	// URLへクエリを再設定
	u.RawQuery = q.Encode()

	// 通常のURL（エンコード済み）
	encoded = u.String()

	// デコードURL生成（表示用）

	// RawQuery はURLエンコード済みなので、
	// 表示用にQueryUnescapeでデコードする
	decodedQuery, dqErr := url.QueryUnescape(u.RawQuery)
	if dqErr != nil {
		return "", "", fmt.Errorf("デコードURL生成に失敗しました: %w", dqErr)
	}

	// 手動で組み立て（String()は再エンコードしてしまうため）
	base := ""
	if u.Scheme != "" {
		base = u.Scheme + "://"
	}
	base += u.Host + u.Path

	if decodedQuery != "" {
		base += "?" + decodedQuery
	}

	if u.Fragment != "" {
		base += "#" + u.Fragment
	}

	decoded = base

	return encoded, decoded, nil
}
