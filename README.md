### BigFloat

BigFloat implements multi-precision floating-point numbers.

[![codecov](https://codecov.io/gh/tihomirmagdic/bigfloat/graph/badge.svg?token=PTXHUP5GKZ)](https://codecov.io/gh/tihomirmagdic/bigfloat) [![Go Reference](https://pkg.go.dev/badge/github.com/tihomirmagdic/bigfloat.svg)](https://pkg.go.dev/github.com/tihomirmagdic/bigfloat)

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
  n1, err := bigfloat.Set("7.005")
  n2 := bigfloat.SetInt(4)
  n3 := bigfloat.New() // zero value is 0

  n3.Add(n1, n2)
  fmt.Printf("%v + %v = %v\n", n1, n2, n3)
  // Output: 7.005 + 4 = 11.005

  n3.Sub(n1, n2)
  fmt.Printf("%v - %v = %v\n", n1, n2, n3)
  // Output: 7.005 - 4 = 3.005

  n3.Mul(n1, n2)
  fmt.Printf("%v * %v = %v\n", n1, n2, n3)
  // Output: 7.005 * 4 = 28.020

  n3.Mul(n3.Mul(n3.Sub(n1, n2), n2), n1)
  fmt.Printf("((%v - %v) * %v) * %v = %v\n", n1, n2, n2, n1, n3)
  // Output: ((7.005 - 4) * 4) * 7.005 = 84.200100

  n3.Mul(n1, n2).Sub(n3, n1)
  fmt.Printf("%v * %v - %v = %v\n", n1, n2, n1, n3)
  // Output: 7.005 * 4 - 7.005 = 21.015

  n3.Div(n1, n2)
  fmt.Printf("%v / %v = %v\n", n1, n2, n3)
  // Output: 7.005 / 4 = 1.75125

  d := 4
  fmt.Printf("round(%v, %v) = ", n3, d)
  fmt.Printf("%v\n", n3.Round(d))
  // Output: round(1.75125, 4) = 1.7513

  fmt.Printf("trunc(%v) = ", n3)
  fmt.Printf("%v\n", n3.Trunc())
  // Output: trunc(1.75130) = 1.00000

  n1.Set(23)
  n2.Set(-11)

  _, remainder, _ := n3.DivMod(n1, n2)
  fmt.Printf("%v div %v = %v (remainder: %v)\n", n1, n2, n3, remainder)
  // Output: 23 div -11 = -2 (remainder: 1)

  n3.Div(n1, n2, bigfloat.WithDivDecimalPlaces(10))
  fmt.Printf("%v / %v = %v\n", n1, n2, n3)
  // Output: 23 / -11 = -2.0909090909

  _, repeatingDecimals, _ := n3.Div(n1, n2)
  fmt.Printf("%v / %v = %v\n", n1, n2, n3.StringF(repeatingDecimals))
  // Output: 23 / -11 = -2.(09)

  fmt.Printf("abs = %v\n", n3.Abs().StringF(repeatingDecimals, bigfloat.WithRepeatingOptions("r", ""), bigfloat.ForceSign(true)))
  // Output: abs = +2.r09
}
```

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

