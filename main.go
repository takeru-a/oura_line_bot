package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// 活動量
type DailyActivityResponse struct {
    Data      []Activity `json:"data"`
    NextToken *string    `json:"next_token"`
}
type Activity struct {
	ActiveScore int `json:"score"` // アクティブスコア
	Calories    int `json:"total_calories"` // 総消費カロリー
	NonWearTime int `json:"non_wear_time"` // 着用していない時間
}

// 睡眠データ
type SleepResponse struct {
    Data      []Sleep `json:"data"`
    NextToken *string      `json:"next_token"`
}

type Sleep struct {
    BedtimeStart          time.Time             `json:"bedtime_start"` // 就寝時間
	BedtimeEnd            time.Time             `json:"bedtime_end"` // 起床時間
    DeepSleepDuration     int                   `json:"deep_sleep_duration"` // 深い睡眠時間
	LightSleepDuration    int                   `json:"light_sleep_duration"` // 浅い睡眠時間
	RemSleepDuration      int                   `json:"rem_sleep_duration"` // REM睡眠時間
	TotalSleepDuration    int                   `json:"total_sleep_duration"`// 合計睡眠時間
    Efficiency            int                   `json:"efficiency"` // 睡眠効率
    LowBatteryAlert       bool                  `json:"low_battery_alert"` // 低バッテリーアラート
}

// 日々の睡眠スコア
type SleepScoreResponse struct {
    Data      []SleepScore `json:"data"`
    NextToken *string      `json:"next_token"`
}

type SleepScore struct {
    Contributors SleepContributors  `json:"contributors"`
    Score        int               `json:"score"` // 睡眠スコア
}

type SleepContributors struct {
    DeepSleep    int `json:"deep_sleep"` // 深い睡眠のスコア
    RemSleep     int `json:"rem_sleep"`  // REM睡眠のスコア
    Restfulness  int `json:"restfulness"`// 安眠度のスコア
    TotalSleep   int `json:"total_sleep"`// 合計睡眠時間のスコア
}

// LINEメッセージ
type TextMessage struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

type PushRequest struct {
    To       string        `json:"to"`
    Messages []TextMessage `json:"messages"`
}

// データがない場合のメッセージ
var NoDataMessage string = "⚠データが取得できませんでした!\n\nデータが同期されていない可能性があります。\n\nアプリ上でデータの連携状況\n及び、リングの残バッテリー量を確認してください!"

// 環境変数が設定されているかを確認
func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf(".env ファイルが読み込めませんでした: %v", err)
	}
}

func main() {

	// 環境変数から取得
	ouraApiToken := os.Getenv("OURA_API_TOKEN")
	if ouraApiToken == "" {
		log.Fatal("OURA_API_TOKEN が設定されていません")
	}

	lineApiToken := os.Getenv("LINE_API_TOKEN")
	if lineApiToken == "" {
		log.Fatal("LINE_API_TOKEN が設定されていません")
	}

	toLineUserId := os.Getenv("TO_LINE_USER")
	if toLineUserId == "" {
		log.Fatal("TO_LINE_USER が設定されていません")
	}
	// OURA APIのエンドポイントを指定
	ouraApiDailyActivityUrl := "https://api.ouraring.com/v2/usercollection/daily_activity"
	ouraApiSleepUrl := "https://api.ouraring.com/v2/usercollection/sleep"
	ouraApiSleepScoreUrl := "https://api.ouraring.com/v2/usercollection/daily_sleep"

	// 日付を取得
	yesterdayDate := time.Now().Add(-24 * time.Hour) // 昨日の日付を取得
	yesterdayDateFormed := formatDate(yesterdayDate)

	todayDate := time.Now()
	todayDateFormed := formatDate(todayDate)

	// 活動量データを取得
	acRes, err := fetchOuraData(ouraApiDailyActivityUrl, ouraApiToken, yesterdayDateFormed, todayDateFormed)
	if err != nil {
		log.Fatal("活動量データの取得に失敗しました:", err)
	}

	// 睡眠データを取得
	sleepRes, err := fetchOuraData(ouraApiSleepUrl, ouraApiToken, yesterdayDateFormed, todayDateFormed)
	if err != nil {
		log.Fatal("睡眠データの取得に失敗しました:", err)
	}

	/// 睡眠スコアデータを取得
	sleepScoreRes, err := fetchOuraData(ouraApiSleepScoreUrl, ouraApiToken, todayDateFormed, todayDateFormed)
	if err != nil {
		log.Fatal("日々の睡眠スコアデータの取得に失敗しました:", err)
	}

	// 取得したデータをJSON形式にデコード
	var activityRes DailyActivityResponse
	if err := json.Unmarshal(acRes, &activityRes); err != nil {
		log.Fatal("JSONのデコードに失敗しました:", err)
	}
	// データがない場合はメッセージを送信して終了
	if len(activityRes.Data) == 0 {
		if err := sendLineMessage(lineApiToken, toLineUserId, NoDataMessage); err != nil {
			log.Fatal("LINEメッセージの送信に失敗しました:", err)
		}
		return
	}
	activity := activityRes.Data[len(activityRes.Data)-1]

	// 睡眠データをJSON形式にデコード
	var sleepData SleepResponse
	if err := json.Unmarshal(sleepRes, &sleepData); err != nil {
		log.Fatal("JSONのデコードに失敗しました:", err)
	}
	// データがない場合はメッセージを送信して終了
	if len(sleepData.Data) == 0 {
		if err := sendLineMessage(lineApiToken, toLineUserId, NoDataMessage); err != nil {
			log.Fatal("LINEメッセージの送信に失敗しました:", err)
		}
		return
	}
	sleep := sleepData.Data[len(sleepData.Data)-1]

	// 睡眠スコアデータをJSON形式にデコード
	var sleepScoreData SleepScoreResponse
	if err := json.Unmarshal(sleepScoreRes, &sleepScoreData); err != nil {
		log.Fatal("JSONのデコードに失敗しました:", err)
	}
	// データがない場合はメッセージを送信して終了
	if len(sleepScoreData.Data) == 0 {
		if err := sendLineMessage(lineApiToken, toLineUserId, NoDataMessage); err != nil {
			log.Fatal("LINEメッセージの送信に失敗しました:", err)
		}
		return
	}
    sleepScore := sleepScoreData.Data[len(sleepScoreData.Data)-1]

	// LINEに送信するメッセージを作成
	message := fmt.Sprintf(
		"■低バッテリーアラート:\n"+		
		"	%s\n"+
		"\n"+
		"■睡眠データ:\n"+
		"  ・合計睡眠時間:\n　　　　　%s\n"+
		"  ・睡眠効率:		%d%%\n"+
		"  ・就寝時間:\n		　　　　%s\n"+
		"  ・起床時間:\n		　　　　%s\n"+
		"  ・深い睡眠時間:\n		　　　　　%s\n"+
		"  ・浅い睡眠時間:\n		　　　　　%s\n"+
		"  ・REM睡眠時間:\n		　　　　　%s\n"+
		"	\n"+
		"■睡眠スコアデータ:\n"+
		"  ・睡眠スコア:		%d/100\n"+
		"  ・安眠度のスコア:		%d/100\n"+
		"  ・睡眠時間のスコア:		%d/100\n"+
		"  ・深い睡眠のスコア:		%d/100\n"+
		"  ・REM睡眠のスコア:		%d/100\n"+
		"	\n"+
		"■活動量データ:\n"+
		"  ・アクティブスコア:		%d/100\n"+
		"  ・総消費カロリー:		%d kcal\n"+
		"  ・着用していない時間:\n		　　　　　%s",
			getBatteryStatusMessage(sleep.LowBatteryAlert),

			formatSecondsToTime(sleep.TotalSleepDuration),
			sleep.Efficiency,
			sleep.BedtimeStart.Format("2006-01-02 15:04:05"),
			sleep.BedtimeEnd.Format("2006-01-02 15:04:05"),
			formatSecondsToTime(sleep.DeepSleepDuration),
			formatSecondsToTime(sleep.LightSleepDuration),
			formatSecondsToTime(sleep.RemSleepDuration),

			sleepScore.Score,
			sleepScore.Contributors.Restfulness,
			sleepScore.Contributors.TotalSleep,
			sleepScore.Contributors.DeepSleep,
			sleepScore.Contributors.RemSleep,

			activity.ActiveScore,
			activity.Calories,
			formatSecondsToTime(activity.NonWearTime),
		)
  
	// LINEにメッセージを送信
	if err := sendLineMessage(lineApiToken, toLineUserId, message); err != nil {
		log.Fatal("LINEメッセージの送信に失敗しました:", err)
	}
}

// OURA_APIの呼び出し処理
func fetchOuraData(endpoint, token, startdate, enddate string) ([]byte, error){

	// OURA APIのエンドポイントを指定
	url := fmt.Sprintf("%s?start_date=%s&end_date=%s", endpoint, startdate, enddate)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

	// OURA APIのトークンをヘッダーに設定
    req.Header.Set("Authorization", "Bearer "+token)
    restClient := &http.Client{Timeout: 10 * time.Second}

    res, err := restClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("status %d from %s", res.StatusCode, endpoint)
    }
    return io.ReadAll(res.Body)
}

// LINEにメッセージを送信する関数
func sendLineMessage(token, userId, message string) error {
	
	url := "https://api.line.me/v2/bot/message/push"
	payload := PushRequest{
		To: userId,
		Messages: []TextMessage{
			{
				Type: "text",
				Text: message,
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// リクエストの作成
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	restClient := &http.Client{}
	resp, err := restClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d from LINE API", resp.StatusCode)
	}

	return nil
}


// 日付を整形する関数
func formatDate(date time.Time) string {
	loc, err := time.LoadLocation("Asia/Tokyo")
    if err != nil {
        log.Fatal("タイムゾーンの取得に失敗しました:", err)
    }
	JPDate := date.In(loc)
	return JPDate.Format("2006-01-02")
}

// 秒数を時間に変換する関数
func formatSecondsToTime(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02dh:%02dm:%02ds", hours, minutes, secs)
}

// バッテリーの状態のメッセージを生成する関数
func getBatteryStatusMessage(batterystatus bool) string {
	if batterystatus {
		return "バッテリー残量が少なくなっています。"
	} else {
		return "バッテリー残量は十分です。"
	}
}