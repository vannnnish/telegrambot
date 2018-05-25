/**
 * Created by vannnnish on 28/04/2018.
 * Copyright © 2018年 yeeyuntech. All rights reserved.
 */

package entity

import "sync"

const MessageHead = `%s(%s):
------------------
`

var NewFormatFront = `💲%s(%s)实时行情💲
───────────
💡排名：%s
💡价格：  ￥%.4f
💡1小时涨幅： %.2f%%
💡24小时涨幅: %.2f%%
`
var NewFormatMiddle = `💡今日净流入：￥%.2f万`
var NewFormatTail = `
───────────
获取资金走势图: http://dl.niuyan.com/?utm_source=tg

牛眼行情【%s】
`

var NewFormatFrontEnglish = `💲%s(%s)💲
───────────
💡Rank：%s
💡Price：$%.4f
💡1H：%.2f%%
💡24H：%.2f%%
`

var NewFormatMiddleEnglish = `💡Net Inflow：$ %.2f`

var NewFormatTailEnglish = `
───────────
From Bullseye.io
App download: http://dl.niuyan.com/?utm_source=en
Bullseye 【%s】
`

var MessageHeadGlobal = `%s(%s):
全网净流入: ￥%.2f
------------------
`

var FormatMessage = `%s（%s/%s）
最新价格：$%.4f
今日涨幅：%.2f%%
今日净流入：$%.2f
------------------
`
var FormatMessageCNY = `%s（%s/%s）
最新价格：￥%.4f
今日涨幅：%.2f%%
今日净流入：￥%.2f
------------------
`

var GlobalMessage = `
币种:    %s
全网价格：$%.4f
今日涨幅：%.2f%%
------------------
`
var GlobalMessageCNY = `
币种:    %s
全网价格：￥%.4f
今日涨幅：%.2f%%
`
var GlobalPureCNY = `全网净流入: ￥%.2f`

var GlobalMessgeCNYTail = `
------------------
`

var FormatTail = `更新时间:%s
数据来源:牛眼行情
`

var (
	// 获取全部币种信息
	AllCoinApi = "http://market.niuyan.com/api/v3/app/coins/marketcap?pagesize=9999&offset=0&order_type=in_order"
	// 单个交易所单个币信息
	PairApi = "https://market.niuyan.com/api/v3/app/exchange/coin?coin_id=%s&exchange_id=%s"
	// 资金进入
	CashInApi = "http://market.niuyan.com/api/v3/app/finance/ticker/today?exchange_id=%s&base_symbol=%s&quote_symbol=%s"
	// 获取交易所交易对详情
	PairInExchangeApi = "http://market.niuyan.com/api/v3/app/ticker?exchange_id=%s&base_symbol=%s&quote_symbol=%s"
	// 获取单个交易所全部的交易对
	AllExchangePairApi = "http://market.niuyan.com/api/v3/app/exchange/tickers?exchange_id=%s"
	// 获取币种详情
	SingleCoinInfoApi = "http://market.niuyan.com/api/v3/app/coin?coin_id=%s"
	// 获取币种全网净流入
	GlobalCoinFinance = "https://market.niuyan.com/api/v3/app/finance/coin/today?coin_id=%s"
	// 汇率接口
	FinanceRateApi = "https://market.niuyan.com/api/v3/common/financerate"
)
var (
	// 全部的币种
	CoinIdMap = sync.Map{}

	// 行情信息
	MarketIno = sync.Map{}

	// 交易所的信息
	PrimaryExchangePairInfo = sync.Map{}

	// 火币根据BaseCoin的索引
	HuoBiExchangePair = sync.Map{}
	// OKEx
	OKExExchangePair = sync.Map{}
	// 币安
	BinanceExchangePair = sync.Map{}
	// ZB
	ZbExchangePair = sync.Map{}
	// GateIo
	GateIoExchangePair = sync.Map{}
)

var ExchangeMaps = []*sync.Map{&HuoBiExchangePair, &OKExExchangePair, &BinanceExchangePair, &ZbExchangePair, &GateIoExchangePair,}

// 待查询的交易所列表
var PrimaryExhcnageNameMap = map[string]string{
	"huobi.pro": "火币",
	"zb":        "ZB",
	"binance":   "币安",
	"okex":      "OKEx",
	"gateio":    "Gate.io",
}
// map[symbol]coinId
var CoinMap = make(map[string]string)

// 汇率
var FinalRateSyncMap = sync.Map{}
