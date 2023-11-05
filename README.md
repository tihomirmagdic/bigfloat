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

