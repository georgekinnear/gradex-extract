/*
 * Get form field data for a specific field from a PDF file.
 *
 * Run as: go run pdf_form_get_field_data <input.pdf> [full field name]
 * If no field specified will output values for all fields.
 */

package main

import (
	"encoding/json"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"regexp"
	"time"
	"sort"

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
	
	var partsCSV string
	flag.StringVar(&partsCSV, "parts", "../parts_and_marks.csv", "path to the csv of parts and marks")

	flag.Parse()

	
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		// inputDir does not exist
		fmt.Println(err)
		os.Exit(1)
	}
	
	// Try to find parts_and_marks.csv
	if *multiMarker && partsCSV == "../parts_and_marks.csv" {
		// see if the default CSV value needs to be changed - in multimarker mode, we expect it to be in the current directory instead
		if _, err := os.Stat(partsCSV); os.IsNotExist(err) {
			partsCSV = "parts_and_marks.csv"
		}
	}
	if _, err := os.Stat(partsCSV); os.IsNotExist(err) {
		fmt.Println("Could not locate", partsCSV)
		os.Exit(1)
	}
	parts := getPartsAndMarks(partsCSV)
	PrettyPrintStruct(parts)
	
	
	report_time := time.Now().Format("2006-01-02-15-04-05")

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
		
		// Read the raw form values, and save them as a csv in the same folder as the scripts
		csv_path := fmt.Sprintf("%s/01_raw_form_values-%s.csv", inputDir, report_time)
		form_values := readFormsInDirectory(inputDir, csv_path)

		// Now summarise the marks and perform validation checks
		csv_path = fmt.Sprintf("%s/00_marks_summary-%s.csv", inputDir, report_time)
		validation := validateMarking(form_values, parts, csv_path)
		
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

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func getPartsAndMarks(csv_path string) []*PaperStructure {
	
	marksFile, err := os.OpenFile(csv_path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println("File: ",csv_path, err)
		panic(err)
	}
	defer marksFile.Close()

	parts := []*PaperStructure{}
	if err := gocsv.UnmarshalFile(marksFile, &parts); err != nil {
		panic(err)
	}
	return parts
}

func readFormsInDirectory(formsPath string, outputCSV string) []FormValues {

	form_vals := []FormValues{}
	
	filename_examno, err := regexp.Compile("(B[0-9]{6})-mark.pdf")
	
	var num_scripts int
	filepath.Walk(formsPath, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if filepath.Ext(f.Name()) != ".pdf" {
				return nil
			}
			proper_filename := filename_examno.MatchString(f.Name())
			if proper_filename {
				extracted_examno := filename_examno.FindStringSubmatch(f.Name())[1]
				vals_on_this_form := readFormFromPDF(path)
				// check that extracted_examno matches the one on the script!
				if vals_on_this_form[0].ExamNumber != extracted_examno {
					fmt.Println("Exam number mismatch: file",path,"has value",vals_on_this_form[0].ExamNumber)
				}
				
				form_vals = append(form_vals, vals_on_this_form...)
				num_scripts++
			} else {
				fmt.Println("Malformed filename: ", f.Name())
			}
		}
		return nil
	})
	
	
	file, err := os.OpenFile(outputCSV, os.O_RDWR|os.O_CREATE, os.ModePerm)
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
	
	var form_values int
	for key, val := range field_data {
		// TODO - perhaps don't bother storing the empty values? consider moving all this inside the "if"
		this_form_entry := form_vals
		this_form_entry.Field = key
		this_form_entry.Value = val
		all_form_vals = append(all_form_vals, this_form_entry)
		if hasContent(val) {
			form_values++
		}
	}
	
	fmt.Printf("%s has %d entries\n", form_vals.ExamNumber, form_values)
	//PrettyPrintStruct(all_form_vals)
	
	return all_form_vals
}

func validateMarking(form_values []FormValues, parts []*PaperStructure, outputCSV string) (error) {
	
	// understand the parts structure
	// NB - not really necessary, maybe can just do parts[3].Mark where 3 = the integer used in the PDF form ID
	marks_available := make(map[int]int)
	part_name := make(map[int]string)
	part_to_marks := make(map[string]int)
	for pnum, part := range parts {
		if part.Part != "" {
			marks_available[pnum] = part.Marks
			part_name[pnum] = part.Part
			part_to_marks[part.Part] = part.Marks
		}
	}
	fmt.Println(marks_available,"\n", part_name)
	
	// Set up maps to store data
	mark_details := make(map[string]map[string][]string) // mark_details[ExamNo][part] = [4,5,6]
	validation := make(map[string][]string) // validation[ExamNo] = ["1a has no mark", "1b noninteger mark"]
	marks_on_page := make(map[string]map[int]int) // marks_on_page[ExamNo][1] = 0
	marks_awarded := make(map[string]int) // marks_awarded[part] = 50 - sum of all student marks on this question
	marks_awarded_count := make(map[string]int) // marks_awarded[part] = 5 - number of students awarded marks
	bad_pages := make(map[string][]int) // bad_pages[ExamNo] = [1,4,5]
	
	for _, entry := range form_values {
		ExamNo := entry.ExamNumber
		if !strings.Contains(entry.Field, "page") {
			continue // quietly skip fields that don't have a page
		}
		page, field_name := whatPageIsThisFrom(entry.Field)
		
		// Prepare nested maps to receive values
		if _, ok := marks_on_page[ExamNo][page]; !ok {
			if _, ok := marks_on_page[ExamNo]; !ok {
				marks_on_page[ExamNo] = make(map[int]int)
			}
			marks_on_page[ExamNo][page] = 0
		}
		if mark_details[ExamNo] == nil {
			mark_details[ExamNo] = make(map[string][]string)
		}
		
		// Bad Page has been selected
		if field_name == "page-bad" && hasContent(entry.Value) {
			bad_pages[ExamNo] = append(bad_pages[ExamNo], page)
			marks_on_page[ExamNo][page]++
		}
		
		// Page Seen has been selected
		if field_name == "page-seen" && hasContent(entry.Value) {
			marks_on_page[ExamNo][page]++
		}
		
		// Marks Awarded field has been completed
		if strings.HasPrefix(field_name, "qn-part-mark-") && hasContent(entry.Value) {
			partnum, _ := strconv.Atoi(strings.TrimPrefix(field_name, "qn-part-mark-"))
			partname := part_name[partnum]
			part_max := marks_available[partnum]
	
			// Prepare the nested maps to receive values
			if validation[ExamNo] == nil {
				validation[ExamNo] = []string{}
			}
			if mark_details[ExamNo][partname] == nil {
				mark_details[ExamNo][partname] = []string{}
			}
			
			// Get the integer value
			var mark_awarded int
			if intval, err := strconv.Atoi(entry.Value); err == nil {
				mark_awarded = intval
				marks_awarded[partname] = marks_awarded[partname] + intval
				marks_awarded_count[partname]++
			} else {
				validation[ExamNo] = append(validation[ExamNo], partname+": noninteger mark")
			}
			
			// Validation of the value
			if mark_awarded > part_max {
				validation[ExamNo] = append(validation[ExamNo], partname+": max mark is "+strconv.Itoa(part_max))			
			}
			
			mark_details[ExamNo][partname] = append(mark_details[ExamNo][partname], entry.Value)
			marks_on_page[ExamNo][page]++
			
		}
		
	}
	
	fmt.Println(marks_awarded)
	
	// Carry out further validation of the marks
	// Also prepare the mark cells of the CSV
	mark_summary := make(map[string]map[string]string) // mark_summary[ExamNo][part] = "4+5" or "4" or "2.5"
	for ExamNo, marks_by_part := range mark_details {
		
		fmt.Println(ExamNo)
		PrettyPrintStruct(marks_by_part)
		
		// Prepare the nested maps to receive values
		if validation[ExamNo] == nil {
			validation[ExamNo] = []string{}
		}
		mark_summary[ExamNo] = make(map[string]string)
		
		// No need to add any more details for scripts that have no marks allocated
		if marks_on_page[ExamNo] == nil {
			continue
		}
		script_has_been_marked := false
		for _, pmarks := range marks_on_page[ExamNo] {
			if pmarks > 0 {
				script_has_been_marked = true
				//break
			}
		}
		if !script_has_been_marked {
			mark_summary[ExamNo]["Unmarked"] = "Unmarked"
			continue
		}
		
		// Further validation of each part
		for _, pname := range part_name {
		
			// Represent a lack of marks by an empty list
			if marks_by_part[pname] == nil {
				marks_by_part[pname] = []string{}
			}
		
			// If marks have been awarded to at least one student for this part, check that this student has a mark too
			if  _, ok := marks_awarded[pname]; ok {
				if len(marks_by_part[pname]) == 0 {
					validation[ExamNo] = append(validation[ExamNo], pname+": not marked")
				}
			}
		
			// Warn if marks are awarded on more than 1 occasion
			if len(marks_by_part[pname]) > 1 {
				validation[ExamNo] = append(validation[ExamNo], pname+": multiple marks")
			}
			
			mark_summary[ExamNo][pname] = strings.Join(marks_by_part[pname], " + ")
		
		}
		
		// put the validation messages into alphabetical order
		sort.Strings(validation[ExamNo])
		mark_summary[ExamNo]["Validation"] = strings.Join(validation[ExamNo], "; ")
		
		
		// Unmarked Pages - add column to mark_summary
		unmarked_pages := []int{}
		// Determine unmarked pages
		for pagenum, markings := range marks_on_page[ExamNo] {
			if markings == 0 {
				unmarked_pages = append(unmarked_pages, pagenum)
			}
		}
		sort.Ints(unmarked_pages)
		mark_summary[ExamNo]["Unmarked Pages"] = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(unmarked_pages)), ", "), "[]") // https://stackoverflow.com/a/37533144
		
		// Bad Pages - add column to mark_summary
		sort.Ints(bad_pages[ExamNo])
		mark_summary[ExamNo]["Bad Pages"] = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(bad_pages[ExamNo])), ", "), "[]") // https://stackoverflow.com/a/37533144
		
		PrettyPrintStruct(mark_summary[ExamNo])
		PrettyPrintStruct(validation[ExamNo])
		
	}
	
	/* 
	|   Produce the CSV output
	*/
	
	fmt.Println("MARK SUMMARY")
	PrettyPrintStruct(mark_summary)
	PrettyPrintStruct(part_name)
	
	file, err := os.OpenFile(outputCSV, os.O_RDWR|os.O_CREATE, os.ModePerm)
	check(err)
	defer file.Close()
	w := csv.NewWriter(file)
	
	// First come up with the csv_headers - have to work pretty hard to get that as a slice in the right order
/*	keys := make([]string, 0, len(mark_summary))
	values := make([]map[string]string, 0, len(mark_summary))
	for k, v := range mark_summary {
		keys = append(keys, k)
		values = append(values, v)
	}	
	csv_headers := []string{}
	for k, _ := range values[0] {
		if strings.Contains(k, "Page") { continue } // Manually add those to the end
		csv_headers = append(csv_headers, k)
	}
	sort.Strings(csv_headers)
	csv_headers = append([]string{"Exam Number"}, csv_headers...)
	csv_headers = append(csv_headers, []string{"Unmarked Pages", "Bad Pages"}...)
	
	
*/
	partnames := make([]string, 0, len(part_name))

	for  _, value := range part_name {
	   partnames = append(partnames, value)
	}
	sort.Strings(partnames)
	csv_headers := append([]string{"Exam Number"}, partnames...)
	csv_headers = append(csv_headers, []string{"Total", "Validation", "Unmarked Pages", "Bad Pages"}...)

	//
	// Write the header and stats summary rows
	err = w.Write(append([]string{""}, partnames...))
	check(err)
	
	// Add a row showing what each question is marked out of
	row_outof := []string{"out of:"}
	for _, val := range csv_headers {
		if outof, ok := part_to_marks[val]; ok {
			row_outof = append(row_outof, fmt.Sprintf("%v", outof))
		}
	}
	err = w.Write(row_outof)
	check(err)
	// Add a row showing what each question is marked out of
	row_means := []string{"mean:"}
	for _, val := range csv_headers {
		if _, ok := marks_awarded[val]; ok {
			fmt.Println(marks_awarded[val], "/", marks_awarded_count[val], "=", float64(marks_awarded[val])/float64(marks_awarded_count[val]))
			row_means = append(row_means, fmt.Sprintf("%.2f", float64(marks_awarded[val])/float64(marks_awarded_count[val])))
		}
	}
	err = w.Write(row_means)
	check(err)
	
	
	// Separate the Validation/Complete blocsk and sort both lists of students by Exam Number
	student_records_invalid := []string{}
	student_records_valid := []string{}
	student_records_unmarked := []string{}
	for enum, _ := range mark_summary {
		if (len(mark_summary[enum]["Validation"]) +
			len(mark_summary[enum]["Unmarked Pages"])) >0 {
			student_records_invalid = append(student_records_invalid, enum)
		} else if mark_summary[enum]["Unmarked"] == "Unmarked" {
			student_records_unmarked = append(student_records_unmarked, enum)
		} else {
			student_records_valid = append(student_records_valid, enum)		
		}
	}
	sort.Strings(student_records_invalid)
	sort.Strings(student_records_valid)
	sort.Strings(student_records_unmarked)

	// Print each row for the invalid records - again, have to work hard
	// to make sure the slice for each student is in the correct order
	err = w.Write([]string{""}) // blank row
	err = w.Write([]string{"Validation problems ("+strconv.Itoa(len(student_records_invalid))+" scripts):"})
	check(err)
	err = w.Write(csv_headers)
	check(err)
	for _, ExamNo := range student_records_invalid {
		record := []string{fmt.Sprintf("%v", ExamNo)}
		for _, val := range csv_headers {
			if val == "Exam Number" { continue }
			record = append(record, fmt.Sprintf("%v", mark_summary[ExamNo][val]))
		}
		err := w.Write(record)
		check(err)
	}

	
	// Now do the valid ones
	err = w.Write([]string{""}) // blank row
	err = w.Write([]string{"Marking completed ("+strconv.Itoa(len(student_records_valid))+" scripts):"})
	check(err)
	err = w.Write(csv_headers)
	check(err)
	for _, ExamNo := range student_records_valid {
		record := []string{fmt.Sprintf("%v", ExamNo)}
		for _, val := range csv_headers {
			if val == "Exam Number" { continue }
			record = append(record, fmt.Sprintf("%v", mark_summary[ExamNo][val]))
		}
		err := w.Write(record)
		check(err)
	}
	
	// TODO - now list all the "yet to be marked" ones
	
	// Now do the unmarked ones
	err = w.Write([]string{""}) // blank row
	err = w.Write([]string{"Yet to be marked ("+strconv.Itoa(len(student_records_unmarked))+" scripts):"})
	check(err)
	for _, ExamNo := range student_records_unmarked {
		record := []string{fmt.Sprintf("%v", ExamNo)}
		err := w.Write(record)
		check(err)
	}
/*	
	// Now loop through mark_summary and print out each row - again, have to work hard
	// to make sure the slice for each student is in the correct order
	for ExamNo, recordset := range mark_summary {
		record := []string{fmt.Sprintf("%v", ExamNo)}
		for _, val := range csv_headers {
			if val == "Exam Number" { continue }
			record = append(record, fmt.Sprintf("%v", recordset[val]))
		}
		err := w.Write(record)
		check(err)
	}
*/
	w.Flush()
	
	
	
	//fmt.Println(mark_summary)
	//PrettyPrintStruct(mark_details)
	//PrettyPrintStruct(validation)
	//PrettyPrintStruct(marks_on_page)
	
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

func hasContent(str string) bool {
	return strings.Compare(str, "") != 0
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
/*	if strings.HasPrefix(key, "page-000-") {
		basekey := strings.TrimPrefix(key, "page-000-")
		return 1, basekey
	}
	*/
	// Pick out page number and basekey https://regex101.com/r/vGyDbg/1
	parse_field_name, _ := regexp.Compile(".*page-([0-9]+)-(.*)")
	parsed_key := parse_field_name.FindStringSubmatch(key)
	parsed_pageno, err := strconv.Atoi(parsed_key[1])
	if err != nil {
		return -1, ""
	}
	return parsed_pageno + 1, parsed_key[2] // the basekey is the 2nd submatch

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
