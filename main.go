package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
)

func main() {
	// 读取配置
	config := readConfig()
	port, ok := config["port"]
	if !ok {
		panic("Failed to read port from config")
	}

	fmt.Println("Starting the FRPS Auth Service...")
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.POST("/login", login)
	r.POST("/new_proxy", newProxy)
	r.Run(fmt.Sprintf(":%v", port)) // Use the port from the config
}

func login(c *gin.Context) {
	// 读取配置文件
	config := readConfig()
	users := config["users"].([]map[string]any)
	// 获取请求的user参数
	var req map[string]any
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{
			"reject":        true,
			"reject_reason": fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}
	ruser := req["content"].(map[string]any)["user"].(string)
	// 判断user是否在用户列表中
	result := false
	for _, user := range users {
		cuser, ok := user["user"].(string)
		if ok && cuser == ruser {
			result = true
		}
	}
	if !result {
		c.JSON(200, gin.H{
			"reject":        true,
			"reject_reason": "用户不存在",
		})
		return
	}
	c.JSON(200, gin.H{
		"reject":   false,
		"unchange": true,
	})
}

func newProxy(c *gin.Context) {
	// 读取配置文件
	config := readConfig()
	users := config["users"].([]map[string]any)
	// 获取请求的user参数
	var req map[string]any
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{
			"reject":        true,
			"reject_reason": fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}
	ruser := req["content"].(map[string]any)["user"].(map[string]any)["user"].(string)
	rport := int64(req["content"].(map[string]any)["remote_port"].(float64))
	cuser := make(map[string]any)
	for _, user := range users {
		userStr, ok := user["user"].(string)
		if ok && userStr == ruser {
			cuser = user
			break
		}
	}
	_, ok := cuser["user"]
	if !ok {
		c.JSON(200, gin.H{
			"reject":        true,
			"reject_reason": "用户不存在",
		})
		return
	}
	ports, ok := cuser["ports"]
	if ok {
		portsArr := ports.([]any)
		// 检查端口是否包含在列表中
		bh := false
		for _, port := range portsArr {
			if port.(int64) == rport {
				bh = true
				break
			}
		}
		if !bh {
			c.JSON(200, gin.H{
				"reject":        true,
				"reject_reason": fmt.Sprintf("端口 %d 不在用户 %s 的授权端口列表中", rport, ruser),
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"reject":   false,
		"unchange": true,
	})
}

func readConfig() map[string]any {
	var config map[string]any
	if _, err := toml.DecodeFile("frps_auth.toml", &config); err != nil {
		panic(err)
	}
	return config
}
