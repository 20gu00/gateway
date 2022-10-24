package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	fmt.Println(test("admin", "passwd"))
}
func test(salt, password string) string {
	s1 := sha256.New()
	s1.Write([]byte(password))
	str1 := fmt.Sprintf("%x", s1.Sum(nil))
	s2 := sha256.New()
	s2.Write([]byte(str1 + salt))
	return fmt.Sprintf("%x", s2.Sum(nil))
}

//f6da0c07372854658b17f13b696614989029773bd1457f68225d8b4339e48cec
