/*
 * Get form field data for a specific field from a PDF file.
 *
 * Run as: go run pdf_form_get_field_data <input.pdf> [full field name]
 * If no field specified will output values for all fields.
 */

package main

import (
	pdf "github.com/georgekinnear/gradex-extract/pdfextract"
	"flag"
	"time"
	"os"
	"fmt"
)

func main() {

	var inputDir string
	flag.StringVar(&inputDir, "inputdir", "./", "path of the folder containing the PDF files to be processed (if in multimarker mode, will also check sub-folders with 'marker' in their name")
	
	var partsCSV string
	flag.StringVar(&partsCSV, "parts", "../parts_and_marks.csv", "path to the csv of parts and marks")

	flag.Parse()

	
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		// inputDir does not exist
		fmt.Println(err)
		os.Exit(1)
	}
	
	// see if the default CSV value needs to be changed - in multimarker mode, we expect it to be in the inputDir itself
	if partsCSV == "../parts_and_marks.csv" {
		if _, err := os.Stat(partsCSV); os.IsNotExist(err) {
			partsCSV = inputDir+"/parts_and_marks.csv"
		}
	}
	
	if _, err := os.Stat(partsCSV); os.IsNotExist(err) {
		fmt.Println("Could not locate", partsCSV)
		os.Exit(1)
	}
	parts := pdf.GetPartsAndMarks(partsCSV)
	pdf.PrettyPrintStruct(parts)
	
	
	report_time := time.Now().Format("2006-01-02-15-04-05")

	// Look at all PDFs in inputDir (including subdirectories)
	fmt.Println("Looking at input directory: ",inputDir)
	
	// Read the raw form values, and save them as a csv
	csv_path := fmt.Sprintf("%s/01_raw_form_values-%s.csv", inputDir, report_time)
	form_values := pdf.ReadFormsInDirectory(inputDir, csv_path)
	
	// Check the scripts are all from the same course
	coursecode := make(map[string]bool)
	for _, entry := range form_values {
		coursecode[entry.CourseCode] = true
	}
	if len(coursecode) != 1 {
		fmt.Println("Error - found scripts from multiple courses:",coursecode)
		os.Exit(1)
	}

	// Now summarise the marks and perform validation checks
	csv_path = fmt.Sprintf("%s/00_marks_summary-%s.csv", inputDir, report_time)
	pdf.ValidateMarking(form_values, parts, csv_path)

	os.Exit(1)

}
