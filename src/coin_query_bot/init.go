/**
 * Created by vannnnish on 28/04/2018.
 * Copyright © 2018年 yeeyuntech. All rights reserved.
 */

package main

import (
	"coin_query_bot/task"
)

func Init() {
	// 开始初始化
	task.UpdateFinaceRate()
	task.UpdateAllCoins()
	//	task.UpdatePrimaryExchangePair()
	task.UpdateCoin()

}
