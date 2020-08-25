package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"

	"encoding/json"
	"io/ioutil"
	"strconv"
)

var bot *linebot.Client

type StationObsResponse struct {
	Success string `json:"success"`
	Records struct {
		Location []StationObsLocation `json:"location"`
	} `json:"records"`
}

type StationObsLocation struct {
	Lat          string `json:"lat"`
	Lon          string `json:"lon"`
	LocationName string `json:"locationName"`
	StationID    string `json:"stationId"`
	Time         struct {
		ObsTime string `json:"obsTime"`
	} `json:"time"`
	WeatherElement []StationObsElement `json:"weatherElement"`
}

type StationObsElement struct {
	ElementName  string `json:"elementName"`
	ElementValue string `json:"elementValue"`
}

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)

}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:

				resp, _ := http.Get("https://opendata.cwb.gov.tw/api/v1/rest/datastore/O-A0003-001?Authorization=CWB-9A1DBFCC-F4B4-4083-9EE2-A241B193D707&locationName=高雄")
				defer resp.Body.Close()              //關閉連線
				body, _ := ioutil.ReadAll(resp.Body) //讀取body的內容
				fmt.Println(decoding(body))

				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Hello"+message.Text+"\n"+decoding(body)+"\n"+event.Source.UserID)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func decoding(b []byte) string {

	var t StationObsResponse
	json.Unmarshal([]byte(b), &t)
	var weatherState string = ""
	nowWeather := t.Records.Location[0].WeatherElement

	for _, i := range nowWeather {
		if i.ElementValue != "-99" {
			switch i.ElementName {
			case "TEMP":
				weatherState += "溫度:" + i.ElementValue[0:len([]rune(i.ElementValue))-1] + "°C\n"
			case "HUMD":
				hm, err := strconv.ParseFloat(i.ElementValue, 64)
				if err == nil {
					hm = hm * 100
					weatherState += "相對溼度:" + fmt.Sprintf("%.0f", hm) + "%\n"
				}
			case "SUN":
				weatherState += "日照時數:" + i.ElementValue + "H\n"
			case "H_UVI":
				uvi, err := strconv.ParseFloat(i.ElementValue, 64)
				if err == nil {
					if uvi == 0 {
					} else if uvi <= 2 {
						weatherState += "紫外線指數:" + i.ElementValue + " (低量)\n"
					} else if uvi <= 5 {
						weatherState += "紫外線指數:" + i.ElementValue + " (中量)\n"
					} else if uvi <= 7 {
						weatherState += "紫外線指數:" + i.ElementValue + " (高量)\n"
					} else if uvi <= 10 {
						weatherState += "紫外線指數:" + i.ElementValue + " (過量)\n"
					} else {
						weatherState += "紫外線指數:" + i.ElementValue + " (危險)\n"
					}
				}
			case "24R":
				rain, err := strconv.ParseFloat(i.ElementValue, 64)
				if err == nil && rain != 0 {
					weatherState += "累積雨量:" + i.ElementValue + " ml\n"
				}

			case "D_TX":
				weatherState += "最高溫:" + i.ElementValue[0:len([]rune(i.ElementValue))-1] + "°C\n"
			case "D_TN":
				weatherState += "最低溫:" + i.ElementValue[0:len([]rune(i.ElementValue))-1] + "°C\n"
			default:
			}
		}
	}
	getTime := t.Records.Location[0].Time.ObsTime
	weatherState += getTime[0 : len([]rune(getTime))-3]
	return weatherState

	// fmt.Println(bot.GetGroupMemberProfile)
	// fmt.Println(bot.IssueAccessToken)
	// fmt.Println(bot.GetGroupMemberIDs)

}
