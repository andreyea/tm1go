package main

import (
	"fmt"

	"github.com/andreyea/tm1go/tm1"
	"github.com/tealeg/xlsx"
)

func main() {
	cubes := []tm1.Cube{}
	sdata := tm1.NewSession("https://localhost:8010/api/v1", "admin", "apple", "")
	sdata.Login()
	xlFile, err := xlsx.OpenFile("tm1Cubes.xlsx")
	if err != nil {
		fmt.Println("error opening file")
	}

	sheet := xlFile.Sheets[0]
	for i, row := range sheet.Rows {
		//skip headers
		if i == 0 {
			continue
		}
		c := tm1.Cube{}
		for j, cell := range row.Cells {
			text := cell.String()
			//if no more dimensions defined, stop
			if text == "" {
				break
			}
			//first cell in a row is a cube

			if j == 0 {
				//fmt.Printf("Cube: %s\n", text)
				c, _ = tm1.CubeCreate(text)
				cubes = append(cubes, c)
				//following cells are dimensions
			} else {
				//fmt.Printf("Dimension: %s\n", text)
				d, _ := tm1.DimensionCreate(text)
				cubes[i-1].Dimensions = append(cubes[i-1].Dimensions, d)

			}

		}
	}

	for _, c := range cubes {
		err = sdata.CubeCreate(c)
		if err != nil {
			fmt.Printf("Error creating %s cube", c.Name)
		}
	}

	sdata.Logout()
}
