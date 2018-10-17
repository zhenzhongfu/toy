package main

import (
	"toy/hello" // import hello包
	//"toy_util"

	"github.com/gin-gonic/gin"
)

func main() {
	//result := toy_util.Add(1, 2)
	//hello.Print(result)
	hello.Print(3)

	gin.New()
}
