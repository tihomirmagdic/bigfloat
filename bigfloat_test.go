package bigfloat_test

import (
	"bigfloat"
	"fmt"
	"strconv"
	"testing"
)

func toString(f int64) string {
	return strconv.FormatInt(f, 10)
}

func createBigFloat(t *testing.T, s string) (*bigfloat.BigFloat, error) {
	n := &bigfloat.BigFloat{}
	err := n.SetString(s)
	if err != nil && t != nil {
		t.Errorf("ERROR: %q is not valid big float number\n", s)
	}
	return n, err
}

func create2BigFloats(t *testing.T, s1, s2 string) (*bigfloat.BigFloat, *bigfloat.BigFloat, error) {
	n1, err := createBigFloat(t, s1)
	if err == nil {
		n2, err := createBigFloat(t, s2)
		return n1, n2, err
	}
	return nil, nil, err
}

func printResult(t *testing.T, result, expectedStr string, err error) {
	if (result != expectedStr) || (err != nil) {
		errorStr := fmt.Sprintf("should be %v", expectedStr)
		if err != nil {
			errorStr += fmt.Sprintf("(error: %v)", err)
		}
		fmt.Printf(errorStr + "\n")
		t.Errorf(errorStr)
	}
}

func TestDiv(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   string
		decimals int
		expected string
	}{
		{"2", "0.002", -1, "1000"},
		{"2", "-0.002", -1, "-1000"},
		{"-2", "0.002", -1, "-1000"},
		{"-2", "-0.002", -1, "1000"},
		{"-0.01", "2", -1, "-0.005"},
		{"-0.01", "3", -1, "-0.00(3)"},
		{"0.01", "-3", -1, "-0.00(3)"},
		{"0.01", "-3", 10, "-0.0033333333"},
		{"0.01", "0.01", -1, "1"},
		{"-1", "-1", 0, "1"},
		{"-1", "1", 0, "-1"},
		{"1", "-1", 0, "-1"},
		{"1", "-1", 2, "-1.00"},
		{"0", "1", 0, "0"},
		{"0", "-1", 0, "0"},
		{"0", "0.01", 0, "0"},
		{"0", "-0.01", 0, "0"},
		{"100", "-0.7", -1, "-142.(857142)"},
		{"100", "-0.7", 5, "-142.85714"},
		{"100", "-0.7", 0, "-143"},
		{"-1", "20000", 5, "-0.00005"},
		{"1", "12345", 5, "0.00008"},
		{"12345", "1", -1, "12345"},
		{"-12345", "1", -1, "-12345"},
		{"-12345", "-1", -1, "12345"},
		{"12345", "-1", -1, "-12345"},
		{"0.01", "0.01", -1, "1"},
		{"1", "22", -1, "0.0(45)"},
		{"1", "22", 5, "0.04545"},
		{"1", "22", 4, "0.0455"},
		{"1", "6", -1, "0.1(6)"},
		{"100", "-0.7", -1, "-142.(857142)"},
		{"100", "-0.7", 5, "-142.85714"},
		{"100", "-0.7", 15, "-142.857142857142857"},
		{"1", "9", -1, "0.(1)"},
		{"1", "12", -1, "0.08(3)"},
		{"1", "11", -1, "0.(09)"},
		{"1", "3", -1, "0.(3)"},
		{"1", "70", -1, "0.0(142857)"},
		{"-1", "70", -1, "-0.0(142857)"},
		{"17253428", "32459", 13, "531.5452724976124"},
		{"30", "15.0000001", 53, "1.99999998666666675555555496296296691358022057613186283"},
		{"30", "15.0000001", -1, "1.99999998666666675555555496296296691358022057613186282578758116141612559055916272960558180262945464"},
	}
	fmt.Printf("\nTestDiv...\n")
	for _, c := range cases {
		fmt.Printf("div(%v, %v, %v) = ", c.param1, c.param2, c.decimals)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}

		var result string
		expectedStr := c.expected

		n3 := &bigfloat.BigFloat{}
		_, repDec, errDiv := n3.Div(n1, n2, bigfloat.WithDivDecimalPlaces(c.decimals), bigfloat.WithDivMaxDecimalPlaces(int(1e2)))
		if errDiv != nil {
			fmt.Printf("%v\n", errDiv)
			t.Errorf("Division error %v", errDiv)
			continue
		}
		result = bigfloat.StringWithRepeatingDecimals(n3, repDec)
		//result = bigfloat.StringWithRepeatingDecimals(n3, repDec, bigfloat.ForceSign(false), bigfloat.WithRepeatingOptions("(", ")"))
		/* 		if repDec > 0 {
		   		} else {
		   			result = n3.String()
		   		}
		*/
		fmt.Printf("%v\n", result)
		if len(result) > 100 {
			fmt.Printf("result len: %v\n", len(result))
			result = result[:100]
		}
		printResult(t, result, expectedStr, errDiv)
	}
}

func TestRepeatingDecimals(t *testing.T) {
	var cases = []struct {
		param1         string
		param2         string
		startIndicator string
		endIndicator   string
		expected       string
	}{
		{"-1", "70", "(", ")", "-0.0(142857)"},
		{"-1", "70", "R", "", "-0.0R142857"},
		{"-1", "70", "r", "", "-0.0r142857"},
		{"-1", "70", "#", "$", "-0.0#142857$"},
		{"1", "70", "r", "", "+0.0r142857"},
	}
	fmt.Printf("\nTestDiv...\n")
	for _, c := range cases {
		fmt.Printf("div(%v, %v) = ", c.param1, c.param2)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}

		var result string
		expectedStr := c.expected

		n3 := &bigfloat.BigFloat{}
		_, repDec, errDiv := n3.Div(n1, n2) //, bigfloat.WithDecimalPlaces(-1))
		if errDiv != nil {
			fmt.Printf("%v\n", errDiv)
			t.Errorf("Division error %v", errDiv)
			continue
		}
		result = bigfloat.StringWithRepeatingDecimals(n3, repDec, bigfloat.ForceSign(true), bigfloat.WithRepeatingOptions(c.startIndicator, c.endIndicator))
		//fmt.Printf("%v\n", n3.String())

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, errDiv)
	}
}

func TestDivMod(t *testing.T) {
	var cases = []struct {
		param1    string
		param2    string
		expected1 string
		expected2 string
	}{
		{"1", "-3", "0", "1"},
		{"10", "7", "1", "3"},
		{"18", "7", "2", "4"},
		{"-18", "7", "-2", "4"},
		{"-18", "7.2", "-2", "3.6"},
		{"-1", "8", "0", "1"},
		{"-1", "20000", "0", "1"},
		{"2", "12345", "0", "2"},
		{"1", "-3", "0", "1"},
		{"43", "22", "1", "21"},
		{"43", "-22", "-1", "21"},
		{"-43", "-22", "1", "21"},
		{"-43", "22", "-1", "21"},
	}
	fmt.Printf("\nTestDivMod...\n")
	for _, c := range cases {
		fmt.Printf("divMod(%v, %v) = ", c.param1, c.param2)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}

		var result string

		n3 := &bigfloat.BigFloat{}
		_, remainder, errDiv := n3.DivMod(n1, n2)
		if errDiv != nil {
			fmt.Printf("%v\n", errDiv)
			t.Errorf("Division error %v", errDiv)
			continue
		}
		result = n3.String()
		//fmt.Printf("%v (%v)\n", n3.String(), remainder.String())
		result = fmt.Sprintf("%v (%v)", result, remainder.String())
		fmt.Printf("%v\n", result)

		printResult(t, result, fmt.Sprintf("%v (%v)", c.expected1, c.expected2), errDiv)
		//printResult(t, remainder.String(), c.expected2, errDiv)
	}
}

var cases = []struct {
	param1      string
	param2      string
	expectedAdd string
	expectedSub string
	expectedMul string
}{
	{"0", "1", "1", "-1", "0"},
	{"0", "0", "0", "0", "0"},
	{"1", "0", "1", "1", "0"},
	{"0", "1", "1", "-1", "0"},
	{"1", "1", "2", "0", "1"},
	{"-1", "0", "-1", "-1", "0"},
	{"0", "-1", "-1", "1", "0"},
	{"1", "-1", "0", "2", "-1"},
	{"-1", "1", "0", "-2", "-1"},
	{"-2", "1", "-1", "-3", "-2"},
	{"2", "-1", "1", "3", "-2"},
	{"-1", "2", "1", "-3", "-2"},
	{"1", "-2", "-1", "3", "-2"},
	{"-2", "2", "0", "-4", "-4"},
	{"2", "-2", "0", "4", "-4"},
	{"2", "2", "4", "0", "4"},
	{"100", "1", "101", "99", "100"},
	{"800", "799", "1599", "1", "639200"},
	{"-800.01", "12.00002", "-788.00998", "-812.01002", "-9600.1360002"},
	{"12.00002", "-800.01", "-788.00998", "812.01002", "-9600.1360002"},
	{"12.00002", "800.01", "812.01002", "-788.00998", "9600.1360002"},
	{"800.01", "800.01", "1600.02", "0.00", "640016.0001"},
	{"800.01", "-800.01", "0.00", "1600.02", "-640016.0001"},
	{"800.01", "-800.0100", "0.0000", "1600.0200", "-640016.000100"},
	{"800.0100", "-800.01", "0.0000", "1600.0200", "-640016.000100"},
	{"-800.01", "800.01", "0.00", "-1600.02", "-640016.0001"},
	{"-800.01", "-800.01", "-1600.02", "0.00", "640016.0001"},
	{"-800.01", "-800.0100", "-1600.0200", "0.0000", "640016.000100"},
	{"-800.0100", "-800.01", "-1600.0200", "0.0000", "640016.000100"},
	{"-800.01", "-12.00002", "-812.01002", "-788.00998", "9600.1360002"},
	{"0.0001", "0.04545", "0.04555", "-0.04535", "0.000004545"},
}

func TestAdd(t *testing.T) {
	fmt.Printf("\nTestAdd...\n")
	for _, c := range cases {
		fmt.Printf("%v + %v = ", c.param1, c.param2)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}

		var result string
		expectedStr := c.expectedAdd

		n3 := &bigfloat.BigFloat{}
		_, err = n3.Add(n1, n2)
		if err == nil {
			result = n3.String()
		}

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, err)
	}
}

func TestSub(t *testing.T) {
	fmt.Printf("\nTestSub...\n")
	for _, c := range cases {
		fmt.Printf("%v - %v = ", c.param1, c.param2)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}

		var result string
		expectedStr := c.expectedSub

		n3 := bigfloat.BigFloat{}
		_, err = n3.Sub(n1, n2)
		if err == nil {
			result = n3.String()
		}

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, err)
	}
}

func TestMul(t *testing.T) {
	fmt.Printf("\nTestMul...\n")
	for _, c := range cases {
		fmt.Printf("%v * %v = ", c.param1, c.param2)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}

		var result string
		expectedStr := c.expectedMul

		n3 := bigfloat.BigFloat{}
		_, err = n3.Mul(n1, n2)
		if err == nil {
			result = n3.String()
		}

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, err)
	}
}

func TestIsInt64(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   int64
		expected bool
	}{
		{"-800.01", -800, false},
		{"-800.00", -800, true},
	}
	fmt.Printf("\nTestIsInt64...\n")
	for _, c := range cases {
		fmt.Printf("isInt64(%v, %v) = ", c.param1, c.param2)
		n1, err := createBigFloat(t, c.param1)
		if err != nil {
			continue
		}

		expectedStr := strconv.FormatBool(c.expected)
		result := strconv.FormatBool(n1.IsInt64(c.param2))

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestFrac(t *testing.T) {
	var cases = []struct {
		param    string
		expected string
	}{
		{"-800.00", "0.00"},
		{"-800.01", "-0.01"},
		{"-800.0", "0.0"},
		{"-800", "0"},
	}
	fmt.Printf("\nTestFrac...\n")
	for _, c := range cases {
		fmt.Printf("frac(%v) = ", c.param)
		n1, err := createBigFloat(t, c.param)
		if err != nil {
			continue
		}

		expectedStr := c.expected
		result := n1.Frac().String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestTrunc(t *testing.T) {
	var cases = []struct {
		param1    string
		param2    int
		expected  string
		wantError bool
	}{
		{"-800.01", 2, "-800.00", false},
		{"-800.01", -1, "-800.00", true},
		{"123.45", 1, "123.0", false},
		{"-123.45", 1, "-123.0", false},
		{"-123.45", 2, "-123.00", false},
		{"-123.45", 3, "-123.000", false},
		{"-0.45", 3, "0.000", false},
	}
	fmt.Printf("\nTestTrunc...\n")
	for _, c := range cases {
		fmt.Printf("trunc(%v, %v) = ", c.param1, c.param2)
		n1, err := createBigFloat(t, c.param1)
		if err != nil {
			continue
		}

		_, err = n1.Trunc(bigfloat.WithDecimalPlaces(c.param2))
		if c.wantError {
			if err == nil {
				t.Errorf("ERROR: should be error\n")
			} else {
				fmt.Printf("OK: ERROR%v\n", err)
				continue
			}
		}

		result := n1.String()
		expectedStr := c.expected

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, err)
	}
}

func TestMulInt64(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   int64
		expected string
	}{
		{"-800.01", 12, "-9600.12"},
		{"-800.01", 0, "0.00"},
		{"-800.01", 1, "-800.01"},
		{"-800.01", -1, "800.01"},
		{"-800.01", 10, "-8000.1"},
		{"-800.01", -10, "8000.1"},
	}
	fmt.Printf("\nTestMulInt64...\n")
	for _, c := range cases {
		fmt.Printf("%v * %v = ", c.param1, c.param2)
		n, err := createBigFloat(t, c.param1)
		if err != nil {
			continue
		}

		_, err = n.MulInt64(c.param2)
		result := n.String()
		expectedStr := c.expected

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, err)
	}
}

func TestSetDecimals(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   int
		expected string
	}{
		{"-800.01", 12, "-800.010000000000"},
		{"-800.01", 1, "-800.0"},
		{"-0.01", 1, "0.0"},
	}
	fmt.Printf("\nTestSetDecimals...\n")
	for _, c := range cases {
		fmt.Printf("setDecimals(%v, %v) = ", c.param1, c.param2)
		n, err := createBigFloat(t, c.param1)
		if err != nil {
			continue
		}

		n.SetDecimals(c.param2)
		result := n.String()
		expectedStr := c.expected

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestCopy(t *testing.T) {
	var cases = []struct {
		param    string
		expected string
	}{
		{"-800.01", "-800.01"},
		{"800.01", "800.01"},
		{"0", "0"},
		{"0.01", "0.01"},
	}
	fmt.Printf("\nTestCopy...\n")
	for _, c := range cases {
		fmt.Printf("copy(%v) = ", c.param)
		n1, err := createBigFloat(t, c.param)
		if err != nil {
			continue
		}

		expectedStr := c.expected
		n2 := n1.Copy()
		result := n2.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestGetSign(t *testing.T) {
	var cases = []struct {
		param    string
		expected int
	}{
		{"-800.01", -1},
		{"800.01", 1},
		{"-0.01", -1},
		{"0.01", 1},
		{"0.0", 1},
		{"0", 1},
	}
	fmt.Printf("\nTestGetSign...\n")
	for _, c := range cases {
		fmt.Printf("getSign(%v) = ", c.param)
		n1, err := createBigFloat(t, c.param)
		if err != nil {
			continue
		}

		expectedStr := toString(int64(c.expected))
		result := toString(int64(n1.GetSign()))

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestSign(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   int
		expected string
	}{
		{"-800.01", 1, "800.01"},
		{"-800.01", -1, "-800.01"},
		{"0", 1, "0"},
		{"0", -1, "0"},
	}
	fmt.Printf("\nTestSign...\n")
	for _, c := range cases {
		fmt.Printf("sign(%v, %v) = ", c.param1, c.param2)
		n1, err := createBigFloat(t, c.param1)
		if err != nil {
			continue
		}

		n1.Sign(c.param2)

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestNeg(t *testing.T) {
	var cases = []struct {
		param    string
		expected string
	}{
		{"-800.01", "800.01"},
		{"800.01", "-800.01"},
		{"0", "0"},
	}
	fmt.Printf("\nTestNeg...\n")
	for _, c := range cases {
		fmt.Printf("neg(%v) = ", c.param)
		n1, err := createBigFloat(t, c.param)
		if err != nil {
			continue
		}

		n1.Neg()

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestAbs(t *testing.T) {
	var cases = []struct {
		param    string
		expected string
	}{
		{"-800.01", "800.01"},
		{"800.01", "800.01"},
		{"0", "0"},
	}
	fmt.Printf("\nTestNeg...\n")
	for _, c := range cases {
		fmt.Printf("abs(%v) = ", c.param)
		n1, err := createBigFloat(t, c.param)
		if err != nil {
			continue
		}

		n1.Abs()

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestCompare(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   string
		expected int
	}{
		{"-800.01", "800.01", -1},
		{"800.01", "800.01", 0},
		{"800.01", "-800.01", 1},
		{"800.01", "1", 1},
		{"1", "800.01", -1},
		{"1", "0", 1},
		{"0", "1.01", -1},
		{"0", "0", 0},
		{"0", "0.00", 0},
	}
	fmt.Printf("\nTestCompare...\n")
	for _, c := range cases {
		fmt.Printf("compare(%v, %v) = ", c.param1, c.param2)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}

		expectedStr := toString(int64(c.expected))
		result := toString(int64(n1.Compare(n2)))

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestCompareAbs(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   string
		expected int
	}{
		{"-800.01", "800.01", 0},
		{"800.01", "800.01", 0},
		{"-800.01", "0.01", 1},
		{"-0.01", "800.01", -1},
		{"800.01", "-800.01", 0},
		{"0", "0", 0},
	}
	fmt.Printf("\nTestCompareAbs...\n")
	for _, c := range cases {
		fmt.Printf("compareAbs(%v, %v) = ", c.param1, c.param2)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}

		expectedStr := toString(int64(c.expected))
		result := toString(int64(n1.CompareAbs(n2)))

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestRound(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   int
		expected string
	}{
		{"-800.01", 1, "-800.00"},
		{"-1.55555", 1, "-1.60000"},
		{"1.25", 3, "1.25"},
		{"1.25", 2, "1.25"},
		{"1.25", 1, "1.30"},
		{"0.12", 1, "0.10"},
		{"-0.02", 1, "0.00"},
	}
	fmt.Printf("\nTestRound...\n")
	for _, c := range cases {
		fmt.Printf("round(%v, %v) = ", c.param1, c.param2)
		n1, err := createBigFloat(t, c.param1)
		if err != nil {
			continue
		}

		n1.Round(c.param2)

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestPow10(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   int
		expected string
	}{
		{"-800.01", 1, "10"},
		{"-800.01", 0, "1"},
		{"-800.01", -1, "0.1"},
	}
	fmt.Printf("\nTestPow10...\n")
	for _, c := range cases {
		fmt.Printf("pow10(%v, %v) = ", c.param1, c.param2)
		n1, err := createBigFloat(t, c.param1)
		if err != nil {
			continue
		}

		n1.Pow10(c.param2)

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestMul10(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   int
		expected string
	}{
		{"-800.01", 1, "-8000.1"},
		{"-800.01", 2, "-80001"},
		{"-800.01", 3, "-800010"},
	}
	fmt.Printf("\nTestMul10...\n")
	for _, c := range cases {
		fmt.Printf("mul10(%v, %v) = ", c.param1, c.param2)
		n1, err := createBigFloat(t, c.param1)
		if err != nil {
			continue
		}

		n1.Mul10(c.param2)

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestDiv10(t *testing.T) {
	var cases = []struct {
		param1   string
		param2   int
		expected string
	}{
		{"-800.01", 1, "-80.001"},
		{"-800.01", 3, "-0.80001"},
	}
	fmt.Printf("\nTestDiv10...\n")
	for _, c := range cases {
		fmt.Printf("div10(%v, %v) = ", c.param1, c.param2)
		n1, err := createBigFloat(t, c.param1)
		if err != nil {
			continue
		}

		n1.Div10(c.param2)

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestString(t *testing.T) {
	var cases = []struct {
		param    string
		expected string
	}{
		{" - 800.01", "-800.01"},
		{"-80001e-2", "-800.01"},
		{"-  0.01", "-0.01"},
		{"-0.0", "0.0"},
		{"-0", "0"},
	}
	fmt.Printf("\nTestString...\n")
	for _, c := range cases {
		fmt.Printf("string(%v) = ", c.param)
		n1, err := createBigFloat(t, c.param)
		if err != nil {
			continue
		}

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestSetString(t *testing.T) {
	var cases = []struct {
		param    string
		expected string
	}{
		{" - 800.01", "-800.01"},
		{"-80001e-2", "-800.01"},
		{"-  0.01", "-0.01"},
		{"-0.0", "0.0"},
		{"-0", "0"},
	}
	fmt.Printf("\nTestSetString...\n")
	for _, c := range cases {
		fmt.Printf("SetString(%v) = ", c.param)
		n1, err := bigfloat.SetString(c.param)
		if err != nil {
			t.Errorf("ERROR: %q is not valid big float number\n", err)
			continue
		}

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestSetInt64(t *testing.T) {
	var cases = []struct {
		param    int64
		expected string
	}{
		{-800, "-800"},
		{-80001e2, "-8000100"},
		{0, "0"},
		{-0, "0"},
		{-0e5, "0"},
	}
	fmt.Printf("\nTestSetInt64...\n")
	for _, c := range cases {
		fmt.Printf("SetInt64(%v) = ", c.param)
		n1 := bigfloat.SetInt64(c.param)

		expectedStr := c.expected
		result := n1.String()

		fmt.Printf("%v\n", result)
		printResult(t, result, expectedStr, nil)
	}
}

func TestErrors(t *testing.T) {
	cases := []interface{}{
		"1/0",
		"0/0",
		"-15/0",
		nil,
		"-9999999999999999999999999999999999999999999999999999999999 1/3",
		"ar",
		"a",
	}

	fmt.Printf("\nTestErrors...\n")
	for _, c := range cases {
		fmt.Printf("setString(%v) = ", c)
		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("\nOK: panic occurred: %v\n", err)
				}
			}()
			n, err := createBigFloat(nil, c.(string))
			if err != nil {
				panic(err)
			}
			fmt.Printf("%v\n", n.String())
			errorStr := fmt.Sprintf("%v should raise panic", c)
			fmt.Printf("\n" + errorStr + "\n")
			t.Errorf(errorStr)
		}()
	}
}

func TestErrorsDiv(t *testing.T) {
	var cases = []struct {
		param1 string
		param2 string
	}{
		{"1", "0"},
	}

	fmt.Printf("\nTestErrorsDiv...\n")
	for _, c := range cases {
		fmt.Printf("div(%v, %v) = ", c.param1, c.param2)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("\nOK: panic occurred: %v\n", err)
				}
			}()

			n3 := &bigfloat.BigFloat{}
			_, _, err := n3.Div(n1, n2)
			if err != nil {
				panic(err)
			}

			errorStr := fmt.Sprintf("%v should raise panic", c)
			fmt.Printf("\n" + errorStr + "\n")
			t.Errorf(errorStr)
		}()
	}
}

func TestErrorsDivMod(t *testing.T) {
	var cases = []struct {
		param1 string
		param2 string
	}{
		{"1", "0"},
	}

	fmt.Printf("\nTestErrorsDivMod...\n")
	for _, c := range cases {
		fmt.Printf("divmod(%v, %v) = ", c.param1, c.param2)
		n1, n2, err := create2BigFloats(t, c.param1, c.param2)
		if err != nil {
			continue
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("\nOK: panic occurred: %v\n", err)
				}
			}()

			n3 := &bigfloat.BigFloat{}
			_, _, err := n3.DivMod(n1, n2)
			if err != nil {
				panic(err)
			}

			errorStr := fmt.Sprintf("%v should raise panic", c)
			fmt.Printf("\n" + errorStr + "\n")
			t.Errorf(errorStr)
		}()
	}
}

func TestErrorsStringWithRepeatingDecimals(t *testing.T) {
	var cases = []struct {
		param interface{}
	}{
		{"1"},
		{nil},
		{1},
		{[]byte("abc")},
	}

	fmt.Printf("\nTestErrorsDivMod...\n")
	for _, c := range cases {
		fmt.Printf("div(1, 1) = ")
		n1, n2, err := create2BigFloats(t, "1", "1")
		if err != nil {
			continue
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("\nOK: panic occurred: %v\n", err)
				}
			}()

			n3 := &bigfloat.BigFloat{}
			_, repDec, err := n3.Div(n1, n2)
			if err != nil {
				panic(err)
			}

			result := bigfloat.StringWithRepeatingDecimals(n3, repDec, c.param)
			fmt.Printf("%v\n", result)

			errorStr := fmt.Sprintf("%v should raise panic", c)
			fmt.Printf("\n" + errorStr + "\n")
			t.Errorf(errorStr)
		}()
	}
}
