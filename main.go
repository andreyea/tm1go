package main

import (
	"fmt"
	"time"

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
	//initialize logs
	transactions, err := sdata.GetTransactionLogs("")
	if err != nil {
		fmt.Println("error during get transcations")
		fmt.Println(err)
	}

	for {
		transactions, err = sdata.GetTransactionLogs(transactions.OdataDeltaLink)
		if err != nil {
			fmt.Println("error")
			break
		}

		//wait 5 sec
		time.Sleep(5 * time.Second)
	}

	/*
		cs, err := sdata.ExecuteView("x Cube", "tst2")

		if err != nil {
			fmt.Println("error during view get")
			fmt.Println(err)
		}

		err = sdata.DestroyCellset(cs.ID)
		if err != nil {
			fmt.Println("error during cs delete")
			fmt.Println(err)
		}

		matrix := cs.CreateMatrix()

		for _, v1 := range matrix {
			for _, v2 := range v1 {
				//fmt.Printf("%v %v \t", v2.NValue, v2.SValue)
				fmt.Printf("%v\t", returnMatrixCellValue(v2))
			}
			fmt.Printf("\n")
		}

		//fmt.Println(matrix)

			err = sdata.CellPutN(123123, "x Cube", "v1", "a", "value", "test")
			if err != nil {
				fmt.Println("error during cellputN")
				fmt.Println(err)
			}

				c123, _ := tm1.CubeCreate("c123")
				d1, _ := tm1.DimensionCreate("region")
				d2, _ := tm1.DimensionCreate("c123 Measures")

				c123.Dimensions = append(c123.Dimensions, d1, d2)
				//fmt.Println(c123)
				check, _ := sdata.CubeExists(c123.Name)
				if check {
					sdata.CubeDestroy(c123.Name)
				}
				err = sdata.CubeCreate(c123)
				if err != nil {
					fmt.Println("error during cube create")
					fmt.Println(err)
				} else {
					fmt.Println("Cube Created")
				}


					tis, err := sdata.GetProcesses()
					if err != nil {
						fmt.Println("error during proccesses get")
						fmt.Println(err)
						return
					}

					for _, v := range tis {
						//fmt.Println(v.Name)
						if v.Name == "!temp" {
							v.Parameters[0].Value = 10
							//fmt.Println(v)
							zzz, err := sdata.ExecuteProcess(v)
							if err != nil {
								fmt.Println("error during proccesses run")
								fmt.Println(err)
							}
							fmt.Println("Execute Process Result:")
							fmt.Println(zzz)
						}
					}

					//h1 := tm1.Hierarchy{}
					//d1 := tm1.Dimension{}
					//d1.Name = "tm1goDimensionNew_"
					//h1.Name = "tm1goDimensionNew_"
					//h1.Elements = []tm1.Element{}
					//d1.Hierarchies = []tm1.Hierarchy{h1}
					//fmt.Println(d1)
					//err = sdata.DimensionCreate(d1)
					//if err != nil {
					//	fmt.Println(err)
					//}


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

func returnMatrixCellValue(c tm1.Tm1MatrixCell) interface{} {
	if c.BStr {
		return c.SValue
	} else {
		return c.NValue
	}
}
