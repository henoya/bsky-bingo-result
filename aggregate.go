package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

type Rankings []*RankingRow

type RankingRow struct {
	DateStr        string
	Rank           int
	UserHandleUrl  string
	UserHandle     string
	UserName       string
	UserPoint      int
	UserHistoryUri string
}

type PersonalScoreRow struct {
	Id            int     `gorm:"primary_key"`
	UserHandleUrl string  `gorm:"not null"`
	DateTimeStr   string  `gorm:"not null"`
	Category      string  `gorm:"not null"`
	Point         float32 `gorm:"not null"`
}

type DayRankingTableRow struct {
	Id             int    `gorm:"primary_key;AUTO_INCREMENT"`
	DateStr        string `gorm:"not null"`
	Rank           int    `gorm:"not null"`
	UserHandleUrl  string `gorm:"not null"`
	UserHandle     string `gorm:"null"`
	UserName       string `gorm:"null"`
	UserPoint      int    `gorm:"not null"`
	UserHistoryUri string `gorm:"not null"`
}

type MonthryRankingTableRow struct {
	MonthStr       string `gorm:"not null;primary_key"`
	Rank           int    `gorm:"not null;primary_key"`
	UserHandleUrl  string `gorm:"not null"`
	UserHandle     string `gorm:"null"`
	UserName       string `gorm:"null"`
	UserPoint      int    `gorm:"not null"`
	UserHistoryUri string `gorm:"not null"`
}

// doAggregate Bingoゲームのランキング集計をおこなう
func doImport(cCtx *cli.Context) (err error) {
	importUrlBase := "https://bingo.b35.jp/view_ranking.php?m="
	importUrlLists := []string{
		"202306",
		"202307",
		"202308",
	}
	// DBファイルのオープン
	var db *gorm.DB
	db, err = openDB("bingo.db")
	if err != nil {
		return fmt.Errorf("failed to connect database")
	}

	err = migrateDB(db)
	if err != nil {
		return fmt.Errorf("failed to migrate database")
	}

	for _, monthStr := range importUrlLists {
		// 月ごとのランキングURLを取得
		monthRankingUrl, err := url.JoinPath(importUrlBase, monthStr)
		if err != nil {
			return fmt.Errorf("failed to join url: %s", err)
		}
		// 月ごとのランキングURLをGET
		currentHtml, err := getUrlContents(monthRankingUrl)
		if err != nil {
			return err
		}
		fmt.Printf("currentHtml: %s\n", string(currentHtml))
	}
	return nil
}

// doAggregate Bingoゲームのランキング集計をおこなう
func doAggregate(cCtx *cli.Context) (err error) {
	//if cCtx.Args().Present() {
	//	return cli.ShowSubcommandHelp(cCtx)
	//}

	// DBファイルのオープン
	var db *gorm.DB
	db, err = openDB("bingo.db")
	if err != nil {
		return fmt.Errorf("failed to connect database")
	}

	err = migrateDB(db)
	if err != nil {
		return fmt.Errorf("failed to migrate database")
	}

	// Blueskyログイン用のアカウントとApp Passを取得する
	bsky_acount, exists := os.LookupEnv("BSKY_ACCOUNT")
	if !exists {
		return err
	}
	bsky_apppass, exists := os.LookupEnv("BSKY_APPPASS")
	if !exists {
		return err
	}

	pwd := os.Getenv("PWD")

	// 結果保存用の result ディレクトリの存在を確認
	resultDir := path.Join(pwd, "result")
	dirInfo, err := os.Stat(resultDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(resultDir, 0755)
		if err != nil {
			return err
		}
	} else if !dirInfo.IsDir() {
		fmt.Errorf("%s is not a directory\n", resultDir)
	}

	// Bingo ゲームのページのベースURL
	bingoBaseUri := "https://bingo.b35.jp"
	// Bingo ゲームのランキングページ
	currentRankingPageUrl, err := url.JoinPath(bingoBaseUri, "view_ranking.php")
	if err != nil {
		return err
	}

	// 日付文字列の取得
	now := time.Now().Local()
	todayStr := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1)
	yesterdayStr := yesterday.Format("2006-01-02")

	fmt.Printf("BSKY_ACCOUNT: %s\n", bsky_acount)
	fmt.Printf("BSKY_APPPASS: %s\n", bsky_apppass)
	fmt.Printf("RESULT_DIR: %s\n", resultDir)
	fmt.Printf("TODAY: %s\n", todayStr)
	fmt.Printf("YESTERDAY: %s\n", yesterdayStr)

	// 昨日のランキング取得
	currentRankingFile := "bingo_ranking.html"
	_ = currentRankingFile
	currentRankingHtml, err := getUrlContents(currentRankingPageUrl)
	if err != nil {
		return err
	}
	fmt.Printf("currentRankingHtml: %s\n", string(currentRankingHtml))

	// ランキングのhtmlからランキングのtsvを生成
	rankingTsv, err := generateRankingTsv(string(currentRankingHtml), yesterdayStr)
	if err != nil {
		return err
	}
	for i, row := range *rankingTsv {
		row.UserHistoryUri, err = url.JoinPath(bingoBaseUri, row.UserHistoryUri)
		if err != nil {
			return err
		}
		(*rankingTsv)[i] = row
	}

	// Bluesky から、昨日の結果のポスト検索
	// Bluesky にログイン
	_, err = execLogin(cCtx, "bsky.app", bsky_acount, bsky_apppass)
	if err != nil {
		return err
	}

	// Bingo ゲームアカウントのタイムライン取得
	//bingoAccountTl := "${temp_dir}/bingo_account_tl.json"
	//bingoResultPostJson="${temp_dir}/bingo_result_post.json"
	//bingoResultPostJsonTmp="${bingo_result_post_json}.tmp"
	//bsky tl -H "${bingo_account_handle}" -n 100 -json > "${bingo_account_tl}"
	bingoAccountHandle := "bingo.b35.jp"

	feed, err := execTimeline(cCtx, bingoAccountHandle, 100)
	if err != nil {
		return err
	}
	for _, row := range feed {
		json.NewEncoder(os.Stdout).Encode(row)
	}
	return nil
}

func generateRankingTsv(html string, dateStr string) (rankingList *Rankings, err error) {
	regexRankRow := regexp.MustCompile(`<td class="rank(| p[0-9]*)">([0-9]+)</td>`)
	regexUserRow := regexp.MustCompile(`<td class="user"><a href="([^"]+)">(.*)</td>`)
	regexUserAndTitleRow := regexp.MustCompile(`<td class="user"><a href="([^"]+)" title="(.*)">(.*)</td>`)
	regexPointRow := regexp.MustCompile(`<td class="point">([0-9\.]+)</td>`)
	regexHistoryRow := regexp.MustCompile(`<td class="history"><a href="(.+)">履歴</td>`)
	regexTrRow := regexp.MustCompile(`</tr>`)
	// header出力
	rankRow := RankingRow{}
	ranking := make(Rankings, 0)
	for _, line := range strings.Split(html, "\n") {
		match := regexRankRow.FindAllStringSubmatch(line, -1)
		if len(match) > 0 {
			rank, err := strconv.Atoi(match[1][2])
			if err != nil {
				return nil, err
			}
			rankRow = RankingRow{
				DateStr:        dateStr,
				Rank:           rank,
				UserHandleUrl:  "",
				UserHandle:     "",
				UserName:       "",
				UserPoint:      0,
				UserHistoryUri: "",
			}
		}
		match = regexUserRow.FindAllStringSubmatch(line, -1)
		if len(match) > 0 {
			rankRow.UserHandleUrl = match[1][1]
			rankRow.UserHandle = match[1][2]
			rankRow.UserName = match[1][2]
		}
		match = regexUserAndTitleRow.FindAllStringSubmatch(line, -1)
		if len(match) > 0 {
			rankRow.UserHandleUrl = match[1][1]
			rankRow.UserHandle = match[1][2]
			rankRow.UserName = match[1][3]
		}
		match = regexPointRow.FindAllStringSubmatch(line, -1)
		if len(match) > 0 {
			userPoint, err := strconv.Atoi(match[1][1])
			if err != nil {
				return nil, err
			}
			rankRow.UserPoint = userPoint
		}
		match = regexHistoryRow.FindAllStringSubmatch(line, -1)
		if len(match) > 0 {
			userHistoryUri := match[1][1]
			rankRow.UserHistoryUri = userHistoryUri
		}
		match = regexTrRow.FindAllStringSubmatch(line, -1)
		if len(match) > 0 {
			if rankRow.Rank > 0 {
				ranking = append(ranking, &rankRow)
				rankRow = RankingRow{
					DateStr:        dateStr,
					Rank:           0,
					UserHandleUrl:  "",
					UserHandle:     "",
					UserName:       "",
					UserPoint:      0,
					UserHistoryUri: "",
				}
			}
		}
	}
	return &ranking, nil
}

func getUrlContents(url string) (body []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
