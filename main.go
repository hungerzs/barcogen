package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/jackmanlabs/errors"
	"github.com/jung-kurt/gofpdf"
	"log"
	"golang.org/x/image/tiff"
)

const (
	PAGE_WIDTH     float64 = 8.5
	PAGE_HEIGHT    float64 = 11.0
	BARCODE_HEIGHT float64 = 0.5    // Height of the barcode
	BARCODE_WIDTH  float64 = 2.0    // Width of the barcode
	LABEL_HEIGHT   float64 = 1.0    // Height of the label
	LABEL_MARGIN   float64 = 0.125  // Space between the columns of labels
	LABEL_PADDING  float64 = 0.125  // Padding between the content and the label edge
	LABEL_WIDTH    float64 = 2.625  // Width of the label
	OFFSET_X       float64 = 0.1875 // Left margin of the label page
	OFFSET_Y       float64 = 0.5    // Top margin of the label page
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

	pdf := gofpdf.New("P", "in", "Letter", "")

	for i := *start; i <= *end; i++ {

		// This creates a new page.
		if (i-*start)%30 == 0 {

			pdf.AddPage()

			// Draw a lines around labels for debugging
			var (
				x1, x2, y1, y2 float64
			)

			for col := 0.0; col < 3; col++ {

				// left edge
				x1 = OFFSET_X + col*(LABEL_WIDTH+LABEL_MARGIN)
				x2 = x1
				y1 = 0
				y2 = PAGE_HEIGHT
				pdf.Line(x1, y1, x2, y2)

				// right edge
				x1 = OFFSET_X + col*(LABEL_WIDTH+LABEL_MARGIN) + LABEL_WIDTH
				x2 = x1
				y1 = 0
				y2 = PAGE_HEIGHT
				pdf.Line(x1, y1, x2, y2)
			}

			for y := OFFSET_Y; y <= PAGE_HEIGHT; y += LABEL_HEIGHT {
				x1 = 0
				x2 = PAGE_WIDTH
				y1 = y
				y2 = y
				pdf.Line(x1, y1, x2, y2)
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
		bc, err := barcode.Scale(bc128, int(BARCODE_WIDTH*72), int(BARCODE_HEIGHT*72))
		if err != nil {
			log.Fatal(errors.Stack(err))
		}

		bcBuf := bytes.NewBuffer(nil)
		err = tiff.Encode(bcBuf, bc,&tiff.Options{})
		if err != nil {
			log.Fatal(errors.Stack(err))
		}

		bcOptions := gofpdf.ImageOptions{ImageType: "TIF", ReadDpi: true}
		pdf.RegisterImageOptionsReader(serial, bcOptions, bcBuf)
		pdf.ImageOptions(serial, x_, y_, BARCODE_WIDTH, BARCODE_HEIGHT, false, bcOptions, 0, "")

	}

	err := pdf.OutputFileAndClose("out.pdf")
	if err != nil {
		log.Fatal(errors.Stack(err))
	}

}
