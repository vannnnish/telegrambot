/**
 * Created by angelina on 2017/9/2.
 * Copyright © 2017年 yeeyuntech. All rights reserved.
 */

package easyweb

import "gitlab.yeeyuntech.com/yee/easyweb/session"

var (
	globalSessions         *session.Manager
	enableSession          bool                   = false
	defaultSessionProvider string                 = "memory"
	sessionManagerConfig   *session.ManagerConfig = &session.ManagerConfig{
		CookieName:              "easyWebSessionID",  //
		GcLifeTime:              3600,                // 3600s 1h
		MaxLifeTime:             3600,                // 3600s 1h
		CookieLifeTime:          0,                   // 浏览器生命周期
		ProviderConfig:          "",                  //
		SessionIDLength:         32,                  //
		SessionNameInHTTPHeader: "Easyweb-Sessionid", //
	}
)

// 是否开启session
func enableSessionOn(b bool) {
	enableSession = b
}

// 启动session
func StartSession() error {
	if enableSession {
		var err error
		if globalSessions, err = session.NewManager(defaultSessionProvider, sessionManagerConfig); err != nil {
			return err
		}
		go globalSessions.GC()
	}
	return nil
}
