package router

import (
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
		cache := db.GetCache()
		if cache.Exists(msgID) {
			c.String(http.StatusNotFound, "message not found")
			return
		}

		resp, err := cache.Get(msgID)
		if err != nil {
			c.String(http.StatusInternalServerError, "get message failed")
			return
		}
		s, _ := resp.(string)
		m := make(map[string]string)
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			c.String(http.StatusInternalServerError, "invalid message")
			return
		}
		if _, ok := m["response"]; !ok {
			c.String(http.StatusNotFound, "waiting for response")
			return
		}

		content := fmt.Sprintf("message:\n %s\n\n\nresponse:\n %s", m["message"], m["response"])
		c.String(http.StatusOK, content)
	})

	err := r.Run(fmt.Sprintf(":%d", port))
	return errors.Trace(err)
}
