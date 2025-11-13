package logs

import "fmt"

func Success(text string) {
	fmt.Println("Success:", text)
}
func AccidentalFailure(text string) {
	fmt.Println("AccidentalFailure:", text)
}
func IntentionalFailure(text string) {
	fmt.Println("IntentionalFailure:", text)
}
func Debug(text string) {
	fmt.Println("Debug:", text)
}
