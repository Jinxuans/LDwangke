package bootstrap

import (
	"context"
	"log"
	"time"

	chatmodule "go-api/internal/modules/chat"
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
				log.Printf("[ChatCleanup] 归档失败: %v", err)
			} else if archived > 0 {
				log.Printf("[ChatCleanup] 归档了 %d 条过期消息", archived)
			}

			trimmed, err := trimSessionMessagesFn()
			if err != nil {
				log.Printf("[ChatCleanup] 截断失败: %v", err)
			} else if trimmed > 0 {
				log.Printf("[ChatCleanup] 截断了 %d 条超限消息", trimmed)
			}
		}
	}()
}
