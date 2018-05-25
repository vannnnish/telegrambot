/**
 * Created by vannnnish on 28/04/2018.
 * Copyright © 2018年 yeeyuntech. All rights reserved.
 */

package entity

import "github.com/spf13/cast"

// 行情机器人返回的数据
type ExchangePairInfo struct {
	ExchangeName string
	Symbol       string
	QuoteSymbol  string
	Price        float64
	Change24h    float64
	CashIn       float64
	GlobalCashIn float64
}

// 币种在相应的交易所的涨跌幅与价格
type CoinPriceInfo struct {
	Data struct {
		Data struct {
			Price_usd            float64
			Percent_change_today float64
		}
	}
}

// 资金流入返回的数据
type CashIn struct {
	Data struct {
		Data struct {
			Buy_vol_usd  float64
			Sell_vol_usd float64
		}
	}
}

// 全部币种信息
type CoinInfo struct {
	Data struct {
		Coin   [][]string
		Fields []string
	}
}

// 交易所交易对信息
type AllExchangeInfo struct {
	Data struct {
		Data [][]interface{}
	}
}

// 单个币种
type SingleCoin struct {
	Data struct {
		Coin struct {
			Price_usd          string
			Percent_change_1h  string
			Percent_change_24h string
			Rank               string
		}
	}
}

// 汇率
type FinanceRate struct {
	Data struct {
		Legal_rate struct {
			CNY string
			JPY string
		}
	}
}

// 交易所交易对数据结构
type ExchangePair struct {
	ExchangeId   string
	Symbol       string
	QuoteSymbol  string
	PriceUsd     float64
	ChangeToday  float64
	VolumeUse24h float64
}

// 按照流通市值排序的
type SortByVolume24 []ExchangePair

func (p SortByVolume24) Len() int {
	return len(p)
}

func (p SortByVolume24) Less(i, j int) bool {
	t1 := cast.ToInt(p[i].VolumeUse24h)
	t2 := cast.ToInt(p[j].VolumeUse24h)
	return t1 > t2
}

func (p SortByVolume24) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
