package age

func code_smell1() {
	prepare("This should be a constant") // Noncompliant; 'This should ...' is duplicated 3 times
	execute("This should be a constant")
	release("This should be a constant")
}

func code_smell2() { // Noncompliant
}
