package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"go.uber.org/zap"

	"github.com/k8scat/wechat-openai/db"
	"github.com/k8scat/wechat-openai/log"
	"github.com/k8scat/wechat-openai/openai"
	"github.com/k8scat/wechat-openai/wechat"
)

func Run(port int) error {
	r := gin.Default()
	r.Use(ginzap.Ginzap(log.GetLogger(), time.RFC3339, false))

	r.Any("/callback", func(c *gin.Context) {
		oa := wechat.GetOfficialAccount()
		if err := wechat.HandleMessage(oa, c.Request, c.Writer); err != nil {
			log.Error("wechat handle message failed", zap.Error(err), zap.Stack("stack"))
		}
	})

	r.GET("/user/:userID/message/:msgID", func(c *gin.Context) {
		userID := c.Param("userID")
		msgID := c.Param("msgID")
		msgKey := fmt.Sprintf("message:%s", msgID)

		conn := db.GetRedisClient().Conn()
		defer conn.Close()
		ctx := context.Background()

		if !conn.HExists(ctx, userID, msgKey).Val() {
			c.String(http.StatusNotFound, "chat not found")
			return
		}

		cmd := conn.HGet(ctx, userID, msgKey)
		if cmd.Err() != nil {
			c.String(http.StatusInternalServerError, "get chat failed")
			return
		}
		s := cmd.Val()
		var chat openai.Chat
		if err := json.Unmarshal([]byte(s), &chat); err != nil {
			c.String(http.StatusInternalServerError, "invalid chat")
			return
		}
		if chat.Answer == "" {
			c.String(http.StatusNotFound, "waiting for answer")
			return
		}

		content := fmt.Sprintf("question:\n %s\n\n\nanswer:\n %s", chat.Question, chat.Answer)
		c.String(http.StatusOK, content)
	})

	err := r.Run(fmt.Sprintf(":%d", port))
	return errors.Trace(err)
}
