package evaluator

import (
	"myinterpreter/ast"
	"myinterpreter/object"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatements(node)
	case *ast.IfExpression:
		return evalIfExpression(node)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		return &object.ReturnValue{Value: val}
	}
	return nil
}

func evalBlockStatements(bs *ast.BlockStatement) object.Object {
	var res object.Object

	for _, stmt := range bs.Statements {
		res = Eval(stmt)
		if res != nil && res.Type() == object.RETURN_VALUE_OBJ {
			return res
		}
	}
	return res
}

func evalProgram(p *ast.Program) object.Object {
	var res object.Object

	for _, stmt := range p.Statements {
		res = Eval(stmt)
		if returnValue, ok := res.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}
	return res
}

func evalIfExpression(ie *ast.IfExpression) object.Object {
	cond := Eval(ie.Condition)
	if isTruthy(cond) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
	} else {
		return NULL
	}
}

func evalPrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return NULL
	}
}

func evalMinusPrefixOperatorExpression(obj object.Object) object.Object {
	if obj.Type() != object.INTEGER_OBJ {
		return NULL
	}
	value := obj.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalBangOperatorExpression(obj object.Object) object.Object {
	switch obj {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func nativeBoolToBooleanObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}

func evalInfixExpression(op string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(op, left, right)
	case op == "==": //==和!=直接进行指针比较
		return nativeBoolToBooleanObject(left == right)
	case op == "!=":
		return nativeBoolToBooleanObject(left != right)

	default:
		return NULL
	}
}

func evalIntegerInfixExpression(op string, left, right object.Object) object.Object {
	lValue, rValue := left.(*object.Integer).Value, right.(*object.Integer).Value

	switch op {
	case "+":
		return &object.Integer{Value: lValue + rValue}
	case "-":
		return &object.Integer{Value: lValue - rValue}
	case "*":
		return &object.Integer{Value: lValue * rValue}
	case "/":
		return &object.Integer{Value: lValue / rValue}
	case "<":
		return nativeBoolToBooleanObject(lValue < rValue)
	case ">":
		return nativeBoolToBooleanObject(lValue > rValue)
	case "==":
		return nativeBoolToBooleanObject(lValue == rValue)
	case "!=":
		return nativeBoolToBooleanObject(lValue != rValue)
	default:
		return NULL
	}
}

//func evalStatements(stmts []ast.Statement) object.Object {
//	var res object.Object
//
//	for _, stmt := range stmts {
//		res = Eval(stmt)
//		if returnValue, ok := res.(*object.ReturnValue); ok {
//			return returnValue.Value
//		}
//	}
//	return res
//}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}
