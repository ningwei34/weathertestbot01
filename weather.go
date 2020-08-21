package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

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
	StationId    string `json:"stationId"`
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

	resp, _ := http.Get("https://opendata.cwb.gov.tw/api/v1/rest/datastore/O-A0003-001?Authorization=CWB-9A1DBFCC-F4B4-4083-9EE2-A241B193D707&locationName=高雄")
	defer resp.Body.Close()              //關閉連線
	body, _ := ioutil.ReadAll(resp.Body) //讀取body的內容
	fmt.Println(decoding(body))

}

func decoding(b []byte) string {

	var t StationObsResponse
	json.Unmarshal([]byte(b), &t)
	var weatherState string = ""

	nowWeather := t.Records.Location[0].WeatherElement
	fmt.Println(nowWeather)

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
}
