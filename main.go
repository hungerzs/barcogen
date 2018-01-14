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

const (
	PPI            float64 = 72 // Default points per inch
	PAGE_WIDTH     float64 = PPI * (8 + 1/2)
	PAGE_HEIGHT    float64 = PPI * (11)
	BARCODE_HEIGHT float64 = PPI * (0 + 1/2)  // Height of the barcode
	BARCODE_WIDTH  float64 = PPI * (2)        // Width of the barcode
	LABEL_HEIGHT   float64 = PPI * (1)        // Height of the label
	LABEL_MARGIN   float64 = PPI * (0 + 1/8)  // Space between the columns of labels
	LABEL_PADDING  float64 = PPI * (0 + 1/8)  // Padding between the content and the label edge
	LABEL_WIDTH    float64 = PPI * (2 + 5/8)  // Width of the label
	OFFSET_X       float64 = PPI * (0 + 3/16) // Left margin of the label page
	OFFSET_Y       float64 = PPI * (0 + 1/2)  // Top margin of the label page
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
	c.SetPageSize(creator.PageSize{PAGE_WIDTH, PAGE_HEIGHT})
	c.SetPageMargins(0, 0, 0, 0)

	for i := *start; i <= *end; i++ {

		// This creates a new page.
		if (i-*start)%30 == 0 {

			c.NewPage()

			// Draw a lines around labels for debugging
			var (
				x1, x2, y1, y2 float64
				line           creator.Drawable
			)

			for col := 0.0; col < 3; col++ {

				// left edge
				x1 = OFFSET_X + col*(LABEL_WIDTH+LABEL_MARGIN)
				x2 = x1
				y1 = 0
				y2 = c.Height()
				line = creator.NewLine(x1, y1, x2, y2)
				c.Draw(line)

				// right edge
				x1 = OFFSET_X + col*(LABEL_WIDTH+LABEL_MARGIN) + LABEL_WIDTH
				x2 = x1
				y1 = 0
				y2 = c.Height()
				line = creator.NewLine(x1, y1, x2, y2)
				c.Draw(line)
			}

			for y := OFFSET_Y; y <= c.Height(); y += LABEL_HEIGHT {
				x1 = 0
				x2 = c.Width()
				y1 = y
				y2 = y
				line = creator.NewLine(x1, y1, x2, y2)
				c.Draw(line)
			}
		}

		row := ((i - *start) / 3) % 10
		col := (i - *start) % 3

		var (
			// These should reflect the top left corner of the physical label.
			x float64 = OFFSET_X + LABEL_WIDTH*float64(col) + LABEL_MARGIN*float64(col)
			y float64 = OFFSET_Y + LABEL_HEIGHT*float64(row)

			// These represent the top left corner of the printable area of the label.
			x_ float64 = x + LABEL_PADDING
			y_ float64 = y + LABEL_PADDING
		)

		// Create the barcode.
		serial := fmt.Sprintf("%s%d", *prefix, i)

		bc128, err := code128.Encode(serial)
		if err != nil {
			log.Fatal(errors.Stack(err))
		}

		// The width/height constants don't really mean anything in this context,
		// but it gives us the right aspect ratio for later.
		bc, err := barcode.Scale(bc128, int(BARCODE_WIDTH), int(BARCODE_HEIGHT))
		if err != nil {
			log.Fatal(errors.Stack(err))
		}

		// Convert the Go-image compatible barcode into a PDF image.
		img, err := creator.NewImageFromGoImage(bc)
		if err != nil {
			log.Fatal(errors.Stack(err))
		}

		// Scale the image to the desired width.
		img.ScaleToWidth(BARCODE_WIDTH)

		// Finally, draw the image in the proper location.
		c.MoveTo(x_, y_)
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
