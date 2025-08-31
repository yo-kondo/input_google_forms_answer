// Googleフォーム自動入力URL作成

package main

import (
	"fmt"
	"net/url"
	"time"
)

// フォームの入力項目を定義する構造体
type FormEntry struct {
	QuestionID string
	Answer     string
	Comment    string
}

func main() {
	// プログラム内のグローバル変数として設定
	const baseURL = "https://docs.google.com/forms/d/e/TEST/viewform?usp=sf_link"

	// 質問番号、回答内容、コメントの配列
	formEntries := []FormEntry{
		{QuestionID: "917226918", Answer: "東京", Comment: "所属拠点"},
		{QuestionID: "59099188", Answer: "{today}", Comment: "日付（yyyy-mm-dd）"},
		{QuestionID: "646785265", Answer: "1234567890", Comment: "社員番号"},
		{QuestionID: "1446251705", Answer: "山田 太郎", Comment: "氏名"},
		{QuestionID: "237993201", Answer: "__other_option__", Comment: "ラジオボタンのその他を選択"},
		{QuestionID: "237993201.other_option_response", Answer: "テキスト", Comment: "ラジオボタンのその他のテキスト"},
	}

	// 自動入力URLを生成
	encodedURL, err := createAutoFillURL(baseURL, formEntries)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// URLをデコード
	// デコードは表示のためだけに行う
	// (url.PathEscapeやurl.QueryEscapeでエンコードされた文字列をデコードする)
	decodedURL, err := url.QueryUnescape(encodedURL)
	if err != nil {
		fmt.Printf("警告: URLのデコードに失敗しました。詳細: %v\n", err)
		// エラーが発生しても、エンコードされたURLは有効なので処理を続行する
		decodedURL = encodedURL
	}

	// 出力
	fmt.Println("=== 自動入力用URL ===")
	fmt.Printf("デコードされたURL:\n%s\n\n", decodedURL)
	fmt.Printf("エンコードされたURL:\n%s\n\n", encodedURL)
}

// createAutoFillURL は、GoogleフォームのベースURLと入力項目から自動入力URLを生成します。
func createAutoFillURL(baseURL string, entries []FormEntry) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("URLの解析中にエラーが発生しました: %w", err)
	}

	// "usp=sf_link"のクエリストリングを削除
	queries := parsedURL.Query()
	queries.Del("usp")

	// "usp=pp_url"のクエリストリングを追加
	queries.Set("usp", "pp_url")

	// フォームの入力値をクエリストリングに追加
	for _, entry := range entries {
		value := entry.Answer
		if value == "{today}" {
			value = time.Now().Format("2006-01-02")
		}
		queries.Set(fmt.Sprintf("entry.%s", entry.QuestionID), value)
	}

	fmt.Println("クエリストリングの一覧")
	fmt.Println(queries)

	parsedURL.RawQuery = queries.Encode()

	return parsedURL.String(), nil
}
