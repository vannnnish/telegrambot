/**
 * Created by vannnnish on 2018/5/25.
 * Copyright © 2018年 yeeyuntech. All rights reserved.
 */

package notice

import (
	"coin_query_bot/entity"
	"github.com/yeeyuntech/yeego/yeeHttp"
	"github.com/yeeyuntech/yeego"
	"github.com/yeeyuntech/yeego/yeeStrconv"
	"fmt"
	"encoding/json"
	"time"
)

// 获取全网数据
func GetGlobalData(symbol string) (string, bool) {
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
	data, err = yeeHttp.Get(fmt.Sprintf(entity.GlobalCoinFinance, entity.CoinMap[symbol])).Exec().ToBytes()
	if err != nil {
		yeego.Print("全网数据获取失败:", err)
		return returnData, false
	}
	var cashInfo entity.CashIn
	err = json.Unmarshal(data, &cashInfo)
	if err != nil {
		yeego.Print("json解析失败:", err)
	}
	pureCashIn := (cashInfo.Data.Data.Buy_vol_usd - cashInfo.Data.Data.Sell_vol_usd) * cnyFinance / 10000
	priceCny := yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Price_usd) * cnyFinance
	percentChange := yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Percent_change_24h) * 100
	if priceCny == 0 {
		return returnData, false
	}
	if err != nil || (cashInfo.Data.Data.Buy_vol_usd == 0 && cashInfo.Data.Data.Sell_vol_usd == 0) {
		returnData = fmt.Sprintf(entity.NewFormatFront+entity.NewFormatTail, entity.CoinMap[symbol],
			symbol, coinInfo.Data.Coin.Rank, priceCny, yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Percent_change_1h)*100, percentChange, time.Now().Format("15:04:03"))
		return returnData, true
	} else {
		returnData = fmt.Sprintf(entity.NewFormatFront+entity.NewFormatMiddle+entity.NewFormatTail, entity.CoinMap[symbol],
			symbol, coinInfo.Data.Coin.Rank, priceCny, yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Percent_change_1h)*100, percentChange, pureCashIn, time.Now().Format("15:04:03"))
		return returnData, true
	}
}

// 获取全网数据
func GetGlobalDataEnglish(symbol string) (string, bool) {
	// 获取人民币的汇率

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
	data, err = yeeHttp.Get(fmt.Sprintf(entity.GlobalCoinFinance, entity.CoinMap[symbol])).Exec().ToBytes()
	if err != nil {
		yeego.Print("全网数据获取失败:", err)
		return returnData, false
	}
	var cashInfo entity.CashIn
	err = json.Unmarshal(data, &cashInfo)
	if err != nil {
		yeego.Print("json解析失败:", err)
	}
	pureCashIn := (cashInfo.Data.Data.Buy_vol_usd - cashInfo.Data.Data.Sell_vol_usd)
	priceCny := yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Price_usd)
	percentChange := yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Percent_change_24h) * 100
	if priceCny == 0 {
		return returnData, false
	}
	if err != nil || (cashInfo.Data.Data.Buy_vol_usd == 0 && cashInfo.Data.Data.Sell_vol_usd == 0) {
		returnData = fmt.Sprintf(entity.NewFormatFrontEnglish+entity.NewFormatTailEnglish, entity.CoinMap[symbol],
			symbol, coinInfo.Data.Coin.Rank, priceCny, yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Percent_change_1h)*100, percentChange, time.Now().Format("15:04:03"))
		return returnData, true
	} else {
		returnData = fmt.Sprintf(entity.NewFormatFrontEnglish+entity.NewFormatMiddleEnglish+entity.NewFormatTailEnglish, entity.CoinMap[symbol],
			symbol, coinInfo.Data.Coin.Rank, priceCny, yeeStrconv.ParseFloat64Default0(coinInfo.Data.Coin.Percent_change_1h)*100, percentChange, pureCashIn, time.Now().Format("15:04:03"))
		return returnData, true
	}
}

// 最初的处理
func GetFormatData(symbol string) (string, bool) {
	var returnData string
	// 获取人民币的汇率
	cnyFinanceInterface, ok := entity.FinalRateSyncMap.Load("CNY")
	var cnyFinance float64
	if !ok {
		cnyFinance = 6.5
	} else {
		cnyFinance = cnyFinanceInterface.(float64)
	}
	result, ok := entity.MarketIno.Load(symbol)
	// symbolName,ok:=key.(string)
	// if !ok{
	//	return true
	// 	}
	if !ok {
		return returnData, false
	}
	exInfos, ok := result.([]entity.ExchangePairInfo)
	if len(exInfos) == 0 {
		return returnData, false
	}
	returnData = fmt.Sprintf(entity.MessageHeadGlobal, entity.CoinMap[symbol], symbol, exInfos[0].GlobalCashIn*cnyFinance)
	if !ok {
		return returnData, false
	}
	for _, exInfo := range exInfos {
		returnData = returnData + fmt.Sprintf(entity.FormatMessageCNY, exInfo.ExchangeName, symbol, exInfo.QuoteSymbol, exInfo.Price*cnyFinance, exInfo.Change24h*100, exInfo.CashIn*cnyFinance)
	}

	return returnData, true
}
