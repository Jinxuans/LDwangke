package bootstrap

import (
	"context"
	"time"

	chatmodule "go-api/internal/modules/chat"
	obslogger "go-api/internal/observability/logger"
)

var (
	archiveOldMessagesFn  = chatmodule.Chat().ArchiveOldMessages
	trimSessionMessagesFn = chatmodule.Chat().TrimSessionMessages
)

// StartChatCleanup 启动聊天消息归档与裁剪任务。
func StartChatCleanup(ctx context.Context) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, now.Location())
			if !sleepContextFn(ctx, time.Until(next)) {
				return
			}

			archived, err := archiveOldMessagesFn()
			if err != nil {
				obslogger.L().Warn("ChatCleanup 归档失败", "error", err)
			} else if archived > 0 {
				obslogger.L().Info("ChatCleanup 已归档过期消息", "archived", archived)
			}

			trimmed, err := trimSessionMessagesFn()
			if err != nil {
				obslogger.L().Warn("ChatCleanup 截断失败", "error", err)
			} else if trimmed > 0 {
				obslogger.L().Info("ChatCleanup 已截断超限消息", "trimmed", trimmed)
			}
		}
	}()
}
