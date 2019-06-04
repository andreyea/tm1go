package main

import (
	"fmt"
	"tm1/tm1/tm1"
)

func main() {
	sdata := tm1.NewSession("https://localhost:8010/api/v1", "usr1", "apple", "")

	err := sdata.Login()
	if err != nil {
		fmt.Println("error during login")
		fmt.Println(err)
		return
	}

	mdx := "select [x version].members on 0, [x el].members on 1 from [x Cube]"

	cellset, err := sdata.ExecuteMdx(mdx)

	fmt.Printf("%v\n", len(cellset.Cells))

	for _, v := range cellset.Cells {
		if v.Value == nil {
			fmt.Println(0)
			continue
		}
		fmt.Println(v.Value)
	}
	if err != nil {
		fmt.Println(err)

	}

	
	err= sdata.Logout()
	if err != nil {
		fmt.Println(err)
		return
	}

}
