/**
 * Created by vannnnish on 27/04/2018.
 * Copyright © 2018年 yeeyuntech. All rights reserved.
 */

package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/yeeyuntech/yeego/yeeStrconv"
	"fmt"
	"strings"
	"coin_query_bot/entity"
	"gitlab.yeeyuntech.com/yee/easyweb"
	"github.com/yeeyuntech/yeego"
	"time"
	"github.com/astaxie/beego/toolbox"
	"coin_query_bot/task"
	"coin_query_bot/module/notice"
)

func main() {
	globalConfig()
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
			message, ok := notice.GetFormatData(strings.ToUpper(update.ChannelPost.Text))
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
		if update.Message == nil {
			continue
		}
		// 获取管理员列表
		memebers, err := bot.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: update.Message.Chat.ID})
		if err != nil {
			fmt.Println("获取管理员失败:", err)
			continue
		}

		var isDelete = true
		for _, memeber := range memebers {
			if update.Message.From.ID == memeber.User.ID {
				isDelete = false
			}
		}
		// 删除加群信息
		if update.Message.NewChatMembers != nil {
			message := update.Message
			fmt.Println("开始删除这条信息")
			_, err := bot.DeleteMessage(tgbotapi.DeleteMessageConfig{message.Chat.ChatConfig().ChatID, message.MessageID})
			if err != nil {
				fmt.Println("这个错误:", err)
			}
		}
		// 删除图片
		if update.Message.Photo != nil {
			message := update.Message
			if isDelete {
				_, err := bot.DeleteMessage(tgbotapi.DeleteMessageConfig{message.Chat.ChatConfig().ChatID, message.MessageID})
				if err != nil {
					fmt.Println("这个错误:", err)
				}
				continue
			}
		}
		if strings.Contains(update.Message.Text, "http://") || strings.Contains(update.Message.Text, "https://") || strings.Contains(update.Message.Text, ".com") ||
			strings.Contains(update.Message.Text, ".net") ||
			strings.Contains(update.Message.Text, ".io") ||
			strings.Contains(update.Message.Text, ".net") ||
			strings.Contains(update.Message.Text, ".org") ||
			strings.Contains(update.Message.Text, ".gov") ||
			strings.Contains(update.Message.Text, ".info") ||
			strings.Contains(update.Message.Text, ".pro") ||
			strings.Contains(update.Message.Text, "好消息，") ||
			strings.Contains(update.Message.Text, "代币制作") ||
			strings.Contains(update.Message.Text, "邀请码") ||
			strings.Contains(update.Message.Text, "私聊我") ||
			strings.Contains(update.Message.Text, "电报群") ||
			strings.Contains(update.Message.Text, "加我") ||
			strings.Contains(update.Message.Text, "电报拉人") ||
			strings.Contains(update.Message.Text, "举报") ||
			strings.Contains(update.Message.Text, "spam") ||
			strings.Contains(update.Message.Text, "又是什么群") ||
			strings.Contains(update.Message.Text, "怎么进来的") ||
			strings.Contains(update.Message.Text, "垃圾群") ||
			strings.Contains(update.Message.Text, "需要帮助科学上网") ||
			strings.Contains(update.Message.Text, "微信") ||
			strings.Contains(update.Message.Text, "糖果空投") ||
			strings.Contains(update.Message.Text, "Imtoken钱包") ||
			strings.Contains(update.Message.Text, "联系我") ||
			strings.Contains(update.Message.Text, "im钱包") ||
			strings.Contains(update.Message.Text, "数量有限先到先得") ||
			strings.Contains(update.Message.Text, "电报群拉人") {
			message := update.Message

			if isDelete {
				_, err := bot.DeleteMessage(tgbotapi.DeleteMessageConfig{message.Chat.ChatConfig().ChatID, message.MessageID})
				if err != nil {
					fmt.Println("这个错误:", err)
				}
			}
		}
		if update.Message.Chat.UserName == "bullseye_official" {

			message, ok := notice.GetGlobalDataEnglish(strings.ToUpper(update.Message.Text))
			if !ok {
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			continue

		}
		message, ok := notice.GetGlobalData(strings.ToUpper(update.Message.Text))
		if !ok {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)

	}
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
