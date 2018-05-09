/**
 * Created by vannnnish on 28/04/2018.
 * Copyright © 2018年 yeeyuntech. All rights reserved.
 */

package task

import (
	"coin_query_bot/entity"
	"github.com/yeeyuntech/yeego/yeeHttp"
	"log"
	"encoding/json"
	"github.com/kataras/go-errors"
	"fmt"
	"github.com/yeeyuntech/yeego"
	"time"
	"sync"
	"sort"
	"github.com/astaxie/beego/toolbox"
	"github.com/yeeyuntech/yeego/yeeStrconv"
)

const (
	ThreeMinuteConfig = "0 */3 * * * *"
	TwoMinuteConfig   = "0 */2 * * * *"
)

func InitTask() {
	toolbox.AddTask("update finance rate", toolbox.NewTask("update finance rate", TwoMinuteConfig, UpdateFinaceRate))
	toolbox.AddTask("update pair info", toolbox.NewTask("update pair info", ThreeMinuteConfig, UpdateCoin))
}

// 汇率更新接口
func UpdateFinaceRate() error {
	yeego.Print("开始更新汇率信息")
	startTime := time.Now()
	defer func(timer time.Time) {
		durTime := time.Since(startTime)
		yeego.Print(fmt.Sprintf("更新汇率用时:%.2f秒", durTime.Seconds()))
	}(startTime)
	data, err := yeeHttp.Get(entity.FinanceRateApi).Exec().ToBytes()
	if err != nil {
		return err
	}
	var financeRate entity.FinanceRate
	err = json.Unmarshal(data, &financeRate)
	if err != nil {
		return err
	}
	entity.FinalRateSyncMap.Store("CNY", yeeStrconv.ParseFloat64Default0(financeRate.Data.Legal_rate.CNY))
	return nil
}

// 缓存五个主要交易所的全部交易对信息
func UpdatePrimaryExchangePair() error {
	yeego.Print("开始缓存五家主流交易所全部交易对信息")
	startTime := time.Now()
	defer func(timer time.Time) {
		durTiem := time.Since(startTime)
		yeego.Print(fmt.Sprintf("缓存五家主流交易所全部交易对信息:%.2f秒", durTiem.Seconds()))
	}(startTime)
	// 先删除之前所有的记录
	for _, syncMap := range entity.ExchangeMaps {
		syncMap.Range(func(key, value interface{}) bool {
			syncMap.Delete(key)
			return true
		})
	}
	// 更新交易所的交易对，交易对按照BaseCoin进行分类。
	for k := range entity.PrimaryExhcnageNameMap {
		// 判断是不是该交易所第一次运行
		var exchangePair entity.AllExchangeInfo
		pairInfo, err := yeeHttp.Get(fmt.Sprintf(entity.AllExchangePairApi, k)).Exec().ToBytes()
		if err != nil {
			continue
		}
		err = json.Unmarshal(pairInfo, &exchangePair)
		if err != nil {
			yeego.Print("json格式化失败:", err)
			continue
		}
		for _, pairs := range exchangePair.Data.Data {
			var exchangePair entity.ExchangePair
			if 16 > len(pairs) {
				continue
			}
			exchangePair.ExchangeId = pairs[0].(string)
			exchangePair.Symbol = pairs[1].(string)
			exchangePair.QuoteSymbol = pairs[2].(string)
			exchangePair.PriceUsd = pairs[4].(float64)
			exchangePair.ChangeToday = pairs[15].(float64)
			exchangePair.VolumeUse24h = pairs[4].(float64) * pairs[6].(float64)
			switch k {
			case "huobi.pro":
				if !isDealWithSyncMapWell(&entity.HuoBiExchangePair, exchangePair.Symbol, exchangePair) {
					continue
				}
			case "zb":
				if !isDealWithSyncMapWell(&entity.ZbExchangePair, exchangePair.Symbol, exchangePair) {
					continue
				}
			case "binance":
				if !isDealWithSyncMapWell(&entity.BinanceExchangePair, exchangePair.Symbol, exchangePair) {
					continue
				}
			case "okex":
				if !isDealWithSyncMapWell(&entity.OKExExchangePair, exchangePair.Symbol, exchangePair) {
					continue
				}
			case "gateio":
				if !isDealWithSyncMapWell(&entity.GateIoExchangePair, exchangePair.Symbol, exchangePair) {
					continue
				}
			}

		}
	}
	return nil
}

func isDealWithSyncMapWell(p *sync.Map, symbol string, exchangePair entity.ExchangePair) bool {
	vaule, ok := p.Load(symbol)
	if !ok {
		var exchangePairs entity.SortByVolume24
		exchangePairs = append(exchangePairs, exchangePair)
		p.Store(symbol, exchangePairs)
	} else {
		exchangePairs, ok := vaule.(entity.SortByVolume24)
		if !ok {
			return ok
		}
		exchangePairs = append(exchangePairs, exchangePair)
		p.Store(symbol, exchangePairs)
	}
	return true
}

// 更新资金流入情况
func UpdateCoin() error {
	UpdatePrimaryExchangePair()
	yeego.Print("开始更新资金流入情况")
	startTime := time.Now()

	defer func(timer time.Time) {
		durTiem := time.Since(startTime)
		yeego.Print(fmt.Sprintf("开始更新资金流入情况:%.2f秒", durTiem.Seconds()))
	}(startTime)
	var okexPairMap = make(map[string]bool)

	entity.OKExExchangePair.Range(func(key, value interface{}) bool {
		// 币的symbol
		symbol := key.(string)
		okexPairMap[symbol] = true
		var returnDatas []entity.ExchangePairInfo
		pushExchangePairIn(key, returnDatas)
		return true
	})
	entity.HuoBiExchangePair.Range(func(key, value interface{}) bool {
		symbol := key.(string)
		if !okexPairMap[symbol] {
			var returnDatas []entity.ExchangePairInfo
			pushExchangePairIn(key, returnDatas)
		}
		return true
	})
	entity.BinanceExchangePair.Range(func(key, value interface{}) bool {
		symbol := key.(string)
		if !okexPairMap[symbol] {
			var returnDatas []entity.ExchangePairInfo
			pushExchangePairIn(key, returnDatas)
		}
		return true
	})
	// 币的symbol
	entity.ZbExchangePair.Range(func(key, value interface{}) bool {
		symbol := key.(string)
		if !okexPairMap[symbol] {
			var returnDatas []entity.ExchangePairInfo
			pushExchangePairIn(key, returnDatas)
		}
		return true
	})
	// 币的symbol
	entity.GateIoExchangePair.Range(func(key, value interface{}) bool {
		symbol := key.(string)
		if !okexPairMap[symbol] {
			var returnDatas []entity.ExchangePairInfo
			pushExchangePairIn(key, returnDatas)
		}
		return true
	})
	entity.MarketIno.Store("updateTime", time.Now().Unix())
	return nil
}

func pushExchangePairIn(key interface{}, returnDatas []entity.ExchangePairInfo) {
	for _, exchangeSyncMap := range entity.ExchangeMaps {
		var returnData entity.ExchangePairInfo
		var cash entity.CashIn
		exchange, ok := exchangeSyncMap.Load(key)
		if ok {
			sortedExchange := exchange.(entity.SortByVolume24)
			sort.Sort(sortedExchange)
			if len(sortedExchange) != 0 {
				cashInfo, err := yeeHttp.Get(fmt.Sprintf(entity.CashInApi, sortedExchange[0].ExchangeId, sortedExchange[0].Symbol, sortedExchange[0].QuoteSymbol)).Exec().ToBytes()
				if err != nil {
					return
				}
				err = json.Unmarshal(cashInfo, &cash)
				if err != nil {
					yeego.Print("json解析失败:", err)
					return
				}
				returnData.CashIn = cash.Data.Data.Buy_vol_usd - cash.Data.Data.Sell_vol_usd
				returnData.Price = sortedExchange[0].PriceUsd
				returnData.Change24h = sortedExchange[0].ChangeToday
				returnData.ExchangeName = entity.PrimaryExhcnageNameMap[sortedExchange[0].ExchangeId]
				returnData.Symbol = sortedExchange[0].Symbol
				returnData.QuoteSymbol = sortedExchange[0].QuoteSymbol
				returnDatas = append(returnDatas, returnData)
			}
		}
	}
	entity.MarketIno.Store(key, returnDatas)
	return
}

// 将coinid和coin的symbo对应
func UpdateAllCoins() error {
	yeego.Print("开始更新币的名称和Symbol对应的map")
	startTime := time.Now()
	defer func(timer time.Time) {
		durTiem := time.Since(startTime)
		yeego.Print(fmt.Sprintf("开始更新币的名称和Symbol对应的map:%.2f秒", durTiem.Seconds()))
	}(startTime)
	var updateFailErr = errors.New("update fail")
	var allCoin entity.CoinInfo
	// 获取所有的币种
	allCoinData, err := yeeHttp.Get(entity.AllCoinApi).Exec().ToBytes()
	if err != nil {
		allCoinData, err = yeeHttp.Get(entity.AllCoinApi).Exec().ToBytes()
		if err != nil {
			log.Println("网络请求:", err)
			return updateFailErr
		}
	}
	err = json.Unmarshal(allCoinData, &allCoin)
	if err != nil {
		log.Println("json解析:", err)
		return updateFailErr
	}
	for _, coin := range allCoin.Data.Coin {
		if 11 > len(allCoin.Data.Fields) || len(coin) != len(allCoin.Data.Fields) {
			return updateFailErr
		}
		entity.CoinMap[coin[1]] = coin[0]
	}
	return nil
}
