/*
Copyright 2023 Tihomir Magdic. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/

/*
Package bigfloat implements big float number functionalities: adding, substraction, multiplication, division, rounding, truncation, converting to/from string, comparison, automatic precision.

Also, supports repeating decimals with various formatting.

BigFloat variables can be declared as:

	n1, _ := bigfloat.Set("7.005")
	n2, _ := bigfloat.Set(4)
	n3 := bigfloat.New() // zero value is 0

Operations are implemented as methods:

	n3.Add(n1, n2)
	n3.Sub(n1, n2)
	n3.Mul(n1, n2)
	n3.Div(n1, n2)

Methods of this form typically return the incoming receiver, to enable simple call chaining:

	n3.Mul(n1, n2).Sub(n3, n1)

BigFloat implements stringer so it can be simple printed as:

	fmt.Printf("%v\n", n3)

Division supports remainders:

	n1.Set(23)
	n2.Set(-11)

	_, remainder, _ := n3.DivMod(n1, n2)
	fmt.Printf("%v div %v = %v (remainder: %v)\n", n1, n2, n3, remainder)

	// Output: 23 div -11 = -2 (remainder: 1)

The division also supports arbitrary decimals:

	n3.Div(n1, n2, bigfloat.WithDivDecimalPlaces(10))
	fmt.Printf("%v / %v = %v\n", n1, n2, n3)

	// Output: 23 / -11 = -2.0909090909

Repeating decimals:

	_, repeatingDecimals, _ := n3.Div(n1, n2)
	fmt.Printf("%v / %v = %v\n", n1, n2, n3.StringF(repeatingDecimals))

	// Output: 23 / -11 = -2.(09)

Repeating decimals with different formatting:

	fmt.Printf("abs = %v\n", n3.Abs().StringF(repeatingDecimals, bigfloat.WithRepeatingOptions("r", ""), bigfloat.ForceSign(true)))

	// Output: abs = +2.r09

Rounding:

	n3.Set("1.75125")
	d := 4
	fmt.Printf("%v\n", n3.Round(d))

	// Output: 1.7513

Truncate:

	fmt.Printf("trunc(%v) = ", n3)
	fmt.Printf("%v\n", n3.Trunc())

	// Output: trunc(1.75130) = 1.00000
*/
package bigfloat

import (
	"bigfloat/stranalyzer"
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

/*
Function type for division operation.

See: Div
*/
type DivOption func(*divOptionsType)

type divOptionsType struct {
	decimalPlaces    int
	maxDecimalPlaces int
}

/*
Function defines precision in division operation.

Recommended to use repeating decimals.
*/
func WithDivDecimalPlaces(decimalPlaces int) DivOption {
	return func(ro *divOptionsType) {
		ro.decimalPlaces = decimalPlaces
	}
}

/*
Function defines maximum decimals in division
Effective with very long decimals
*/
func WithDivMaxDecimalPlaces(maxDecimalPlaces int) DivOption {
	return func(ro *divOptionsType) {
		ro.maxDecimalPlaces = maxDecimalPlaces
	}
}

/*
Function type for rounding option.
*/
type RoundOption func(*roundOptionsType)

type roundOptionsType struct {
	decimalPlaces int
}

/*
Function defines number of decimal places in rounding

See: Round
*/
func WithDecimalPlaces(decimalPlaces int) RoundOption {
	return func(ro *roundOptionsType) {
		ro.decimalPlaces = decimalPlaces
	}
}

/*
Function type for repeating options
*/
type RepeatingOptions func(*repeatingOptionsType)

type repeatingOptionsType struct {
	indicatorStart string
	indicatorEnd   string
}

/*
Function for repeating options

-indicatorStart is placed before first repeating decimal - default is '('
-indicatorEnd is placed after last repeating decimal - default is ')'
*/
func WithRepeatingOptions(indicatorStart, indicatorEnd string) RepeatingOptions {
	return func(ro *repeatingOptionsType) {
		ro.indicatorStart = indicatorStart
		ro.indicatorEnd = indicatorEnd
	}
}

/*
Basic type for BigFloat number
*/
type BigFloat struct {
	analysis stranalyzer.Analysis
}

/*
Creates new BigFloat number with zero value
*/
func New() *BigFloat {
	return &BigFloat{
		analysis: stranalyzer.Analysis{
			Norm:     []byte{'0'},
			Sign:     1,
			Decimals: 0,
			Len:      1,
		},
	}
}

/*
Sets value of BigFloat number based on value and type of input parameter.

For specific type call SetInt64, SetString etc.
*/
func (f *BigFloat) Set(arg interface{}) (*BigFloat, error) {
	var err error
	switch value := arg.(type) {
	case string:
		err = f.SetString(value)
		return f, err
	case int:
		f.SetInt64(int64(value))
	case int64:
		f.SetInt64(value)
	case int8:
		f.SetInt64(int64(value))
	case int16:
		f.SetInt64(int64(value))
	case int32:
		f.SetInt64(int64(value))
	case *BigFloat:
		f.analysis = value.Copy().analysis
	case BigFloat:
		f.analysis = (&value).Copy().analysis
	default:
		panic("unknown argument")
	}
	return f, err
}

/*
Creates new BigFloat number according input parameter's type
*/
func Set(arg interface{}) (*BigFloat, error) {
	return New().Set(arg)
}

/*
Creates array of BigFloat number according variadic parameters
*/
func NewNumbers(args ...interface{}) ([]*BigFloat, []error) {
	result := make([]*BigFloat, len(args))
	errors := make([]error, len(args))

	for i, arg := range args {
		result[i], errors[i] = Set(arg)
	}

	return result, errors
}

type subProduct struct {
	offset  int
	product []int
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type correction struct {
	offset  int
	reverse int
}

type alignment struct {
	MaxIntLen   int
	MaxDecLen   int
	Len         int
	corrections []correction
}

/*
For internal use
Reads digit at pos position (from beginning), from f BigFloat, with c correction and fixing retrived digit by adding byteFix
*/
func read(pos int, f *BigFloat, c *correction, byteFix int) int {
	cPos := f.analysis.Len - pos - c.offset - 1 // calculate reading position

	if cPos >= 0 && cPos < f.analysis.Len {
		return int(f.analysis.Norm[cPos]) + byteFix
	}
	return 0
}

/*
For internal use
Reads digit at pos position (backwards from end), from f BigFloat, with c correction and fixing retrived digit by adding byteFix
*/
func reverse_read(pos int, f *BigFloat, c *correction, byteFix int) int {
	cPos := pos - c.reverse // calculate reading position

	if cPos >= 0 && cPos < f.analysis.Len {
		return int(f.analysis.Norm[cPos]) + byteFix
	}
	return 0
}

type readFnType = func(int, *BigFloat, *correction, int) int

/*
Reads one digit at pos position for every BigFloat in n array with a alignment
readFn can be read (from the beginning) or reverse_read (backwards from end)
*/
func multiRead(pos int, n []*BigFloat, a *alignment, byteFix int, readFn readFnType) []int {
	result := make([]int, len(n))
	for i := 0; i < len(n); i++ {
		result[i] = readFn(pos, n[i], &a.corrections[i], byteFix)
	}
	return result
}

/*
Calculate alignment two BigFloat numbers
*/
func align(args ...*BigFloat) alignment {
	var maxIntLen, maxDecLen int

	if len(args) > 0 {
		maxIntLen = args[0].analysis.Len - args[0].analysis.Decimals
		maxDecLen = args[0].analysis.Decimals
	}

	for i := 1; i < len(args); i++ {
		intLen := args[i].analysis.Len - args[i].analysis.Decimals
		decLen := args[i].analysis.Decimals

		if intLen > maxIntLen {
			maxIntLen = intLen
		}

		if decLen > maxDecLen {
			maxDecLen = decLen
		}
	}

	c := make([]correction, len(args))

	for i, arg := range args {
		c[i] = correction{
			offset:  arg.analysis.Decimals - maxDecLen,
			reverse: maxIntLen - (arg.analysis.Len - arg.analysis.Decimals),
		}
	}

	return alignment{
		MaxIntLen:   maxIntLen,
		MaxDecLen:   maxDecLen,
		Len:         maxIntLen + maxDecLen,
		corrections: c,
	}
}

/*
Parse string into BigFloat number
If parsing failed returns error
*/
func (f *BigFloat) SetString(s string) error {
	analysis, error := stranalyzer.Analyze(s)

	if error != nil {
		return error
	}

	f.analysis = analysis

	return nil
}

/*
Create new BigFloat number from string
If parsing failed returns error
*/
func SetString(s string) (*BigFloat, error) {
	f := &BigFloat{}
	err := f.SetString(s)

	return f, err
}

/*
Internal multiplication of two ascii bytes
*/
func mul(a, b byte) int {
	return int((a - 48) * (b - 48))
}

/*
Calculate integer division and remainder
*/
func divmod10(a int) (int, int) {
	d := a / 10
	m := a % 10

	return d, m
}

/*
Modifies []byte by adding value
*/
func add(a []byte, value int) []byte {
	for i := 0; i < len(a); i++ {
		a[i] = byte(int(a[i]) + value)
	}

	return a
}

/*
Creates new []byte vith initial value
*/
func fill(length int, value byte) []byte {
	digits := make([]byte, length)

	if value != 0 {
		for i := 0; i < len(digits); i++ {
			digits[i] = value
		}
	}

	return digits
}

/*
Reverses []byte
*/
func reverse(chars []byte) []byte {
	for i1, i2 := 0, len(chars)-1; i1 < i2; i1, i2 = i1+1, i2-1 {
		chars[i1], chars[i2] = chars[i2], chars[i1]
	}

	return chars
}

/*
Divides two BigFloat numbers with DivOption:

	decimalPlaces - target decimal places (default is -1 for detecting repeating decimals or remainder 0)
	maxDecimalPlaces - safety parameter for the result of division with a very large number of decimal places (default is 1e4 - 10000)
*/
func (f *BigFloat) Div(a, b *BigFloat, options ...DivOption) (*BigFloat, int, error) {
	r := &BigFloat{}

	return f.divmod(a, b, r, false, options...)
}

/*
Integer division of two BigFloat numbers
Returns integer division and modulus
*/
func (f *BigFloat) DivMod(a, b *BigFloat) (*BigFloat, *BigFloat, error) {
	r := &BigFloat{}
	_, _, err := f.divmod(a, b, r, true, WithDivDecimalPlaces(0))

	return f, r, err
}

/*
Internal divmod - used for Div and DIvMod
bTrunc argument determines if result should be truncated
for DivOption see Div method
*/
func (f *BigFloat) divmod(a, b, remainder *BigFloat, bTrunc bool, options ...DivOption) (*BigFloat, int, error) {
	ro := divOptionsType{ // default option values
		decimalPlaces:    -1,
		maxDecimalPlaces: int(1e4),
	}

	for _, option := range options { // process variadic arguments
		option(&ro)
	}

	if a.IsInt64(0) { // if 1st operand is 0 then result is 0
		f.SetInt64(0)
		if ro.decimalPlaces >= 0 {
			f.SetDecimals(ro.decimalPlaces)
		}
		return f, 0, nil
	} else if b.IsInt64(0) { // if 2nd operand is 0 then return error
		return nil, 0, fmt.Errorf("ERROR: Division by zero")
	}

	aCopy := a.Copy().Abs() // prepare copies of both operands with absolute values
	bCopy := b.Copy().Abs()

	if (a.analysis.Decimals + b.analysis.Decimals) > 0 { // eliminate decimals in both operands
		aCopy.Mul10(a.analysis.Decimals + b.analysis.Decimals)
		bCopy.Mul10(a.analysis.Decimals + b.analysis.Decimals)
	}

	if bCopy.IsInt64(1) { // if 2nd operand is 1 then result is 1st operand
		f.analysis = aCopy.analysis
		f.Sign(a.analysis.Sign * b.analysis.Sign) // sign of result

		if ro.decimalPlaces >= 0 {
			f.SetDecimals(ro.decimalPlaces)
		}

		return f, 0, nil
	}

	remainder.SetInt64(0)        // initial remainder (as in/out parameter)
	lastRemainder := SetInt64(0) // penultimate division remainder

	result := make([]byte, 0, aCopy.analysis.Len+bCopy.analysis.Len) // reserve enough space for result
	aBuf := aCopy.analysis.Norm                                      // running buffer

	maxLen := 5 // higher is better - used in converting part of string to int
	if maxLen > bCopy.analysis.Len {
		maxLen = bCopy.analysis.Len
	}
	cmpStr := bCopy.analysis.Norm[:maxLen]
	divPart2Int, _ := strconv.Atoi(string(cmpStr)) // preparing 2nd operand for division

	var divInt byte // int division

	initLen := bCopy.analysis.Len - 1 // initial length of 1st operand
	if initLen > len(aBuf) {
		initLen = len(aBuf)
	}
	divPart := aBuf[:initLen] // initial 1st operand - skip initial 0 integer division

	var digit byte                     // running digit for append to 1st operand
	bDecimals := false                 // idicator for calculation beyond decimal point
	remIndxMap := make(map[string]int) // map of modulus for detecting repeating decimals
	repDecimalsInd := -1               // index of modulus of repeating deimals

	decimals := 0                    // number of divisions decimals
	decimalsGoal := ro.decimalPlaces // target decimals
	if decimalsGoal >= 0 {           // increase target decimals for final rounding
		decimalsGoal++
	}

	for i := bCopy.analysis.Len - 1; true; i++ {
		if i >= len(aBuf) { // if calculates beyond decimal point
			digit = '0'
			if !bDecimals {
				bDecimals = true
			}
		} else {
			digit = aBuf[i]
		}

		if (bDecimals && string(divPart) == "0") || (decimalsGoal > 0 && decimals == decimalsGoal) { // exit loop if calculates beyond decimal point and there is no remainder or if target decimals reached
			break
		}

		if bDecimals { // calculates map for remainder and number of repeating decimals
			ind, exists := remIndxMap[string(divPart)]
			if exists { // exit loop if repeating decimal detected (target decimals calculates after loop exit)
				repDecimalsInd = ind
				break
			} else {
				remIndxMap[string(divPart)] = decimals
			}
		}
		if bDecimals {
			decimals++
		}

		*lastRemainder = BigFloat{ // save remainder in case of exit loop
			analysis: stranalyzer.Analysis{
				Norm:     divPart,
				Len:      len(divPart),
				Decimals: 0,
				Sign:     1,
			},
		}

		if string(divPart) == "0" { // fixes 0 remainder for large numbers
			divPart[0] = digit
		} else {
			divPart = append(divPart, digit)
		}

		if len(divPart) < bCopy.analysis.Len { // integer division is 0 for smaller 1st operand
			divInt = 0
		} else {
			divPart1Int, _ := strconv.Atoi(string(divPart[:len(divPart)-(bCopy.analysis.Len-len(cmpStr))])) // convert str into 1st operand
			divInt = byte(divPart1Int / divPart2Int)                                                        // guess division of two numbers
		}

		if divInt > 0 { // if division is greater then 0 multiply and substract to calculate remainder
			fProduct := bCopy.Copy()
			if divInt > 1 { // if division is greater then 1 then multiply
				fProduct.MulInt64(int64(divInt))
			}
			*remainder = BigFloat{ // create BigFloat for substraction
				analysis: stranalyzer.Analysis{
					Norm:     divPart,
					Len:      len(divPart),
					Decimals: 0,
					Sign:     1,
				},
			}

			if fProduct.Compare(remainder) == 1 { // fix wrong division
				divInt--                      // fix integer division
				fProduct.Sub(fProduct, bCopy) // fix product for substraction
			}

			remainder.Sub(remainder, fProduct) // calculate remainder
			divPart = remainder.analysis.Norm
		}

		if divInt > 0 || len(result) > 0 || bDecimals { // skip leading zeroes
			if bDecimals && len(result) == 0 { // append 0 before decimal point
				result = append(result, 0)
			}
			result = append(result, divInt)      // append devision result digit
			if decimals >= ro.maxDecimalPlaces { // safety loop exit in case of very long division and no repeating decimals detecting
				break
			}
		}
	}

	result = add(result, 48) // add 48 to every digit to get ascii numbers

	repeatDecimals := 0
	if repDecimalsInd >= 0 {
		repeatDecimals = decimals - repDecimalsInd // calculates number of repeating decimals (2nd return value)
	}

	if decimalsGoal >= 0 && decimalsGoal > decimals { // fix decimals to target decimals
		var repeatStr []byte
		if repeatDecimals > 0 { // with repeting decimals
			repeatStr = result[len(result)-repeatDecimals:]
			repeatDecimals = 0
		} else {
			repeatStr = []byte{'0'} // with zeroes
		}
		trailStr := bytes.Repeat(repeatStr, (decimalsGoal-decimals)/len(repeatStr)+1)
		trailStr = trailStr[:decimalsGoal-decimals]
		result = append(result, trailStr...)
		decimals = decimalsGoal
	}

	f.analysis = stranalyzer.Analysis{ // create division result
		Norm:     result,
		Len:      len(result),
		Decimals: decimals,
		Sign:     a.analysis.Sign * b.analysis.Sign, // sign of result
	}

	if ro.decimalPlaces >= 0 && ro.decimalPlaces < f.analysis.Decimals { // need to round result
		if bTrunc { // in case of integer division, decimals are truncated
			f.Trunc(WithDecimalPlaces(0))
		} else { // round penultimate digit
			f.Round(ro.decimalPlaces)
			f.SetDecimals(ro.decimalPlaces)
		}
		*remainder = *lastRemainder                                // prepare out arg as remainder
		remainder.Div10(a.analysis.Decimals + b.analysis.Decimals) // fix decimals in remainder
	}

	return f, repeatDecimals, nil
}

/*
Multiplication of two BigFLoat numbers
*/
func (f *BigFloat) Mul(a, b *BigFloat) *BigFloat {
	if a.IsInt64(0) || b.IsInt64(0) { // check for 0
		newDecimals := maxInt(a.analysis.Decimals, b.analysis.Decimals)

		return f.SetInt64(0).SetDecimals(newDecimals) // result is 0
	} else if a.IsInt64(1) { // if 1st operand is 1 then result is 2nd operand
		f.analysis = b.analysis

		return f
	} else if a.IsInt64(-1) { // if 1st operand is -1 then result is negative 2nd operand
		f.analysis = b.analysis

		return f.Neg()
	} else if b.IsInt64(1) { // if 2nd operand is 1 then result is 1st operand
		f.analysis = a.analysis

		return f
	} else if b.IsInt64(-1) { // if 2nd operand is -1 then result is negative 1st operand
		f.analysis = a.analysis

		return f.Neg()
	}

	var r, overflow int        // running multiplication result
	var resultBuf []subProduct // sub products with offsets
	aStr := a.analysis.Norm
	bStr := b.analysis.Norm

	for bd := 0; bd < len(bStr); bd++ { // long division algorithm
		if bStr[bd] == '0' {
			continue
		}

		overflow = 0

		sp := subProduct{
			offset: b.analysis.Len - bd - 1, // sets offest for sub product addition
		}

		cBuf := make([]int, 0, len(aStr)+1)

		for ad := len(aStr) - 1; ad >= 0; ad-- { // multiply every digit
			total := mul(bStr[bd], aStr[ad])
			total += overflow
			overflow, r = divmod10(total) // calculate overflow
			cBuf = append(cBuf, r)        // append result digit
		}

		if overflow != 0 { // if exists add overflow after loop
			cBuf = append(cBuf, overflow)
		}

		sp.product = cBuf // prepare sub products
		resultBuf = append(resultBuf, sp)
	}

	var ppos, s int
	overflow = 0
	lenP := a.analysis.Len + b.analysis.Len
	totalBuf := make([]byte, 0, lenP)

	for pos := 0; pos < lenP; pos++ { // calculate sum of sub products
		s = overflow

		for r = 0; r < len(resultBuf); r++ {

			ppos = pos - resultBuf[r].offset
			if ppos >= 0 && ppos < len(resultBuf[r].product) {
				s += resultBuf[r].product[ppos]
			}

		}

		overflow, r = divmod10(s)            // calculate overflow
		totalBuf = append(totalBuf, byte(r)) // append sum digit
	}
	totalBuf = append(totalBuf, byte(overflow)) // add overflow after loop

	newDecimals := a.analysis.Decimals + b.analysis.Decimals // calculate number of decimals
	totalBuf = reverse(totalBuf)                             // reverse digits for display (low digits to the right)
	if len(totalBuf) > newDecimals {                         // trim leading zeroes
		iTrim := 0
		for i := 0; i < len(totalBuf)-newDecimals-1; i++ { // except first digit before decimal point
			if totalBuf[i] == 0 {
				iTrim++
			} else {
				break
			}
		}
		totalBuf = totalBuf[iTrim:]
	}

	if newDecimals > 0 { // trim trailing zeroes
		iTrim := 0
		for i := 0; i < newDecimals; i++ {
			if totalBuf[len(totalBuf)-i-1] == 0 {
				iTrim++
			} else {
				break
			}
		}
		totalBuf = totalBuf[:len(totalBuf)-iTrim]
		newDecimals -= iTrim
	}

	totalBuf = add(totalBuf, 48) // add 48 to every digit to get ascii numbers

	f.analysis = stranalyzer.Analysis{ // prepare result
		Norm:     totalBuf,
		Len:      len(totalBuf),
		Decimals: newDecimals,
		Sign:     a.analysis.Sign * b.analysis.Sign,
	}

	return f
}

/*
Returns if BigFloat equals number n
*/
func (f *BigFloat) IsInt64(n int64) bool {
	nFloat := BigFloat{}
	nFloat.SetInt64(n).SetDecimals(f.analysis.Decimals)

	return (f.analysis.Sign == nFloat.analysis.Sign) && // checks if number are equals - cannot compare with f.analysis.Sign == nFloat.analysis as in analysis is []byte
		(f.analysis.Decimals == nFloat.analysis.Decimals) &&
		(f.analysis.Len == nFloat.analysis.Len) &&
		(string(f.analysis.Norm) == string(nFloat.analysis.Norm))
}

/*
Truncates integer part of number and returns decimals only
*/
func (f *BigFloat) Frac() *BigFloat {
	if (f.analysis.Len - f.analysis.Decimals) > 0 {
		f.analysis.Norm = append([]byte{'0'}, f.analysis.Norm[f.analysis.Len-f.analysis.Decimals:]...)
		f.analysis.Len = len(f.analysis.Norm)
		for i := 0; i < f.analysis.Len; i++ {
			if f.analysis.Norm[i] != '0' {
				return f
			}
		}
		f.analysis.Sign = 1
	}

	return f
}

/*
Multiply BigFloat with int64
*/
func (f *BigFloat) MulInt64(n int64) *BigFloat {
	if n == 0 { // result is 0 with a predefined number of decimals
		f.analysis.Sign = 1
		f.analysis.Norm = fill(f.analysis.Decimals+1, '0')
		f.analysis.Len = f.analysis.Decimals + 1

		return f
	} else if n == 1 { // same BigFloat as result

		return f
	} else if n == -1 { // for -1 result is opposite sign except for 0
		if !f.IsInt64(0) {
			f.analysis.Sign *= -1
		}

		return f
	}

	nStr := strconv.FormatInt(n, 10) // check if n is a multiple of 10
	trimmedString := strings.TrimRight(nStr, "0")
	numZeroes := len(nStr) - len(trimmedString)
	if (numZeroes > 0) && ((trimmedString == "1") || (trimmedString == "-1")) { // if n is a multiple of 10
		if trimmedString == "-1" { // calculate sign
			f.analysis.Sign *= -1
		}

		return f.Mul10(numZeroes) // simple move decimals
	}

	nFloat := BigFloat{} // in every other case multiply two BigFloat numbers
	nFloat.SetInt64(n)

	return f.Mul(f, &nFloat)
}

/*
Truncate decimals in BigFloat number
RoundOption defines number of decimals in result
*/
func (f *BigFloat) Trunc(options ...RoundOption) *BigFloat {
	ro := roundOptionsType{
		decimalPlaces: f.analysis.Decimals,
	}
	for _, option := range options {
		option(&ro)
	}

	if ro.decimalPlaces < 0 {
		panic("ERROR: Negative decimal places. Decimal places should be 0 or positive")
	}

	for i := f.analysis.Len - f.analysis.Decimals; i < f.analysis.Len; i++ { // set '0' as decimals digits
		f.analysis.Norm[i] = '0'
	}

	f.SetDecimals(ro.decimalPlaces)

	return f
}

/*
Set target decimals
*/
func (f *BigFloat) SetDecimals(n int) *BigFloat {

	if n == f.analysis.Decimals { // no need to change
		f.Sign(f.analysis.Sign) // checks negative sign for 0 e.g. (-0.1).SetDecimals(0) => 0
		return f
	} else if n > f.analysis.Decimals { // need to add zeroes
		zeroes := fill(n-f.analysis.Decimals, '0')
		f.analysis.Norm = append(f.analysis.Norm, zeroes...)
	} else { // need to trim trailing decimals
		f.analysis.Norm = f.analysis.Norm[:f.analysis.Len-f.analysis.Decimals+n]
	}

	f.analysis.Decimals = n
	f.analysis.Len = len(f.analysis.Norm)
	f.Sign(f.analysis.Sign) // checks negative sign for 0 e.g. (-0.1).SetDecimals(0) => 0

	return f
}

/*
Adds two BigFloat numbers
Warning: Only for numbers with the same sign
*/
func (f *BigFloat) add(a, b *BigFloat) *BigFloat {
	n := []*BigFloat{a, b}   // array of operands
	alignment := align(n...) // calculate alignemnt with decimals

	var r int
	overflow := 0
	sum := 0
	totalBuf := make([]byte, 0, alignment.Len+1) // result of addition

	for p := 0; p < alignment.Len; p++ {
		sum = overflow
		v := multiRead(p, n, &alignment, -48, read) // read operands digits
		sum += v[0] + v[1]                          // calculate sum
		overflow, r = divmod10(sum)                 // calculate result digit and overflow
		totalBuf = append(totalBuf, byte(r))        // append result digit
	}
	totalBuf = append(totalBuf, byte(overflow)) // add overflow after loop

	totalBuf = reverse(totalBuf)
	if len(totalBuf) > alignment.MaxDecLen { // trim leading zeroes
		iTrim := 0
		for i := 0; i < len(totalBuf)-alignment.MaxDecLen-1; i++ { // except first digit before decimal point
			if totalBuf[i] == 0 {
				iTrim++
			} else {
				break
			}
		}
		totalBuf = totalBuf[iTrim:]
	}

	totalBuf = add(totalBuf, 48) // add 48 to every digit to get ascii numbers

	f.analysis = stranalyzer.Analysis{ // prepare result
		Norm:     totalBuf,
		Len:      len(totalBuf),
		Decimals: alignment.MaxDecLen,
		Sign:     a.analysis.Sign, // same signe of both operands
	}

	return f
}

/*
Adds two BigFloat numbers

See following table for cases when operands are swapped and how result sign is set

	Addition:
	| a  |  b  |  a + b | swap |  sign of result  | abs(a) +- abs(b) |
	|---:|----:|:------:|:----:|:----------------:|:----------------:|
	| -5 |  -8 | -(5+8) |  no  |     abs bigger   |          5+8     |
	| -8 |  -5 | -(8+5) |  no  |     abs bigger   |          8+5     |
	|  5 |  -8 | -(8-5) |  yes |     abs bigger   |          8-5     |
	| -8 |   5 | -(8-5) |  no  |     abs bigger   |          8-5     |
	| -5 |   8 |   8-5  |  yes |     abs bigger   |          8-5     |
	|  8 |  -5 |   8-5  |  no  |     abs bigger   |          8-5     |
	|  5 |   8 |   5+8  |  no  |     abs bigger   |          5+8     |
	|  8 |   5 |   8+5  |  no  |     abs bigger   |          8+5     |
*/
func (f *BigFloat) Add(a, b *BigFloat) *BigFloat {
	if a.IsInt64(0) { // if 1st operand is 0 then result is 2nd operand (0 + B = B)
		f.analysis = b.analysis

		return f
	} else if b.IsInt64(0) { // if 2nd operand is 0 then result is 1st operand (A + 0 = A)
		f.analysis = a.analysis

		return f
	}

	if a.analysis.Sign == b.analysis.Sign { // if both operands have same signs call internal add
		return f.add(a, b)
	} else { // opposite signs
		f1, f2 := a.Copy(), b.Copy()
		sign := a.analysis.Sign

		cmp := a.CompareAbs(b)
		if cmp == 0 { // if numbers are opposite then result is 0 (A + -A = 0)
			newDecimals := int(math.Max(float64(a.analysis.Decimals), float64(b.analysis.Decimals)))

			return f.SetInt64(0).SetDecimals(newDecimals)
		} else if cmp < 0 { // if 1st operand is smaller then 2nd then swap operands
			f1, f2 = f2, f1
			sign = b.analysis.Sign
		}
		f1.Abs() // ignore signs
		f2.Abs()
		f.sub(f1, f2) // substract smaller operand from bigger operand
		f.Sign(sign)  // set result sign

		return f
	}
}

/*
Internal method for substraction
*/
func (f *BigFloat) sub(a, b *BigFloat) (*BigFloat, error) {
	n := []*BigFloat{a, b}   // array of operands
	alignment := align(n...) // calculate alignemnt with decimals

	var diff, overflow int                             // running difference as result and overflow
	totalBuf := make([]byte, 0, alignment.MaxIntLen+1) // buffer for result digits

	for i := 0; i < alignment.Len; i++ { // for all digits
		v := multiRead(i, n, &alignment, -48, read) // read aligned digits as i position

		v[0] = v[0] - overflow // substract overflow from previous substraction
		if v[0] < v[1] {       // negative difference so calculate for overflow
			overflow = 1
			v[0] += 10
		} else {
			overflow = 0
		}

		diff = int(v[0]) - int(v[1])            // calculate difference digit
		totalBuf = append(totalBuf, byte(diff)) // append result
	}
	totalBuf = reverse(totalBuf) // add overflow digit

	if len(totalBuf) > alignment.MaxDecLen { // trim leading zeroes
		iTrim := 0
		for i := 0; i < len(totalBuf)-alignment.MaxDecLen-1; i++ {
			if totalBuf[i] == 0 {
				iTrim++
			} else {
				break
			}
		}
		totalBuf = totalBuf[iTrim:]
	}

	add(totalBuf, 48) // add 48 to every digit to get ascii numbers

	f.analysis = stranalyzer.Analysis{ // prepare result
		Norm:     totalBuf,
		Len:      len(totalBuf),
		Decimals: alignment.MaxDecLen,
		Sign:     1, // both operands have sign 1
	}

	return f, nil
}

/*
Substracts two BigFloat numbers

See following table for cases when operands are swapped and how result sign is set

	Subtraction:
	| a  |   b |  a - b | swap |  sign of result  | abs(a) +- abs(b) |
	|---:|----:|:------:|:----:|:----------------:|:----------------:|
	| -5 |  -8 |   8-5  |  yes |      neg 2nd     |       8-5        |
	| -8 |  -5 | -(8-5) |  no  |        1st       |       8-5        |
	|  5 |  -8 |   5+8  |  no  |        1st       |       5+8        |
	| -8 |   5 | -(8+5) |  no  |        1st       |       8+5        |
	| -5 |   8 | -(5+8) |  no  |        1st       |       5+8        |
	|  8 |  -5 |   8+5  |  no  |        1st       |       8+5        |
	|  5 |   8 | -(8-5) |  yes |      neg 2nd     |       8-5        |
	|  8 |   5 |   8-5  |  no  |        1st       |       8-5        |
*/
func (f *BigFloat) Sub(a, b *BigFloat) *BigFloat {
	if a.IsInt64(0) { // if 1st operand is 0 then result is opposite 2nd operand (0 - B = -B)
		f.analysis = b.analysis

		return f.Neg()
	} else if b.IsInt64(0) { // if 2nd operand is 0 then result is 1st operand (A - 0 = A)
		f.analysis = a.analysis

		return f
	}
	f1, f2 := a.Copy(), b.Copy()
	if a.analysis.Sign != b.analysis.Sign { // if operands have opposite signs
		sign := f1.analysis.Sign
		f1.Abs()
		f2.Abs()
		f.add(f1, f2)
		f.Sign(sign)

		return f
	} else { // if operands have same signs
		cmp := a.CompareAbs(b)

		if cmp == 0 { // opposite numbers (A - A = 0 or -A - -A = 0)
			newDecimals := maxInt(a.analysis.Decimals, b.analysis.Decimals)
			return f.SetInt64(0).SetDecimals(newDecimals)
		}

		sign := a.analysis.Sign
		if cmp < 0 { // swap operands when 1st operand is smaller then 2nd operand
			f1, f2 = f2, f1
			sign = b.analysis.Sign * -1
		}
		f1.Abs()
		f2.Abs()
		f.sub(f1, f2)
		f.Sign(sign)

		return f
	}
}

/*
Copy BigFloat number
*/
func (f *BigFloat) Copy() *BigFloat {
	return &BigFloat{f.analysis}
}

/*
Retrieve sign of BigFloat number
For negative numbers sign is -1, for 0 or positive numbers sign is 1

Warning: There is no sign == 0
*/
func (f *BigFloat) GetSign() int {
	return f.analysis.Sign
}

/*
Set sign of BigFloat number

Warning: There is no sign == 0
*/
func (f *BigFloat) Sign(s int) *BigFloat {
	if s == -1 {
		f.analysis.Sign = 1
		if f.IsInt64(0) { // 0 has sign == 0
			s = 1
		}
	}
	f.analysis.Sign = s

	return f
}

/*
Sets opposite sign, except for 0
*/
func (f *BigFloat) Neg() *BigFloat {
	f.Sign(f.analysis.Sign * -1)

	return f
}

/*
Returns aboslute value of BigFloat number
*/
func (f *BigFloat) Abs() *BigFloat {
	return f.Sign(1)
}

/*
Compares two BigFloat numbers and returns:
-1 if 1st number is smaller then 2nd
0 if 1st number is equal to 2nd
1 if 1st number is bigger then 2nd
*/
func (f *BigFloat) Compare(a *BigFloat) int {
	return f.compare(a, false)
}

/*
Compares absolute values of two BigFloat numbers and returns:
-1 if 1st number is smaller then 2nd
0 if 1st number is equal to 2nd
1 if 1st number is bigger then 2nd
*/
func (f *BigFloat) CompareAbs(a *BigFloat) int {
	return f.compare(a, true)
}

/*
Internal method for comparing two BigFloat numbers
*/
func (f *BigFloat) compare(a *BigFloat, abs bool) int {
	if abs || (f.analysis.Sign == a.analysis.Sign) { // if signs are same or signs are ignored in case of abs == true
		n := []*BigFloat{f, a}
		alignment := align(n...) // calculate decimals aligment

		var sign int
		if abs {
			sign = 1
		} else {
			sign = f.analysis.Sign
		}

		for i := 0; i < alignment.Len; i++ { // compare every digit
			v := multiRead(i, n, &alignment, -48, reverse_read) // reads from beginning
			if v[0] < v[1] {
				return -sign
			} else if v[0] > v[1] {
				return sign
			}
		}
		return 0
	}

	return f.analysis.Sign
}

/*
Rounds number to n decimals
*/
func (f *BigFloat) Round(n int) *BigFloat {
	if n < 0 {
		panic("Invalid decimal number")
	}
	if n < f.analysis.Decimals { // if n decimal for rounding exists
		pos := f.analysis.Len - f.analysis.Decimals + n // posistion of digit for rounding
		d := f.analysis.Norm[pos]                       // digit for rounding
		f.analysis.Len -= f.analysis.Decimals - n       // fix the length
		f.analysis.Decimals = n                         // fix decimals
		//digits := fill(f.analysis.Len-pos, 48)          // zeroes after rounding digit
		f.analysis.Norm = f.analysis.Norm[:pos]

		if d >= '5' { // rounding up
			c := BigFloat{}        // create new BigFloat number
			c.SetInt64(1).Div10(n) // with rounding digit
			c.Sign(f.GetSign())
			f.Add(f, &c) // calculate new number with addition
		}

		f.Sign(f.analysis.Sign)
	}

	return f
}

/*
Set power of 10 to BigFloat number
*/
func (f *BigFloat) Pow10(n int) (*BigFloat, error) {
	if n < 0 { // support for negative Pow10
		f.SetInt64(1).Div10(-n)
		return f, nil
	}

	zeroes := fill(n+1, 48) // '1' + n zeroes
	zeroes[0] += 1
	f.analysis = stranalyzer.Analysis{
		Norm:     zeroes,
		Len:      len(zeroes),
		Decimals: 0,
		Sign:     1,
	}

	return f, nil
}

/*
Multiply BigFloat number with 10 multiplicator
*/
func (f *BigFloat) Mul10(n int) *BigFloat {
	if (f.analysis.Decimals - n) > 0 { // if there is enough decimals for multiplication
		f.analysis.Decimals -= n // just fix decimals
	} else {
		zeroes := fill(n-f.analysis.Decimals, '0')           // prepare yeroes
		f.analysis.Norm = append(f.analysis.Norm, zeroes...) // prepend zeroes
		f.analysis.Len += n - f.analysis.Decimals
		f.analysis.Decimals = 0

		// trim leading zeroes
		iTrim := 0
		for i := 0; i < f.analysis.Len-f.analysis.Decimals; i++ {
			if f.analysis.Norm[i] == 48 {
				iTrim++
			} else {
				break
			}
		}
		if iTrim > 0 {
			f.analysis.Norm = f.analysis.Norm[iTrim:]
			f.analysis.Len -= iTrim
		}

	}

	return f
}

/*
Divides BigFloat number with 10 multiplicator
*/
func (f *BigFloat) Div10(n int) *BigFloat {
	if (f.analysis.Decimals + n) >= f.analysis.Len { // need to prepend zeroes
		zeroes := fill(n-(f.analysis.Len-f.analysis.Decimals)+1, '0')
		f.analysis.Norm = append(zeroes, f.analysis.Norm...)
		f.analysis.Len += len(zeroes)
	}
	f.analysis.Decimals += n

	return f
}

/*
Sets int64 number to existing BigFloat number
*/
func (f *BigFloat) SetInt64(n int64) *BigFloat {
	s := strconv.FormatInt(n, 10)
	f.SetString(s)

	return f
}

/*
Creates new BigFloat number from int64 number
*/
func SetInt64(n int64) *BigFloat {
	f := &BigFloat{}
	return f.SetInt64(n)
}

/*
Creates new BigFloat number from int number
*/
func SetInt(n int) *BigFloat {
	f := &BigFloat{}
	return f.SetInt64(int64(n))
}

/*
Function type for formatting with StringF

See: StringF
*/
type StringOption func(*stringOptionType)

type stringOptionType struct {
	forceSign bool
}

/*
Function defines if sign is forced in formatting or not.

See: StringF
*/
func ForceSign(forceSign bool) StringOption {
	return func(so *stringOptionType) {
		so.forceSign = forceSign
	}
}

/*
Returns string
Optional arg is forceSign for 0 or positive number to force '+' sign
*/
func (f *BigFloat) String() string {
	return f.StringWith(ForceSign(false))
}

/*
Returns string with repeating decimals and/or StringOption and/or RepeatingOptions
*/
func (f *BigFloat) StringF(RepeatingDecimals int, options ...interface{}) string {
	ro := repeatingOptionsType{
		indicatorStart: "(",
		indicatorEnd:   ")",
	}

	strOptions := make([]StringOption, 0)
	for _, option := range options {
		if reflect.TypeOf(option).String() == "bigfloat.RepeatingOptions" {
			option.(RepeatingOptions)(&ro)
		} else if reflect.TypeOf(option).String() == "bigfloat.StringOption" {
			strOptions = append(strOptions, option.(StringOption))
		} else {
			panic("wrong input type parameter")
		}
	}

	result := f.StringWith(strOptions...)

	if RepeatingDecimals > 0 {
		var b strings.Builder
		b.Grow(len(result) + len(ro.indicatorStart) + len(ro.indicatorEnd))

		fmt.Fprintf(&b, "%s%s%s%s", result[:len(result)-RepeatingDecimals], ro.indicatorStart, result[len(result)-RepeatingDecimals:], ro.indicatorEnd)

		return b.String()
	}

	return result
}

/*
Returns string with formatting options:
-forceSign bool - if true then forces '+' sign for positive numbers
*/
func (f *BigFloat) StringWith(options ...StringOption) string {
	so := stringOptionType{
		forceSign: false,
	}
	for _, option := range options {
		option(&so)
	}

	var b strings.Builder
	b.Grow(f.analysis.Len + 2)

	if f.analysis.Sign == -1 {
		fmt.Fprintf(&b, "%c", '-')
	} else if so.forceSign && !f.IsInt64(0) {
		fmt.Fprintf(&b, "%c", '+')
	}

	fmt.Fprintf(&b, "%s", f.analysis.Norm[:f.analysis.Len-f.analysis.Decimals])

	if f.analysis.Decimals > 0 {
		fmt.Fprintf(&b, ".%s", f.analysis.Norm[f.analysis.Len-f.analysis.Decimals:])
	}

	return b.String()
}
