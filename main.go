package main

import (
	"fmt"
	"time"
	"tm1go/pkg/tm1"
)

func main() {

	//credentials
	tm1Instance := tm1.Tm1Instance{false, "localhost", "55000", "admin", "apple"}

	_ = tm1.Login(tm1Instance)

	//tm1.CellGetS(tm1Instance, "cube1", "a", "1", "str")
	//tm1.CellGetN(tm1Instance, "cube1", "a", "1", "value")
	fmt.Println("running:")
	var params []tm1.ProcessParameter
	go tm1.ExecuteProcess(tm1Instance, "DataLoad", params)

	time.Sleep(15 * time.Second)
	tm1.Logout(tm1Instance)

}
