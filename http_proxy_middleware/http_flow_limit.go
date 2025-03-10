package http_proxy_middleware

import (
	"bytes"
	"fmt"
	"github.com/didi/gatekeeper/dashboard_middleware"
	"github.com/didi/gatekeeper/handler"
	"github.com/didi/gatekeeper/model"
	"github.com/didi/gatekeeper/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPFlowLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceDetail, err := model.GetServiceDetailFromGinContext(c)
		if err != nil {
			public.ResponseError(c, 2001, err)
			c.Abort()
			return
		}
		serviceFlowNum := serviceDetail.PluginConf.GetPath("http_flow_limit", "service_flow_limit_num").MustInt()
		serviceFlowType := serviceDetail.PluginConf.GetPath("http_flow_limit", "service_flow_limit_type").MustInt()
		if serviceFlowNum > 0 {
			limiterBuffer := bytes.NewBufferString(public.FlowServicePrefix)
			limiterBuffer.WriteString(serviceDetail.Info.ServiceName)

			serviceLimiter, err := handler.FlowLimiterHandler.GetLimiter(limiterBuffer.String(), float64(serviceFlowNum), serviceFlowType, true)
			if err != nil {
				public.ResponseError(c, 5001, err)
				c.Abort()
				return
			}
			if !serviceLimiter.Allow() {
				public.ResponseError(c, 5002, errors.New(fmt.Sprintf("service flow limit %v", serviceFlowNum)))
				c.Abort()
				return
			}
		}

		clientIpFlowNum := serviceDetail.PluginConf.GetPath("http_flow_limit", "clientip_flow_limit_num").MustInt()
		clientIpFlowType := serviceDetail.PluginConf.GetPath("http_flow_limit", "clientip_flow_limit_type").MustInt()
		if clientIpFlowNum > 0 {
			cLimiterBuffer := bytes.NewBufferString(public.FlowServicePrefix)
			cLimiterBuffer.WriteString(serviceDetail.Info.ServiceName)
			cLimiterBuffer.WriteString("_")
			cLimiterBuffer.WriteString(c.ClientIP())
			clientLimiter, err := handler.FlowLimiterHandler.GetLimiter(cLimiterBuffer.String(), float64(clientIpFlowNum), clientIpFlowType, true)
			if err != nil {
				public.ResponseError(c, 5003, err)
				c.Abort()
				return
			}

			if clientLimiter == nil {
				dashboard_middleware.ResponseError(c, 5002, errors.New(fmt.Sprintf("clientLimiter is nil")))
				c.Abort()
				return
			}

			if !clientLimiter.Allow() {
				public.ResponseError(c, 5002, errors.New(fmt.Sprintf("%v flow limit %v", c.ClientIP(), clientIpFlowNum)))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
