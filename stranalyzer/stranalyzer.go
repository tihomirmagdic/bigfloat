package stranalyzer

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
)

type Analysis struct {
	Norm     []byte
	Sign     int
	Decimals int
	Len      int
}

func visible(c rune) bool {
	return unicode.IsGraphic(c)
}

func Analyze(s string) (a Analysis, e error) {
	r := []rune(s)
	signFound := false
	a.Sign = 1
	decimalPointFound := false
	eFound := false
	eValue := ""
	eSign := 1
	eSignFound := false
	digitFound := false
	nonZeroDigitFound := false
	normBuf := make([]byte, 0, len(s))
	eBuf := make([]byte, 0)
	for i := 0; i < len(r); i++ {
		if !visible(r[i]) {
			continue
		}
		if (r[i] == '+') || (r[i]) == '-' {
			if signFound && !eFound {
				return a, fmt.Errorf("ERROR: Sign already found before. New sign at pos %d", i)
			}
			if eSignFound && eFound {
				return a, fmt.Errorf("ERROR: E sign already found before. New E sign at pos %d", i)
			}
			if digitFound && !eFound {
				return a, fmt.Errorf("ERROR: Sign found after digit at pos %d", i)
			}
			if decimalPointFound && !eFound {
				return a, fmt.Errorf("ERROR: Sign found after decimal point and before E number at pos %d", i)
			}
			if eFound {
				eSignFound = true
				if r[i] == '-' {
					eSign = -1
				}
			} else {
				signFound = true
				if r[i] == '-' {
					a.Sign = -1
				}
			}
		} else if r[i] == '.' {
			if decimalPointFound {
				return a, fmt.Errorf("ERROR: Decimal point already found before. New decimal point at pos %d", i)
			}
			if eFound {
				return a, fmt.Errorf("ERROR: Decimal point not allowed in E number at pos %d", i)
			}
			decimalPointFound = true
			if !nonZeroDigitFound {
				normBuf = append(normBuf, '0')
				a.Len++
			}
		} else if (r[i] == 'E') || (r[i] == 'e') {
			if eFound {
				return a, fmt.Errorf("ERROR: 'E' already found before. New 'E' at pos %d", i)
			}
			if !nonZeroDigitFound {
				return a, fmt.Errorf("ERROR: Missing number before E number")
			}
			eFound = true
		} else if unicode.IsDigit(r[i]) {
			if !eFound {
				if decimalPointFound || (r[i] != '0') || (digitFound && nonZeroDigitFound) { // ignore leading zeroes
					//if decimalPointFound || !digitFound || (digitFound && nonZeroDigitFound) { // ignore leading zeroes
					normBuf = append(normBuf, byte(r[i]))
					a.Len++
					if decimalPointFound {
						a.Decimals++
					}
					if r[i] != '0' {
						nonZeroDigitFound = true
					}
				}
				digitFound = true
			} else {
				//eValue += string(r[i])
				eBuf = append(eBuf, byte(r[i]))
				digitFound = true
			}
		} else if r[i] == ' ' {
			continue
		} else {
			return a, fmt.Errorf("ERROR: invalid big float number")
		}
	}
	eValue = string(eBuf)
	if eFound && eValue == "" {
		return a, fmt.Errorf("ERROR: invalid E number")
	}
	if !digitFound {
		return a, fmt.Errorf("ERROR: invalid number")
	}
	if digitFound && (a.Len == 0) {
		normBuf = append(normBuf, '0')
		a.Len++
	}
	//a.Norm = string(normBuf)
	a.Norm = normBuf
	if !nonZeroDigitFound && (a.Sign != 1) {
		a.Sign = 1
	}
	if eFound {
		eIntParsed, _ := strconv.ParseInt(eValue, 10, 0)
		eInt := int(eIntParsed)
		if eSign == 1 {
			a.Decimals -= eInt
			if a.Decimals < 0 {
				eInt = -a.Decimals
				a.Decimals = 0
			} else {
				eInt -= a.Decimals
			}
			//a.Norm += strings.Repeat("0", int(eInt))
			//append([]byte("0"), int(eInt))
			a.Norm = append(a.Norm, bytes.Repeat([]byte("0"), int(eInt))...)
			a.Len += eInt
		} else {
			a.Decimals += eInt
			if a.Decimals >= a.Len {
				eInt = a.Decimals - a.Len + 1
				//a.Norm = strings.Repeat("0", int(eInt)) + a.Norm
				a.Norm = append(bytes.Repeat([]byte("0"), int(eInt)), a.Norm...)
				a.Len += eInt
			}
		}
	}
	return a, nil
}
