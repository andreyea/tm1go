package main

import (
	"fmt"

	"github.com/andreyea/tm1go/tm1"
)

func main() {
	sdata := tm1.NewSession("https://localhost:8010/api/v1", "admin", "apple", "")

	err := sdata.Login()
	if err != nil {
		fmt.Println("error during login")
		fmt.Println(err)
		return
	}

	h1 := tm1.Hierarchy{}
	d1 := tm1.Dimension{}
	d1.Name = "tm1goDimensionNew_"
	h1.Name = "tm1goDimensionNew_"
	h1.Elements = []tm1.Element{}
	d1.Hierarchies = []tm1.Hierarchy{h1}
	fmt.Println(d1)
	err = sdata.DimensionCreate(d1)
	if err != nil {
		fmt.Println(err)
	}

	/*
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
	*/
	err = sdata.Logout()
	if err != nil {
		fmt.Println(err)
		return
	}

}
