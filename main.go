package main

import (
	"flag"
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/jackmanlabs/errors"
	"github.com/unidoc/unidoc/pdf/creator"
	"log"
)

func main() {

	var (
		prefix *string = flag.String("prefix", "", "Prefix that will be appended to the serial number.")
		start  *int    = flag.Int("start", 0, "Where to start the serial number sequence.")
		end    *int    = flag.Int("end", -1, "Where to end the serial number sequence (inclusive).")
	)

	flag.Parse()

	if *prefix == "" {
		log.Print("WARNING: No prefix has been specified. Only the serial number will be used.")
	}

	if *end < 0 {
		flag.Usage()
		log.Fatal("A valid end for the serial number must be specified.")
	}

	if *start > *end {
		flag.Usage()
		log.Fatal("The end of the serial number series must be after the beginning.")
	}

	c := creator.New()
	c.SetPageSize(creator.PageSizeLetter)
	c.SetPageMargins(0, 0, 0, 0)

	for i := *start; i <= *end; i++ {

		// This creates a new page and resets the label position.
		if (i-*start)%30 == 0 {
			c.NewPage()
		}

		// Perform x alignment based on column.
		col := (i - *start) % 3
		x := (c.Width()-creator.PPI/2)*float64(col)/3 + creator.PPI/4

		log.Print("Col: ", col)

		// Perform y alignment based on row.
		row := ((i - *start) / 3) % 10
		y := (c.Height()-creator.PPI)*float64(row)/10 + creator.PPI/2

		log.Print("Row: ", row)

		// Create the barcode.
		serial := fmt.Sprintf("%s%d", prefix, i)

		bc128, err := code128.Encode(serial)
		if err != nil {
			log.Fatal(errors.Stack(err))
		}

		//bcWidth := int(c.Width()/3) - 15
		//bcHeight := int(c.Height()/30) - 15

		// This at least gets it to the right proportion.
		bcWidth := 525
		bcHeight := 100
		//log.Print("BC Width: ", bcWidth)
		//log.Print("BC Height: ", bcHeight)

		bc, err := barcode.Scale(bc128, bcWidth, bcHeight)
		if err != nil {
			log.Fatal(errors.Stack(err))
		}

		img, err := creator.NewImageFromGoImage(bc)
		if err != nil {
			log.Fatal(errors.Stack(err))
		}

		colWidth := c.Width() / 3
		img.ScaleToWidth(colWidth)
		c.MoveTo(x, y)
		//img.SetPos(x, y)
		err = c.Draw(img)
		if err != nil {
			log.Fatal(errors.Stack(err))
		}
	}

	err := c.WriteToFile("out.pdf")
	if err != nil {
		log.Fatal(errors.Stack(err))
	}

}
