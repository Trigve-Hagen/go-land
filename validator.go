package main

import (
	"fmt"
	"net/smtp"
	"regexp"
	"strings"
)

// Message is the structure used for the contact form.
type Message struct {
	Name       string
	Email      string
	Subject    string
	Message    string
	Role       int
	IfLoggedIn bool
	Errors     map[string]string
}

// ValidateMessage is used to validate the contact form.
func (msg *Message) ValidateMessage() bool {
	msg.Errors = make(map[string]string)

	re := regexp.MustCompile(".+@.+\\..+")
	matched := re.Match([]byte(msg.Email))
	if matched == false {
		msg.Errors["Email"] = "Please enter a valid email address."
	}
	if strings.TrimSpace(msg.Name) == "" {
		msg.Errors["Name"] = "Name must have value."
	}
	if strings.TrimSpace(msg.Message) == "" {
		msg.Errors["Message"] = "Please write a message."
	}

	return len(msg.Errors) == 0
}

// Deliver is used to deliver the mail for the contact form.
func (msg *Message) Deliver() error {
	to := []string{"someone@example.com"}
	body := fmt.Sprintf("Reply-To: %v\r\nSubject: New Message\r\n%v", msg.Email, msg.Message)

	username := "you@gmail.com"
	password := "..."
	auth := smtp.PlainAuth("", username, password, "smtp.gmail.com")

	return smtp.SendMail("smtp.gmail.com:587", auth, msg.Email, to, []byte(body))
}

// Login is the structure used for the contact form.
type Login struct {
	Uname      string
	Password   string
	Role       int
	IfLoggedIn bool
	Errors     map[string]string
}

// ValidateLogin is used to validate the login form.
func (lgn *Login) ValidateLogin() bool {
	lgn.Errors = make(map[string]string)

	usn := regexp.MustCompile("[A-Za-z\\s]+")
	matched := usn.Match([]byte(lgn.Uname))
	if matched == false {
		lgn.Errors["Uname"] = "Please enter a valid username."
	}
	if strings.TrimSpace(lgn.Password) == "" {
		lgn.Errors["Password"] = "Please enter a valid password."
	}

	return len(lgn.Errors) == 0
}

// Register is the structure used for the contact form.
type Register struct {
	Fname      string
	Lname      string
	Uname      string
	Email      string
	Password   string
	RePassword string
	IfLoggedIn bool
	Userrole   int8
	Errors     map[string]string
}

// ValidateRegister is used to validate the register form.
func (reg *Register) ValidateRegister() bool {
	reg.Errors = make(map[string]string)

	usn := regexp.MustCompile("[A-Za-z\\s]+")
	matched := usn.Match([]byte(reg.Fname))
	if matched == false {
		reg.Errors["Fname"] = "Please enter a valid first name."
	}
	match := usn.Match([]byte(reg.Lname))
	if match == false {
		reg.Errors["Lname"] = "Please enter a valid last name."
	}
	mat := usn.Match([]byte(reg.Uname))
	if mat == false {
		reg.Errors["Uname"] = "Please enter a valid username."
	}
	re := regexp.MustCompile(".+@.+\\..+")
	m := re.Match([]byte(reg.Email))
	if m == false {
		reg.Errors["Email"] = "Please enter a valid email address."
	}
	if strings.TrimSpace(reg.Password) == "" {
		reg.Errors["Password"] = "Please enter a valid password."
	}
	if strings.TrimSpace(reg.RePassword) == "" {
		reg.Errors["RePassword"] = "Please enter a valid password."
	}
	if reg.Password != reg.RePassword {
		reg.Errors["RePassword"] = "Passwords do not match."
	}
	return len(reg.Errors) == 0
}
