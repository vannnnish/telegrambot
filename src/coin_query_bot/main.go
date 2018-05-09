/**
 * Created by vannnnish on 27/04/2018.
 * Copyright © 2018年 yeeyuntech. All rights reserved.
 */

package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/yeeyuntech/yeego/yeeStrconv"
	"sync"
	"fmt"
	"strings"
	"coin_query_bot/entity"
	"coin_query_bot/task"
	"github.com/astaxie/beego/toolbox"
	"gitlab.yeeyuntech.com/yee/easyweb"
	"github.com/yeeyuntech/yeego"
	"time"
	"github.com/yeeyuntech/yeego/yeeHttp"
	"encoding/json"
)

func main() {
	globalConfig()
	//fmt.Println("exchangeSet:", result)
	token := yeego.Config.GetString("app.Token")
	Init()
	task.InitTask()
	toolbox.StartTask()
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		// 这里里面是向订阅号发送消息
		if update.ChannelPost != nil {
			// 获取要返回的数据结构
			message, ok := getFormatData(strings.ToUpper(update.ChannelPost.Text))
			if !ok {
				continue
			}
			updateTimeI, ok := entity.MarketIno.Load("updateTime")
			if !ok {
				continue
			}
			updateTime := updateTimeI.(int64)
			timeStr := time.Unix(updateTime, 0).Format("01-02 15:04")
			message = message + fmt.Sprintf(entity.FormatTail, timeStr)
			msg := tgbotapi.NewMessageToChannel(yeeStrconv.FormatInt(int(update.ChannelPost.Chat.ID)), message)
			msg.ReplyToMessageID = update.ChannelPost.MessageID
			bot.Send(msg)
			continue
		}

		// 获取要返回的数据结构
		message, ok := getFormatData(strings.ToUpper(update.Message.Text))
		if !ok {
			message, ok := getGlobalData(strings.ToUpper(update.Message.Text))
			if !ok {
				continue
			}
			message = message + fmt.Sprintf(entity.FormatTail, time.Now().Format("01-02 15:04"))
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			continue
		}
		// updateTimeI, _ := entity.MarketIno.Load("updateTime")
		// message = message + fmt.Sprintf(entity.FormatTail, time.Now().Format("01-02 15:04"))
		// fmt.Println("这个时间:", updateTimeI)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	}
}

// 获取全网数据
func getGlobalData(symbol string) (string, bool) {
	// 获取人民币的汇率
	cnyFinanceInterface, ok := entity.FinalRateSyncMap.Load("CNY")
	var cnyFinance float64
	if !ok {
		cnyFinance = 6.5
	} else {
		cnyFinance = cnyFinanceInterface.(float64)
	}
	var returnData = fmt.Sprintf(entity.MessageHead, entity.CoinMap[symbol], symbol)
	url := fmt.Sprintf(entity.SingleCoinInfoApi, entity.CoinMap[symbol])
	data, err := yeeHttp.Get(url).Exec().ToBytes()
	if err != nil {
		yeego.Print("全网数据获取失败:", err)
		return returnData, false
	}
	var coinInfo entity.SingleCoin
	err = json.Unmarshal(data, &coinInfo)
	if err != nil {
		yeego.Print("json解析失败:", err)
		return returnData, false
	}
	priceUsd := yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Price_usd) * cnyFinance
	percentChange := yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Percent_change_24h) * 100
	if priceUsd == 0 {
		return returnData, false
	}
	returnData = returnData + fmt.Sprintf(entity.GlobalMessageCNY, symbol, priceUsd, percentChange)
	return returnData, true
}

// 最初的处理
func getFormatData(symbol string) (string, bool) {
	// 获取人民币的汇率
	cnyFinanceInterface, ok := entity.FinalRateSyncMap.Load("CNY")
	var cnyFinance float64
	if !ok {
		cnyFinance = 6.5
	} else {
		cnyFinance = cnyFinanceInterface.(float64)
	}
	var returnData = fmt.Sprintf(entity.MessageHead, entity.CoinMap[symbol], symbol)
	result, ok := entity.MarketIno.Load(symbol)
	// symbolName,ok:=key.(string)
	// if !ok{
	//	return true
	// 	}
	if !ok {
		return returnData, false
	}
	exInfos, ok := result.([]entity.ExchangePairInfo)
	if !ok {
		return returnData, false
	}
	for _, exInfo := range exInfos {
		returnData = returnData + fmt.Sprintf(entity.FormatMessageCNY, exInfo.ExchangeName, symbol, exInfo.QuoteSymbol, exInfo.Price*cnyFinance, exInfo.Change24h*100, exInfo.CashIn*cnyFinance)
	}

	return returnData, true
}

/*func getFormatData(symbol string) string {
	var returnData = entity.MESSAGE_HEADER
	for k, v := range entity.ParimaryExhcnageNameMap {
		var price entity.CoinPriceInfo
		var cash entity.CashIn
		pairData, err := yeeHttp.Get(fmt.Sprintf(entity.PairApi, entity.CoinIdMap[symbol], k)).Exec().ToBytes()
		if err != nil {
			log.Println("请求错误:", err)
			continue
		}
		err = json.Unmarshal(pairData, &price)
		if err != nil {
			log.Println("请求错误:", err)
			continue
		}
		cashInInfo, err := yeeHttp.Get(fmt.Sprintf(entity.CashInApi, k, symbol)).Exec().ToBytes()
		if err != nil {
			log.Println("请求错误:", err)
			continue
		}
		err = json.Unmarshal(cashInInfo, &cash)

		if err != nil {
			log.Println("请求错误:", err)
			continue
		}
		if price.Data.Price_usd == 0 || cash.Data.Data.Sell_vol_usd == 0 {
			continue
		}
		percentChange24 := price.Data.Percent_change_today * 100
		pureCashIn := cash.Data.Data.Buy_vol_usd - cash.Data.Data.Sell_vol_usd
		fmt.Println("价格:", price.Data.Price_usd, "24小时变化:", percentChange24, "进流入:", pureCashIn)
		returnData = returnData + fmt.Sprintf(entity.FormatMessage, v, symbol, price.Data.Price_usd, price.Data.Percent_change_today*100, cash.Data.Data.Buy_vol_usd)
		fmt.Println("返回的结果:", returnData)
	}

	return returnData
}*/

// 获取主流币种的信息
func UpdateCoinInfo(m *sync.Map) {
	// 获取币种信息

}

// 一些全局配置
func globalConfig() {
	yeego.MustInitConfig("conf", "conf")
	conf := easyweb.DefaultConfig()
	conf.TemplateDir = yeego.Config.GetString("app.TemplateDir")
	conf.Mode = yeego.Config.GetString("app.RunMode")
	conf.Pprof = yeego.Config.GetBool("app.Pprof")
	conf.SessionOn = yeego.Config.GetBool("app.SessionOn")
	conf.Port = yeego.Config.GetString("app.Port")
	easyweb.SetConfig(conf)
}
