/*
 * Get form field data for a specific field from a PDF file.
 *
 * Run as: go run pdf_form_get_field_data <input.pdf> [full field name]
 * If no field specified will output values for all fields.
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/timdrysdale/parselearn"
	unicommon "github.com/unidoc/unipdf/v3/common"
	extractor "github.com/unidoc/unipdf/v3/extractor"
	pdf "github.com/unidoc/unipdf/v3/model"
)

type cmdOptions struct {
	pdfPassword string
}

type ScanResult struct {
	ScanPerfect            bool   `csv:"ScanPerfect"`
	ScanRoated             bool   `csv:"ScanRotated"`
	ScanContrast           bool   `csv:"ScanContrast"`
	ScanFaint              bool   `csv:"ScanFaint"`
	ScanIncomplete         bool   `csv:"ScanIncomplete"`
	ScanBroken             bool   `csv:"ScanBroken"`
	ScanComment1           string `csv:"ScanComment1"`
	ScanComment2           string `csv:"ScanComment2"`
	HeadingPerfect         bool   `csv:"HeadingPerfect"`
	HeadingVerbose         bool   `csv:"HeadingVerbose"`
	HeadingNoLine          bool   `csv:"HeadingNoLine"`
	HeadingNoQuestion      bool   `csv:"HeadingNoQuestion"`
	HeadingNoExamNumber    bool   `csv:"HeadingNoExamNumber"`
	HeadingAnonymityBroken bool   `csv:"HeadingAnonymityBroken"`
	HeadingComment1        string `csv:"HeadingComment1"`
	HeadingComment2        string `csv:"HeadingComment2"`
	FilenamePerfect        bool   `csv:"FilenamePerfect"`
	FilenameVerbose        bool   `csv:"FilenameVerbose"`
	FilenameNoCourse       bool   `csv:"FilenameNoCourse"`
	FilenameNoID           bool   `csv:"FilenameNoID"`
	InputFile              string `csv:"InputFile"`
	Submission             parselearn.Submission
}

//type Submission struct {
//	FirstName          string  `csv:"FirstName"`
//	LastName           string  `csv:"LastName"`
//	Matriculation      string  `csv:"Matriculation"`
//	Assignment         string  `csv:"Assignment"`
//	DateSubmitted      string  `csv:"DateSubmitted"`
//	SubmissionField    string  `csv:"SubmissionField"`
//	Comments           string  `csv:"Comments"`
//	OriginalFilename   string  `csv:"OriginalFilename"`
//	Filename           string  `csv:"Filename"`
//	ExamNumber         string  `csv:"ExamNumber"`
//	MatriculationError string  `csv:"MatriculationError"`
//	ExamNumberError    string  `csv:"ExamNumberError"`
//	FiletypeError      string  `csv:"FiletypeError"`
//	FilenameError      string  `csv:"FilenameError"`
//	NumberOfPages      string  `csv:"NumberOfPages"`
//	FilesizeMB         float64 `csv:"FilesizeMB"`
//	NumberOfFiles      int     `csv:"NumberOfFiles"`
//}

func main() {
	// When debugging, enable debug-level logging via console:
	unicommon.SetLogger(unicommon.NewConsoleLogger(unicommon.LogLevelDebug))

	if len(os.Args) < 2 {
		fmt.Printf("Usage: gradex-extract outputPath inputPaths\n")
		os.Exit(1)
	}

	csvPath := os.Args[1]
	inputPaths := os.Args[2:]

	results := []ScanResult{}
	var opt cmdOptions

	files := make(map[string]map[int]string) //map of batchfilename->page->source files

	textfields := make(map[string]map[string]string) //flat key-val for whole batch file

	submissionsByOriginalFile := make(map[string]*parselearn.Submission)
	submissionsByFile := make(map[string]*parselearn.Submission)

	// iterate over the files in our list
	for _, inputPath := range inputPaths {

		// find out what original file each page came from
		// for now - we assume one text per page, and one page per file
		// because this is for interpreting montage output, only, at the moment

		if strings.Compare(filepath.Ext(inputPath), ".pdf") == 0 {
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
		if strings.Compare(filepath.Ext(inputPath), ".csv") == 0 {
			if subs, err := readIngestReport(inputPath); err == nil {
				for _, sub := range subs {
					submissionsByFile[sub.Filename] = sub
					submissionsByOriginalFile[sub.OriginalFilename] = sub
				}
			}
		}
	}

	// Now reconcile fields .... so we can assign to source file (original doc before batching)

	// map batchfilename->page->key->val
	organisedFields := make(map[string]map[int]map[string]string)

	for file, fields := range textfields {
		perFileMap := make(map[int]map[string]string)

		fmt.Printf("%s: %d\n", file, len(fields))
		for key, val := range fields {
			p, basekey := whatPageIsThisFrom(key)
			if _, ok := perFileMap[p]; !ok { //init map for this page if not present
				perFileMap[p] = make(map[string]string)
			}
			perFileMap[p][basekey] = val
		}

		organisedFields[file] = perFileMap

	}

	PrettyPrintStruct(organisedFields)

	// join the two maps to make a per-source-file report on results
	// abracadabra

	// we go by batchfile, and page, same in both maps, unless corrupt files
	// so either or ...
	for batchfile, pageToSourceFileMap := range files {

		fmt.Println(batchfile)

		for page, sourcefile := range pageToSourceFileMap {

			if _, ok := organisedFields[batchfile][page]; ok {
				organisedFields[batchfile][page]["SourceFile"] = sourcefile

				// find original submission details
				var submission *parselearn.Submission

				if sub, ok := submissionsByFile[sourcefile]; ok {
					submission = sub
				} else if sub, ok := submissionsByOriginalFile[sourcefile]; ok {
					submission = sub
				}

				thisScan := ScanResult{}
				thisScan.Submission = *submission

				results = append(results, thisScan) //ScanResult{Submission: *submission})
			}
		}
	}
	err := WriteResultsToCSV(results, csvPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func readIngestReport(inputPath string) ([]*parselearn.Submission, error) {
	subs := []*parselearn.Submission{}
	f, err := os.Open(inputPath)
	if err != nil {
		return subs, errors.New("can't open file")
	}
	defer f.Close()

	if err := gocsv.UnmarshalFile(f, &subs); err != nil { // Load subs
		return subs, errors.New("can't unmarshall from file")
	}

	return subs, nil
}

func WriteResultsToCSV(results []ScanResult, outputPath string) error {
	// wrap the marshalling library in case we need converters etc later
	file, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	return gocsv.MarshalFile(&results, file)
}

func whatPageIsThisFrom(key string) (int, string) {

	// fortunately, a rather fixed format! We get away with using prior knowledge for now
	if strings.HasPrefix(key, "page-000-") {
		basekey := strings.TrimPrefix(key, "page-000-")
		return 0, basekey
	}

	tokens := strings.Split(key, ".")
	if len(tokens) > 1 { //ignore the "docN" empty entries at the start of each page
		pageString := strings.TrimPrefix(tokens[0], "doc")
		pageInt, err := strconv.ParseInt(pageString, 10, 64)
		if err != nil {
			return -1, ""
		}
		basekey := strings.TrimPrefix(tokens[1], "page-000-")
		return int(pageInt), basekey
	}

	return -1, ""

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

func PrettyPrintStruct(layout interface{}) error {

	json, err := json.MarshalIndent(layout, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(json))
	return nil
}
