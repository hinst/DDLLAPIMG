package main

import (
	"bytes"
	"container/list"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

//---CODE REGION UNRELATED

func writeLine(text string) {
	fmt.Println(text)
}

const windowsLineEnding = "\x0d\x0a"

type stringFunc func() string

func assert(condition bool, errorMessage stringFunc) {
	if false == condition {
		panic(errorMessage())
	}
}

func sliceRangeInfoToText(length, start, end int) string {
	return "length is " + strconv.Itoa(length) + "range = [" + strconv.Itoa(start) + ", " + strconv.Itoa(end) + ")"
}

func assertStringSlice(a string, start, end int) {
	length := len(a)
	rangeIsCorrect := (0 <= start) && (end <= length) && (start <= end)
	errorMessage := func() string { return "Incorrect slice range;" + sliceRangeInfoToText(len(a), start, end) }
	assert(rangeIsCorrect, errorMessage)
}

func sliceString(a string, start, end int) string {
	assertStringSlice(a, start, end)
	defer func() {
		recovered := recover()
		if recovered != nil {
			panic("Error while slicing string: " + sliceRangeInfoToText(len(a), start, end))
		}
	}()
	result := a[start:end]
	return result
}

func (a findSectionResult) isFound() bool {
	return (a.startPosition >= 0) && (a.endPosition >= 0)
}

func (a findSectionResult) isEmpty() bool {
	return len(a.content) == 0
}

func stringsToText(strings []string) {
	var result bytes.Buffer
	result.WriteString("strings (" + strconv.Itoa(len(strings)) + ")")
	for i := range strings {
		s := strings[i]
		result.WriteString(s)
	}
}

func trim(a string) string {
	return strings.TrimSpace(a)
}

func replace(source, a, b string) string {
	return strings.Replace(source, a, b, -1)
}

func endsWith(text, ending string) bool {
	if len(ending) < len(text) {
		actualEnding := text[len(text)-len(ending):]
		return actualEnding == ending
	} else if len(ending) == len(text) {
		return text == ending
	} else {
		return false
	}
}

//---CODE REGION RELATED

type findSectionResult struct {
	// Start marker position.
	startPosition int
	// Section content (no markers).
	content string
	// End marker position.
	endPosition int
}

func findSection(text, sectionStartMarker, sectionEndMarker string) findSectionResult {
	result := findSectionResult{}
	result.startPosition = strings.Index(text, sectionStartMarker)
	result.endPosition = strings.Index(text, sectionEndMarker)
	if (result.startPosition >= 0) && (result.endPosition >= 0) {
		result.content = text[result.startPosition+len(sectionStartMarker) : result.endPosition]
	}
	return result
}

func extractSection(text *string, sectionStartMarker, sectionEndMarker string) findSectionResult {
	result := findSection(*text, sectionStartMarker, sectionEndMarker)
	if false == result.isEmpty() {
		textLeft := (*text)[:result.startPosition+len(sectionStartMarker)]
		textRight := (*text)[result.endPosition:]
		*text = textLeft + textRight
	}
	return result
}

type routineKind int

const (
	routineKindUnknown routineKind = iota
	routineKindFunction
	routineKindProcedure
)

func stringToRoutineKind(routineKind string) routineKind {
	switch routineKind {
	case "function":
		return routineKindFunction
	case "procedure":
		return routineKindProcedure
	default:
		return routineKindUnknown
	}
}

func (this routineKind) toText() string {
	switch this {
	case routineKindUnknown:
		return "RoutineKindUnknown"
	case routineKindFunction:
		return "RoutineKindFunction"
	case routineKindProcedure:
		return "RoutineKindProcedure"
	default:
		return "RoutineKind-no-such-value"
	}
}

func (this routineKind) toString() string {
	switch this {
	case routineKindFunction:
		return "function"
	case routineKindProcedure:
		return "procedure"
	default:
		return ""
	}
}

type routineHeader struct {
	routineKind routineKind
	// Like this: "sio_ioctl"
	routineName string
	// Like this: "(port, baud, mode: Longint): Longint; stdcall;"
	routineTail string
	// Like this: port, baud, mode
	routineArguments string
}

func getCleanArgumentsStringFromRoutineTail(tail string) string {
	cutRightFromColon := func(argument string) string {
		colonPosition := strings.Index(argument, ":")
		if colonPosition >= 0 {
			argument = argument[:colonPosition]
			argument = strings.TrimSpace(argument)
		}
		return argument
	}
	cutLeftToSpace := func(argument string) string {
		for {
			spacePosition := strings.Index(argument, " ")
			if spacePosition >= 0 {
				argument = argument[spacePosition+1:]
				argument = trim(argument)
			} else {
				break
			}
		}
		return argument
	}
	argumentsString := findSection(tail, "(", ")").content
	argumentsString = replace(argumentsString, ";", ",")
	arguments := strings.Split(argumentsString, ",")
	for i := range arguments {
		argument := arguments[i]
		argument = trim(argument)
		argument = cutRightFromColon(argument)
		argument = trim(argument)
		argument = cutLeftToSpace(argument)
		argument = trim(argument)
		arguments[i] = argument
	}
	result := strings.Join(arguments, ", ")
	return result
}

func (a *routineHeader) fillArguments() {
	a.routineArguments = getCleanArgumentsStringFromRoutineTail(a.routineTail)
}

func listToFunctionHeaders(list list.List) (result []routineHeader) {
	result = make([]routineHeader, 0, list.Len())
	for i := list.Front(); i != nil; i = i.Next() {
		result = append(result, i.Value.(routineHeader))
	}
	return
}

/*
func ToText(functionHeaders []FunctionHeader) {
	var result bytes.Buffer
	result.WriteString("function headers (" + len(functionHeaders) + ")")
	for i := range functionHeaders {
		result.WriteString(functionHeaders)
	}
}
*/

// Result "index" contains the array index in searchables. Set to -1 if no searchables found.
// Result "position" contains searchable position in text. Set to len(text) if no searchables found.
func findNearest(text string, searchables []string) (index, position int) {
	resultIndex := -1
	bestPosition := len(text)
	for i := range searchables {
		searchable := searchables[i]
		position := strings.Index(text, searchable)
		if position != -1 {
			if position < bestPosition {
				resultIndex = i
				bestPosition = position
			}
		}
	}
	index = resultIndex
	position = bestPosition
	return
}

var routineStartMarkers = []string{"function", "procedure"}

// Search for first function header in the text.
// Result "start" is set to -1 if no function headers were found.
func findFunctionHeader(functionHeadersText string) (start, end int) {
	index, position := findNearest(functionHeadersText, routineStartMarkers)
	if index >= 0 {
		start = position
		leftOffset := start + len(routineStartMarkers[index])
		functionHeadersText = functionHeadersText[leftOffset:]
		_, end = findNearest(functionHeadersText, routineStartMarkers)
		end = end + leftOffset
	} else {
		start = -1
		end = len(functionHeadersText)
	}
	return start, end
}

var routineNameTerminators = []string{" ", "(", ":", ";"}

func parseFunctionHeader(functionHeaderText string) routineHeader {
	result := routineHeader{}
	spacePosition := strings.Index(functionHeaderText, " ")
	if spacePosition >= 0 {
		routineKindString := functionHeaderText[0:spacePosition]
		result.routineKind = stringToRoutineKind(routineKindString)
		functionHeaderText = strings.TrimSpace(functionHeaderText[spacePosition+1:])
		_, routineNameEndPosition := findNearest(functionHeaderText, routineNameTerminators)
		if routineNameEndPosition >= 0 {
			result.routineName = strings.TrimSpace(functionHeaderText[0:routineNameEndPosition])
			functionHeaderText = strings.TrimSpace(functionHeaderText[routineNameEndPosition:])
			result.routineTail = strings.TrimSpace(functionHeaderText)
			result.fillArguments()
		} else {
			result.routineName = strings.TrimSpace(functionHeaderText)
		}
	}
	return result
}

func parseFunctionHeaders(functionHeadersText string) []routineHeader {
	var functionHeaders list.List
	for len(functionHeadersText) > 0 {
		start, end := findFunctionHeader(functionHeadersText)
		if start != -1 {
			functionHeaderText := trim(sliceString(functionHeadersText, start, end))
			functionHeader := parseFunctionHeader(functionHeaderText)
			functionHeaders.PushBack(functionHeader)
		} else {
			break
		}
		functionHeadersText = functionHeadersText[end:]
	}
	return listToFunctionHeaders(functionHeaders)
}

const functionHeadersStartMarker = "{$region function headers}"
const functionHeadersEndMarker = "{$endRegion function headers}"
const functionLoaderTemplateStartMarker = "{$region function loader template}"
const functionLoaderTemplateEndMarker = "{$endRegion function loader template}"
const deferredFunctionsStartMarker = "{$region deferred functions}"
const deferredFunctionsEndMarker = "{$endRegion deferred functions}"
const deferredFunctionsMarker = deferredFunctionsStartMarker + deferredFunctionsEndMarker

func processText(text string) string {
	sectionsFound := true
	headersSection := findSection(text, functionHeadersStartMarker, functionHeadersEndMarker)
	if false == headersSection.isFound() {
		sectionsFound = false
		writeLine("Error: function headers section not found; section markers are: '" + functionHeadersStartMarker + "', '" + functionHeadersEndMarker + "'")
	}
	loaderTemplateSection := findSection(text, functionLoaderTemplateStartMarker, functionLoaderTemplateEndMarker)
	if false == loaderTemplateSection.isFound() {
		sectionsFound = false
		writeLine("Error: function loader template section not found; section markers are: '" + functionLoaderTemplateStartMarker + "', '" + functionLoaderTemplateEndMarker + "'")
	}
	deferredFunctionsSection := extractSection(&text, deferredFunctionsStartMarker, deferredFunctionsEndMarker)
	if false == deferredFunctionsSection.isFound() {
		sectionsFound = false
		writeLine("Error: deferred functions section not found; section markers are: '" + deferredFunctionsStartMarker + "', '" + deferredFunctionsEndMarker + "'")
	}
	if sectionsFound {
		writeLine("Debug: template = '" + loaderTemplateSection.content + "'")
		functionHeaders := parseFunctionHeaders(headersSection.content)
		deferredLoadersText := generateDeferredLoaders(loaderTemplateSection.content, functionHeaders)
		text = replace(text, deferredFunctionsMarker, deferredFunctionsStartMarker+windowsLineEnding+deferredLoadersText+windowsLineEnding+deferredFunctionsEndMarker)
		writeLine("Debug: function headers found: " + strconv.Itoa(len(functionHeaders)))
	}
	return text
}

// Generate deferred function loader code from template
func (this routineHeader) generateDeferredLoaderText(template string) string {
	result := template
	result = replace(result, "$routineKind$", this.routineKind.toString())
	result = replace(result, "$routineName$", this.routineName)
	result = replace(result, "$routineTail$", this.routineTail)
	resultAssignmentPrefix := ""
	if this.routineKind == routineKindFunction {
		resultAssignmentPrefix = "result := "
	}
	result = replace(result, "$resultAssignmentPrefixIfFunction$", resultAssignmentPrefix)
	result = replace(result, "$routineArguments$", this.routineArguments)
	return result
}

func generateDeferredLoaders(template string, functionHeaders []routineHeader) string {
	var headersText bytes.Buffer
	for i := range functionHeaders {
		header := functionHeaders[i]
		headerText := header.generateDeferredLoaderText(template)
		headersText.WriteString(headerText)
		if false == endsWith(headerText, windowsLineEnding) {
			headersText.WriteString(windowsLineEnding)
		}
	}
	return headersText.String()
}

func main() {
	filePath := "MoxaApi.pas"
	writeLine("Now reading file " + filePath + "...")
	fileContent, readFileResult := ioutil.ReadFile(filePath)
	if readFileResult == nil {
		fileContentText := string(fileContent)
		processedText := processText(fileContentText)
		ioutil.WriteFile("generated"+filePath, []byte(processedText), os.ModePerm)
	}
}
