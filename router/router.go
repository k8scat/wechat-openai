package router

import (
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"go.uber.org/zap"

	"github.com/k8scat/wechat-openai/log"
	"github.com/k8scat/wechat-openai/wechat"
)

func Run() error {
	r := gin.Default()
	r.Any("/callback", func(c *gin.Context) {
		oa := wechat.GetOfficialAccount()
		if err := wechat.HandleMessage(oa, c.Request, c.Writer); err != nil {
			log.Error("wechat handle message failed", zap.Error(err), zap.Stack("stack"))
		}
	})

	err := r.Run()
	return errors.Trace(err)
}
