package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"go.uber.org/zap"

	"github.com/k8scat/wechat-openai/db"
	"github.com/k8scat/wechat-openai/log"
	"github.com/k8scat/wechat-openai/wechat"
)

func Run(port int) error {
	r := gin.Default()
	r.Use(ginzap.Ginzap(log.GetLogger(), time.RFC3339, true))

	r.Any("/callback", func(c *gin.Context) {
		oa := wechat.GetOfficialAccount()
		if err := wechat.HandleMessage(oa, c.Request, c.Writer); err != nil {
			log.Error("wechat handle message failed", zap.Error(err), zap.Stack("stack"))
		}
	})

	r.GET("/message/:msgID", func(c *gin.Context) {
		msgID := c.Param("msgID")
		conn := db.GetRedisClient().Conn()
		defer conn.Close()

		ctx := context.Background()
		if conn.Exists(ctx, msgID).Val() == 0 {
			c.String(http.StatusNotFound, "message not found")
			return
		}

		resp := conn.Get(ctx, msgID).Val()
		if resp == "" {
			c.String(http.StatusOK, "waiting for response")
			return
		}
		c.String(http.StatusOK, resp)
	})

	err := r.Run(fmt.Sprintf(":%d", port))
	return errors.Trace(err)
}
