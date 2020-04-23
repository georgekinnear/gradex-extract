/*
 * Get form field data for a specific field from a PDF file.
 *
 * Run as: go run pdf_form_get_field_data <input.pdf> [full field name]
 * If no field specified will output values for all fields.
 */

package main

import (
	"errors"
	"fmt"
	"os"

	unicommon "github.com/unidoc/unipdf/v3/common"
	extractor "github.com/unidoc/unipdf/v3/extractor"
	pdf "github.com/unidoc/unipdf/v3/model"
)

type cmdOptions struct {
	pdfPassword string
}

type ScanResult struct {
	ScanPerfect            bool
	ScanRoated             bool
	ScanContrast           bool
	ScanFaint              bool
	ScanIncomplete         bool
	ScanBroken             bool
	ScanComment1           string
	ScanComment2           string
	HeadingPerfect         bool
	HeadingVerbose         bool
	HeadingNoLine          bool
	HeadingNoQuestion      bool
	HeadingNoExamNumber    bool
	HeadingAnonymityBroken bool
	HeadingComment1        string
	HeadingComment2        string
	FilenamePerfect        bool
	FilenameVerbose        bool
	FilenameNoCourse       bool
	FilenameNoId           bool
	InputFile              string
}

func main() {
	// When debugging, enable debug-level logging via console:
	unicommon.SetLogger(unicommon.NewConsoleLogger(unicommon.LogLevelDebug))

	if len(os.Args) < 2 {
		fmt.Printf("Usage: gradex-extract outputPath inputPaths\n")
		os.Exit(1)
	}

	//csvPath := os.Args[1]
	inputPaths := os.Args[2:]

	//results := []ScanResult{}
	var opt cmdOptions

	files := make(map[string]map[int]string) //map of source files, by batch file + page

	textfields := make(map[string]map[string]string) //flat key-val for whole batch file

	// iterate over the files in our list
	for _, inputPath := range inputPaths {

		// find out what original file each page came from
		// for now - we assume one text per page, and one page per file
		// because this is for interpreting montage output, only, at the moment
		texts, err := getText(inputPath, opt)
		if err == nil {
			//fmt.Println(texts)
			files[inputPath] = texts
		}

		fields, err := mapPdfFieldData(inputPath)
		if err == nil {
			textfields[inputPath] = fields
		}

	}

	fmt.Println(files)
	fmt.Println(textfields)

	/*fieldName := ""
	if len(os.Args) >= 3 {
		fieldName = os.Args[2]
	}

	err := printPdfFieldData(pdfPath, fieldName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var opt cmdOptions

	fmt.Printf("Input file: %s\n", pdfPath)
	err = inspectPdf(pdfPath, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}*/

}

func printPdfFieldData(inputPath, targetFieldName string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	fmt.Printf("Input file: %s\n", inputPath)

	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return err
	}

	acroForm := pdfReader.AcroForm
	if acroForm == nil {
		fmt.Printf(" No formdata present\n")
		return nil
	}

	match := false
	fields := acroForm.AllFields()
	for _, field := range fields {
		fullname, err := field.FullName()
		if err != nil {
			return err
		}
		if fullname == targetFieldName || targetFieldName == "" {
			match = true
			if field.V != nil {
				fmt.Printf("Field '%s': '%v' (%T)\n", fullname, field.V, field.V)
			} else {
				fmt.Printf("Field '%s': not filled\n", fullname)
			}
		}
	}

	if !match {
		return errors.New("field not found")
	}
	return nil
}

func mapPdfFieldData(inputPath string) (map[string]string, error) {

	textfields := make(map[string]string)

	f, err := os.Open(inputPath)
	if err != nil {
		return textfields, errors.New(fmt.Sprintf("Problem opening file %s", inputPath))
	}
	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return textfields, errors.New(fmt.Sprintf("Problem creating reader %s", inputPath))
	}

	acroForm := pdfReader.AcroForm
	if acroForm == nil {
		return textfields, nil
	}

	fields := acroForm.AllFields()
	for _, field := range fields {
		fullname, err := field.FullName()
		if err != nil {
			continue
		}

		val := ""

		if field.V != nil {
			val = field.V.String()
		}

		textfields[fullname] = val

	}

	return textfields, nil
}

func getText(inputPath string, opt cmdOptions) (map[int]string, error) {

	texts := make(map[int]string)

	f, err := os.Open(inputPath)
	if err != nil {
		return texts, err
	}
	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return texts, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return texts, err
	}

	// Try decrypting with an empty one.
	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(opt.pdfPassword))
		if err != nil {
			return texts, err
		}

		if !auth {
			return texts, errors.New("Unable to decrypt password protected file - need to specify pass to Decrypt")
		}
	}
	for p, page := range pdfReader.PageList {

		ex, err := extractor.New(page)

		if err == nil {

			text, err := ex.ExtractText()
			if err == nil {
				texts[p] = text
			}

		}
	}

	return texts, nil

}
