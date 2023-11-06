### BigFloat

BigFloat implements multi-precision floating-point numbers.

[![codecov](https://codecov.io/gh/tihomirmagdic/bigfloat/graph/badge.svg?token=PTXHUP5GKZ)](https://codecov.io/gh/tihomirmagdic/bigfloat)

It provides:
- addition, subtraction, multiplication, division (float and integer with modulus)
- rounding
- truncation
- conversion from/to string and int64
- comparison of numbers
- automatic precision
- each digit (whole number and decimal) occupies 1 byte
- support for repeating decimals

## Usage

```go
package main

import (
	"bigfloat"
	"fmt"
)

func main() {
	n1, _ := bigfloat.SetString("7.005")
	n2 := bigfloat.SetInt64(4)
	n3 := &bigfloat.BigFloat{}

	n3.Add(n1, n2)
	fmt.Printf("%v + %v = %v\n", n1, n2, n3)

	n3.Sub(n1, n2)
	fmt.Printf("%v - %v = %v\n", n1, n2, n3)

	n3.Mul(n1, n2)
	fmt.Printf("%v * %v = %v\n", n1, n2, n3)

	n3.Div(n1, n2)
	fmt.Printf("%v / %v = %v\n", n1, n2, n3)

	d := 4
	fmt.Printf("round(%v, %v) = %v\n", n3, d, n3.Round(d))

	a := -2
	b := 11

	n3.Div(bigfloat.SetInt(a), bigfloat.SetInt(b), bigfloat.WithDivDecimalPlaces(10))
	fmt.Printf("%v / %v = %v\n", a, b, n3)

	_, rd, _ := n3.Div(bigfloat.SetInt(a), bigfloat.SetInt(b))
	fmt.Printf("%v / %v = %v\n", a, b, bigfloat.StringWithRepeatingDecimals(n3, rd))

	fmt.Printf("trunc(%v) = ", n3)
	n3.Trunc()
	fmt.Printf("%v\n", n3)

	a = 1
	n1 = bigfloat.SetInt(a).Div10(2)
	b = 3

	_, rd, _ = n3.Div(n1, bigfloat.SetInt(b))
	fmt.Printf("%v / %v = %v\n", n1, b, bigfloat.StringWithRepeatingDecimals(n3, rd))

	a = 23
	b = -11

	_, remainder, _ := n3.DivMod(bigfloat.SetInt(a), bigfloat.SetInt(b))
	fmt.Printf("divmod(%v, %v) = %v, remainder: %v\n", a, b, n3, remainder)

	n1.SetString("23.85")
	n2.SetString("-11.01")
	_, remainder, _ = n3.DivMod(n1, n2)
	fmt.Printf("divmod(%v, %v) = %v, remainder: %v\n", n1, n2, n3, remainder)

	_, rd, _ = n3.Div(bigfloat.SetInt(1), bigfloat.SetInt(12))
	fmt.Printf("%v / %v = %v\n", 1, 12, bigfloat.StringWithRepeatingDecimals(n3, rd))
	fmt.Printf("%v / %v = %v\n", 1, 12, bigfloat.StringWithRepeatingDecimals(n3, rd, bigfloat.WithRepeatingOptions("r", "")))
}

```

TODO: docs

In the addition operation of two numbers, the sign of each addend is important. To reduce complexity and simplify the code, the replacement of addends is used in cases where addition is actually subtraction over zero.

The sign in addition is the sign of the larger addend (abs).

Just like in addition, the sign of each operand is important in subtraction of two numbers. To reduce complexity and simplify the code, the replacement of operands is used in cases where subtraction is actually addition over zero.

The sign in subtraction is the opposite sign of the 2nd operand if operands swapping is needed, or the sign of the 1st operand.

Below are tables for addition and subtraction for operands 5 and 8 with all combinations of their signs.

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

