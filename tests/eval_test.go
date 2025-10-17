/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// func Eval(exp ast.Expr) int {
// 	switch exp := exp.(type) {
// 	case *ast.BinaryExpr:
// 		return EvalBinaryExpr(exp)
// 	case *ast.BasicLit:
// 		switch exp.Kind {
// 		case token.INT:
// 			i, _ := strconv.Atoi(exp.Value)
// 			return i
// 		}
// 	}

// 	return 0
// }

// func EvalBinaryExpr(exp *ast.BinaryExpr) int {
// 	left := Eval(exp.X)
// 	right := Eval(exp.Y)

// 	switch exp.Op {
// 	case token.ADD:
// 		return left + right
// 	case token.SUB:
// 		return left - right
// 	case token.MUL:
// 		return left * right
// 	case token.QUO:
// 		return left / right
// 	}

// 	return 0
// }

// func ParseNEval(line string) (int, error) {
// 	exp, err := parser.ParseExpr(line)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return Eval(exp), nil
// }

func EvalBinaryExpr(exp *ast.BinaryExpr) int {
	left := Eval(exp.X)
	right := Eval(exp.Y)

	switch exp.Op {
	case token.ADD:
		return left + right
	case token.SUB:
		return left - right
	case token.MUL:
		return left * right
	case token.QUO:
		return left / right
	}

	return 0
}

func Eval(exp ast.Expr) int {
	switch exp := exp.(type) {
	case *ast.BinaryExpr:
		return EvalBinaryExpr(exp)
	case *ast.BasicLit:
		switch exp.Kind {
		case token.INT:
			i, _ := strconv.Atoi(exp.Value)
			return i
		}
	}

	return 0
}
func TestEvalFunction(t *testing.T) {
	testCases := []string{
		"1+2",
		"2-1",
		"1-2",
	}
	for _, v := range testCases {
		exp, err := parser.ParseExpr(v)
		if err != nil {
			fmt.Printf("parsing failed: %s\n", err)
			return
		}
		fmt.Printf("%s = %d\n", v, Eval(exp))

	}
}
