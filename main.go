package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	ubot "github.com/UBotPlatform/UBot.Common.Go"
)

var api *ubot.AppApi

type AnimalImageResponse struct {
	Link string `json:"link"`
}
type AnimalNameMapping struct {
	keywords   []string
	animalType string
}

var mappings = [...]AnimalNameMapping{
	{keywords: []string{"/cat", "吸猫"}, animalType: "cat"},
	{keywords: []string{"/dog", "狗狗"}, animalType: "dog"},
	{keywords: []string{"/red_panda", "小熊猫", "红熊猫", "小猫熊", "红猫熊", "金狗"}, animalType: "red_panda"},
	{keywords: []string{"/panda", "熊猫", "猫熊"}, animalType: "panda"},
	{keywords: []string{"/bird", "小鸟", "鸟图"}, animalType: "birb"},
	{keywords: []string{"/fox", "狐狸"}, animalType: "fox"},
	{keywords: []string{"/koala", "考拉", "无尾熊", "可拉熊", "树懒熊"}, animalType: "koala"},
}

var cachedAnimalImages = make(map[string][]string)

func onReceiveChatMessage(bot string, msgType ubot.MsgType, source string, sender string, message string, info ubot.MsgInfo) (ubot.EventResultType, error) {
	animalType := ""
mappingLoop:
	for _, mapping := range mappings {
		for _, keyword := range mapping.keywords {
			if strings.Contains(message, keyword) {
				animalType = mapping.animalType
				break mappingLoop
			}
		}
	}
	if animalType == "" {
		return ubot.IgnoreEvent, nil
	}
	for {
		var builder ubot.MsgBuilder
		resp, err := http.Get(fmt.Sprintf("https://some-random-api.ml/img/%s?t=%d", animalType, time.Now().Unix()))
		if err != nil {
			break
		}
		defer resp.Body.Close()
		var result AnimalImageResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			cachedList := cachedAnimalImages[animalType]
			if len(cachedList) == 0 {
				_ = api.SendChatMessage(bot, msgType, source, sender, "请求图片失败，可能是网络问题或频率限制")
				break
			}
			result.Link = cachedList[time.Now().Unix()%int64(len(cachedList))]
		} else {
			if len(cachedAnimalImages[animalType]) >= 100 {
				cachedAnimalImages[animalType][time.Now().Unix()%100] = result.Link
			} else {
				cachedAnimalImages[animalType] = append(cachedAnimalImages[animalType], result.Link)
			}
		}
		builder.WriteEntity(ubot.MsgEntity{Type: "image", Args: []string{result.Link}})
		_ = api.SendChatMessage(bot, msgType, source, sender, builder.String())
		break //nolint
	}
	return ubot.CompleteEvent, nil
}

func main() {
	err := ubot.HostApp("AnimalImage", func(e *ubot.AppApi) *ubot.App {
		api = e
		return &ubot.App{
			OnReceiveChatMessage: onReceiveChatMessage,
		}
	})
	ubot.AssertNoError(err)
}
