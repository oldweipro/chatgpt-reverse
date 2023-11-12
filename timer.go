package main

import (
	"fmt"
	"strings"
	"time"
)

func TimerFunc() {
	_, err := Timer.AddTaskByFunc("SyncAccessToken", "0 23 * * *", func() {
		var accounts []MailAccount
		err := DB.Where("openai_status!=0").Find(&accounts).Error
		if err != nil {
			fmt.Println("查询openai_status!=0的账号失败: SyncAccessToken")
			return
		}
		updateToken(accounts)
		if err != nil {
			fmt.Println("定时任务执行失败: SyncAccessToken")
		}
	})
	if err != nil {
		fmt.Println("添加定时任务【SyncAccessToken】 error:", err)
	}
}

func updateToken(accounts []MailAccount) {
	count := 0
	for _, account := range accounts {
		if count > 5 {
			count = 0
			time.Sleep(120 * time.Second)
		}
		println("Updating access token for " + account.Username)
		proxyUrl := ProxyUrl
		authenticator := NewAuthenticator(account.Username, account.OpenaiPassword, proxyUrl)
		err := authenticator.Begin()
		if err != nil {
			println("Location: " + err.Location)
			println("Status code: " + fmt.Sprint(err.StatusCode))
			println("Details: " + err.Details)
			if strings.HasPrefix(err.Details, "Get \"https://chat.openai.com/api/auth/csrf\":") {
				// 出现这个错误，说明是代理ip被暂时封禁了，需要等待一些时间
				count++
			}
			if strings.HasPrefix(err.Details, "You do not have an account because it has been deleted or deactivated.") {
				println("账号被封")
				// TODO 账号被封禁，需要将账号状态openai_status改为0
				DB.Model(&MailAccount{}).Where("id = ?", account.ID).Update("openai_status", 0)

			} else if strings.HasPrefix(err.Details, "Email or password is not correct.") {
				println("账号密码错误")
				// TODO 账号密码错误，需要将账号状态openai_status改为0
				DB.Model(&MailAccount{}).Where("id = ?", account.ID).Update("openai_status", 0)
			}
			continue
		}
		accessToken := authenticator.GetAccessToken()
		updateColumns := make(map[string]interface{})
		updateColumns["openai_access_token"] = accessToken
		updateColumns["openai_access_token_get_time"] = time.Now()
		updateColumns["openai_status"] = 3
		puid, _ := authenticator.GetPUID()
		if puid == "" {
			println("没有PUID")
			// TODO 这是GPT3.5账号，GPT4过期了,需要将账号状态openai_status改为3
			DB.Model(&MailAccount{}).Where("id = ?", account.ID).Updates(&updateColumns)
			continue
		}
		count++
		updateColumns["openai_status"] = 4
		updateColumns["puid"] = puid
		// TODO 这是GPT4账号，需要将账号状态openai_status改为4
		DB.Model(&MailAccount{}).Where("id = ?", account.ID).Updates(&updateColumns)
		println("Success!")
		time.Sleep(10 * time.Second)
	}
}
