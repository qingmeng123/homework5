/*******
* @Author:qingmeng
* @Description:
* @File:main
* @Date2021/11/23
 */

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
)

type UserInfo struct {
	Name 		string `json:"name"`
	Password 	string `json:"password"`
}

func (i *UserInfo) ifPasswordCorrect(password string) bool {
	return password==i.Password
}

var (
	r 				=gin.Default()
	userMap			=make(map[string]UserInfo)
	cookie			UserInfo
	dataFile		*os.File
	filePath		="lv2/user.data"
	sensitiveWords = make([]interface{}, 0)
)

func main() {
	Init()
	ReloadData()		//重载数据
	login()
	register()
	hello()
	r.Run(":80")
}

func hello(){
	auth:=func(c *gin.Context) {
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
	r.GET("/hello",auth, func(c *gin.Context) {
		cookie,value:=c.Get("cookie")
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
}


func login(){
	r.POST("/login", func(c *gin.Context) {
		name:=c.PostForm("name")
		userInfo,ok:=userMap[name]
		if !ok{
			c.JSON(http.StatusForbidden,gin.H{
				"message":"账号不存在",
			})
			return
		}
		password:=c.PostForm("password")
		if !userInfo.ifPasswordCorrect(password){
			c.JSON(http.StatusForbidden,gin.H{
				"message":"密码错误",
			})
			return
		}
		c.SetCookie("my_cookie",name,3600,"/","",false,true)
		c.JSON(200,gin.H{
			"message":"login successfully",
		})
	})
}

func register(){
	r.POST("/register", func(c *gin.Context) {
		name:=c.PostForm("name")
		_, ok := userMap[name]
		if ok{
			c.JSON(http.StatusForbidden,gin.H{
				"message":"账号已存在",
			})
			return
		}
		if name==""||checkIfSensitive(name){
			c.JSON(http.StatusForbidden,gin.H{
				"message":"用户名不合法",
			})
			return
		}
		password:=c.PostForm("password")
		if !checkPasswordLegal(password){
			c.JSON(http.StatusForbidden,gin.H{
				"message":"密码不合法",
			})
			return
		}
		cookie =UserInfo{
			Name: name,
			Password: password,
		}
		userMap[name]=cookie
		OverWriteData()
		c.JSON(200,gin.H{
			"message":"register successfully",
		})
	})

}

//来吧，字典树！
type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		root: newTrieNode(),
	}
}
//字典树节点
type TrieNode struct {
	children map[interface{}]*TrieNode
	isEnd	bool
}

func newTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[interface{}]*TrieNode),
		isEnd:    false,
	}
}

func (trie *Trie) Insert(word []interface{}) {
	if len(word)==0{
		return
	}
	node:=trie.root
	for i := 0; i < len(word); i++ {
		_,ok:=node.children[word[i]]
		if !ok{
			node.children[word[i]]=newTrieNode()
		}
		node=node.children[word[i]]
	}
	node.isEnd=true
}

func (trie *Trie) StartsWith(prefix []interface{}) bool {
	node:=trie.root
	for i := 0; i < len(prefix); i++ {
		_,ok:=node.children[prefix[i]]
		if !ok{
			return false
		}
		node=node.children[prefix[i]]
	}
	return true
}

func checkIfSensitive(s string) bool {
	trie:=NewTrie()
	trie.Insert(sensitiveWords)
	bt:=[]interface{}{s}
	return trie.StartsWith(bt)
}

func checkPasswordLegal(password string) bool {
	return len(password) > 6
}

func OverWriteData() {
	dataFile.Truncate(0)
	dataFile.Seek(0, io.SeekStart)
	for _, info := range userMap {
		bytes, _ := json.Marshal(info)
		bytes = append(bytes, '\n')
		_, err := dataFile.Write(bytes)
		if err != nil {
			log.Println(err)
		}
	}
}

func Init() {
	df, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.ModeAppend|os.ModePerm)
	if err != nil {
		panic(err)
	}
	dataFile = df
	sensitiveWords = append(sensitiveWords, "你妈", "傻逼")
}

func ReloadData(){
	all,err:=io.ReadAll(dataFile)
	if len(all)==0{
		return
	}
	if err!=nil{
		panic(err)
	}
	var u UserInfo
	reader:=bufio.NewReader(bytes.NewReader(all))
	for{
		readString,err:=reader.ReadString('\n')
		if err!=nil{
			if errors.Is(err,io.EOF){
				return
			}
			panic(err)
		}
		err=json.Unmarshal([]byte(readString),&u)
		if err!=nil{
			panic(err)
		}
		userMap[u.Name]=u
	}
}