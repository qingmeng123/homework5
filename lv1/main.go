/*******
* @Author:qingmeng
* @Description:
* @File:main
* @Date2021/11/19
 */

package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r:=gin.Default()
	auth:= func(c *gin.Context) {
		value,err:=c.Cookie("my_cookie")
		if err !=nil{
			c.JSON(http.StatusForbidden,gin.H{
				"message":"认证失败，没有cookie",
			})
			c.Abort()
		}else{
			c.Set("cookie",value)
		}
	}
	r.POST("/login", func(c *gin.Context) {
		username:=c.PostForm("username")
		password:=c.PostForm("password")
		if username=="qingmeng"&&password=="123"{
			c.SetCookie("my_cookie",username,7200,"/","",false,true)
			c.JSON(200,gin.H{
				"msg":"login successfully",
			})
		}else{
			c.JSON(http.StatusForbidden,gin.H{
				"message":"认证失败,账号或密码错误",
			})
		}
	})

	r.GET("/hello",auth, func(c *gin.Context) {
		cookie,value:=c.Get("cookie")
		//当然，这里肯定有cookie
		if !value{
			c.JSON(403,gin.H{
				"message":"登陆失败,没有cookie",
			})
			c.Abort()
		}else {
			str:=cookie.(string)
			c.String(200,"hello world\t "+str)

		}
	})
	r.Run()
}