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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"regexp"

	"github.com/gocarina/gocsv"
	"github.com/timdrysdale/parselearn"
	extractor "github.com/timdrysdale/unipdf/v3/extractor"
	pdf "github.com/timdrysdale/unipdf/v3/model"
)

type FormValues struct {
	CourseCode string `csv:"CourseCode"`
	Marker     string `csv:"Marker"`
	ExamNumber string `csv:"ExamNumber"`
	Field      string `csv:"Field"`
	Value      string `csv:"Value"`
}

// Structure for the optional reading a csv of parts and marks
type PaperStructure struct {
	Part       string  `csv:"part"`
	Marks      int     `csv:"marks"`
}

type cmdOptions struct {
	pdfPassword string
}

type ScanResult struct {
	ScanPerfect            bool   `csv:"ScanPerfect"`
	ScanRotated            bool   `csv:"ScanRotated"`
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
	BatchFile              string `csv:"BatchFile"`
	BatchPage              int    `csv:"BatchPage"`
	Submission             parselearn.Submission
}

func main() {

	multiMarker := flag.Bool("multimarker", false, "consolidate marks from multiple markers? (true/false)")

	var inputDir string
	flag.StringVar(&inputDir, "inputdir", "./", "path of the folder containing the PDF files to be processed (if in multimarker mode, will also check sub-folders with 'marker' in their name")

	flag.Parse()

	
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		// inputDir does not exist
		fmt.Println(err)
		os.Exit(1)
	}

	if *multiMarker {
		
		fmt.Println("Looking at input directory: ",inputDir)
		
		// Identify all the sub-directories in which marking is being done

		//  -  ?? walk over the inputDir to find directories with "Marker" in the name

		//  -  for each one, run readFormsInDirectory to get a csv+struct of the form values

		// For each marker, produce validation of their marking

		//  -  run validateMarking and save the resulting csv in their directory

		// Collate the marks from all markers, and save the resulting csv in inputDir

	} else {
		// Only considering a single marker, and we expect inputDir to be the folder containing their marked PDFs
		fmt.Println("Looking at input directory: ",inputDir)
		form_values := readFormsInDirectory(inputDir)

		//fmt.Println(PrettyPrintStruct(form_values))
		
		validation := validateMarking(form_values)
		
		fmt.Println(validation)
	}

	os.Exit(1)
/*
	csvPath := os.Args[1]
	inputDir := os.Args[2]

	// Find all PDFs in the inputDir
	err := ensureDir(inputDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var inputPaths = []string{}
	filepath.Walk(inputDir, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if filepath.Ext(f.Name()) == ".pdf" {
				inputPaths = append(inputPaths, f.Name())
			}
		}
		return nil
	})
	fmt.Println("input files: ", len(inputPaths))

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

		for key, val := range fields {
			p, basekey := whatPageIsThisFrom(key)
			if _, ok := perFileMap[p]; !ok { //init map for this page if not present
				perFileMap[p] = make(map[string]string)
			}
			perFileMap[p][basekey] = val
		}

		organisedFields[file] = perFileMap

	}

	// join the two maps to make a per-source-file report on results
	// abracadabra

	// we go by batchfile, and page, same in both maps, unless corrupt files
	// so either or ...
	for batchfile, pageToSourceFileMap := range files {

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
				if submission != nil {
					thisScan.Submission = *submission
				}
				thisScan.BatchFile = batchfile
				thisScan.BatchPage = page + 1 //humans often start thinking at page 1

				insertCheckReport(&thisScan, organisedFields[batchfile][page])

				results = append(results, thisScan) //ScanResult{Submission: *submission})
			}
		}
	}

	PrettyPrintStruct(results)
	err = WriteResultsToCSV(results, csvPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
*/
}

func readFormsInDirectory(formsPath string) []FormValues {

	form_vals := []FormValues{}
	
	filename_examno, err := regexp.Compile("(B[0-9]{6})-mark.pdf")
	
	var num_scripts int
	filepath.Walk(formsPath, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			proper_filename := filename_examno.MatchString(f.Name())
			if proper_filename {
				extracted_examno := filename_examno.FindStringSubmatch(f.Name())[1]
				// TODO - check that extracted_examno matches the one on the script!
				fmt.Println(extracted_examno)
				form_vals = append(form_vals, readFormFromPDF(path)...)
				num_scripts++
			} else {
				fmt.Println("Malformed filename: ", f.Name())
			}
		}
		return nil
	})
	
	
	file, err := os.OpenFile("output.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()
	gocsv.MarshalFile(form_vals, file)
	
	return form_vals
}

func readFormFromPDF(path string) []FormValues {

	all_form_vals := []FormValues{}

	form_vals := FormValues{}
	
	// Read the text values from the PDF
	var opt cmdOptions
	text_data, _ := getText(path, opt)
	//PrettyPrintStruct(text_data)
	
	form_vals.Marker = extractMarkerInitials(text_data)
	form_vals.CourseCode = extractCourseCode(text_data)
	form_vals.ExamNumber = extractExamNumber(text_data)
	
	//fmt.Println("Course code: ",form_vals.CourseCode)
	//fmt.Println("Marker initials: ",form_vals.Marker)
	//fmt.Println("Exam number: ",form_vals.ExamNumber)
	
	// Read the form values from the PDF
	field_data, _ := mapPdfFieldData(path)
	//PrettyPrintStruct(field_data)
	
	for key, val := range field_data {
		this_form_entry := form_vals
		this_form_entry.Field = key
		this_form_entry.Value = val
		all_form_vals = append(all_form_vals, this_form_entry)
	}
	
	//PrettyPrintStruct(all_form_vals)
	
	return all_form_vals
}

func validateMarking([]FormValues) (error) {
	
	
	
	return nil
}

func extractMarkerInitials(pdf_text map[int]string) string {
	// TODO - this could check *all* pages to make sure the initials are consistent,
	// but let's be lazy and just use the first page
	raw_string_p1 := pdf_text[0]
	
	// initials appear as the second line of text https://regex101.com/r/9GjHTM/9
	findinitials, _ := regexp.Compile(".*\n([a-zA-Z]+)\n")
	return findinitials.FindStringSubmatch(raw_string_p1)[1]
}

func extractCourseCode(pdf_text map[int]string) string {
	// TODO - this could check *all* pages for consistency
	// but let's be lazy and just use the first page
	raw_string_p1 := pdf_text[0]
	
	// course code is the first word of text https://regex101.com/r/9GjHTM/10
	findinitials, _ := regexp.Compile("([a-zA-Z0-9]+) ")
	return findinitials.FindStringSubmatch(raw_string_p1)[1]
}

func extractExamNumber(pdf_text map[int]string) string {
	// TODO - this could check *all* pages for consistency
	// but let's be lazy and just use the first page
	raw_string_p1 := pdf_text[0]
	
	// exam number is the last word on the first line https://regex101.com/r/9GjHTM/11
	findexamno, _ := regexp.Compile(" ([a-zA-Z0-9]+)\n")
	return findexamno.FindStringSubmatch(raw_string_p1)[1]
}

func boolVal(str string) bool {
	return strings.Compare(str, "") != 0
}

func insertCheckReport(scan *ScanResult, fields map[string]string) {

	for k, v := range fields {
		switch k {
		case "filename-no-course":
			scan.FilenameNoCourse = boolVal(v)
		case "filename-no-id":
			scan.FilenameNoID = boolVal(v)
		case "filename-perfect":
			scan.FilenamePerfect = boolVal(v)
		case "filename-verbose":
			scan.FilenameVerbose = boolVal(v)
		case "heading-anonymity-broken":
			scan.HeadingAnonymityBroken = boolVal(v)
		case "heading-comment-1":
			scan.HeadingComment1 = v
		case "heading-comment-2":
			scan.HeadingComment2 = v
		case "heading-no-exam-number":
			scan.HeadingNoExamNumber = boolVal(v)
		case "heading-no-line":
			scan.HeadingNoLine = boolVal(v)
		case "heading-no-question":
			scan.HeadingNoQuestion = boolVal(v)
		case "heading-perfect":
			scan.HeadingPerfect = boolVal(v)
		case "heading-verbose":
			scan.HeadingVerbose = boolVal(v)
		case "scan-broken":
			scan.ScanBroken = boolVal(v)
		case "scan-comment-1":
			scan.ScanComment1 = v
		case "scan-comment-2":
			scan.ScanComment2 = v
		case "scan-contrast":
			scan.ScanContrast = boolVal(v)
		case "scan-faint":
			scan.ScanFaint = boolVal(v)
		case "scan-incomplete":
			scan.ScanIncomplete = boolVal(v)
		case "scan-perfect":
			scan.ScanPerfect = boolVal(v)
		case "scan-rotated":
			scan.ScanRotated = boolVal(v)
		}

	}

}

//	        "filename-no-course": "x",
//			"filename-no-id": "x",
//			"filename-perfect": "",
//			"filename-verbose": "x",
//			"heading-anonymity-broken": "",
//			"heading-comment-1": "",
//			"heading-comment-2": "",
//			"heading-no-exam-number": "",
//			"heading-no-line": "",
//			"heading-no-question": "",
//			"heading-perfect": "x",
//			"heading-verbose": "",
//			"scan-broken": "",
//			"scan-comment-1": "",
//			"scan-comment-2": "",
//			"scan-contrast": "",
//			"scan-faint": "",
//			"scan-incomplete": "",
//			"scan-perfect": "x",
//			"scan-rotated": ""
//		},
//
//type ScanResult struct {
//	ScanPerfect            bool   `csv:"ScanPerfect"`
//	ScanRoated             bool   `csv:"ScanRotated"`
//	ScanContrast           bool   `csv:"ScanContrast"`
//	ScanFaint              bool   `csv:"ScanFaint"`
//	ScanIncomplete         bool   `csv:"ScanIncomplete"`
//	ScanBroken             bool   `csv:"ScanBroken"`
//	ScanComment1           string `csv:"ScanComment1"`
//	ScanComment2           string `csv:"ScanComment2"`
//	HeadingPerfect         bool   `csv:"HeadingPerfect"`
//	HeadingVerbose         bool   `csv:"HeadingVerbose"`
//	HeadingNoLine          bool   `csv:"HeadingNoLine"`
//	HeadingNoQuestion      bool   `csv:"HeadingNoQuestion"`
//	HeadingNoExamNumber    bool   `csv:"HeadingNoExamNumber"`
//	HeadingAnonymityBroken bool   `csv:"HeadingAnonymityBroken"`
//	HeadingComment1        string `csv:"HeadingComment1"`
//	HeadingComment2        string `csv:"HeadingComment2"`
//	FilenamePerfect        bool   `csv:"FilenamePerfect"`
//	FilenameVerbose        bool   `csv:"FilenameVerbose"`
//	FilenameNoCourse       bool   `csv:"FilenameNoCourse"`
//	FilenameNoID           bool   `csv:"FilenameNoID"`
//	InputFile              string `csv:"InputFile"`
//	Submission             parselearn.Submission
//}
//

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
		return int(pageInt) - 1, basekey //doc nums are out by 1 from our page index
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

// pr-pal @ https://stackoverflow.com/questions/37932551/mkdir-if-not-exists-using-golang
func ensureDir(dirName string) error {

	err := os.Mkdir(dirName, 0700) //probably umasked with 22 not 02

	os.Chmod(dirName, 0700)

	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}

}
