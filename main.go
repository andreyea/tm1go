package main

import (
	"fmt"
	"time"
	"tm1go/pkg/tm1"
)

func main() {

	//credentials
	user := "admin"
	password := "apple"
	server := "localhost"
	httpPort := "55000"
	ssl := false
	tm1Instance := tm1.Tm1Instance{false, "localhost", "55000", "admin", "apple"}

	_ = tm1.Login(ssl, httpPort, server, user, password)

	//tm1.CellGetS(tm1Instance, "cube1", "a", "1", "str")
	//tm1.CellGetN(tm1Instance, "cube1", "a", "1", "value")
	fmt.Println("running:")
	go tm1.ExecuteProcess(tm1Instance, "DataLoad", "a")
	go tm1.ExecuteProcess(tm1Instance, "DataLoad", "b")
	go tm1.ExecuteProcess(tm1Instance, "DataLoad", "c")

	time.Sleep(15 * time.Second)
	tm1.Logout(ssl, httpPort, server)

}
