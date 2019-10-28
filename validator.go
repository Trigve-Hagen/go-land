package main

import (
	"fmt"
	"net/smtp"
	"regexp"
	"strings"
)

// Message is the structure used for the contact form validation.
type Message struct {
	Name    string
	Email   string
	Subject string
	Message string
	Errors  map[string]string
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

// Login is the structure used for the contact form validation.
type Login struct {
	Uname    string
	Password string
	Errors   map[string]string
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

// Register is the structure used for the contact form validation.
type Register struct {
	Fname      string
	Lname      string
	Uname      string
	Email      string
	Userrole   string
	Password   string
	RePassword string
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

// VUser is the structure used for the user form validation.
type VUser struct {
	Fname  string
	Lname  string
	Uname  string
	Email  string
	Errors map[string]string
}

// ValidateUser is used to validate the edit profile form.
func (vuser *VUser) ValidateUser() bool {
	vuser.Errors = make(map[string]string)

	usn := regexp.MustCompile("[A-Za-z\\s]+")
	matched := usn.Match([]byte(vuser.Fname))
	if matched == false {
		vuser.Errors["Fname"] = "Please enter a valid first name."
	}
	match := usn.Match([]byte(vuser.Lname))
	if match == false {
		vuser.Errors["Lname"] = "Please enter a valid last name."
	}
	mat := usn.Match([]byte(vuser.Uname))
	if mat == false {
		vuser.Errors["Uname"] = "Please enter a valid username."
	}
	re := regexp.MustCompile(".+@.+\\..+")
	m := re.Match([]byte(vuser.Email))
	if m == false {
		vuser.Errors["Email"] = "Please enter a valid email address."
	}
	return len(vuser.Errors) == 0
}

// VCreateUser is the structure used for the create user form validation.
type VCreateUser struct {
	Fname      string
	Lname      string
	Uname      string
	Email      string
	Password   string
	RePassword string
	Userrole   string
	Errors     map[string]string
}

// ValidateCreateUser is used to validate the create user form.
func (vuser *VCreateUser) ValidateCreateUser() bool {
	vuser.Errors = make(map[string]string)

	usn := regexp.MustCompile("[A-Za-z\\s]+")
	matched := usn.Match([]byte(vuser.Fname))
	if matched == false {
		vuser.Errors["Fname"] = "Please enter a valid first name."
	}
	match := usn.Match([]byte(vuser.Lname))
	if match == false {
		vuser.Errors["Lname"] = "Please enter a valid last name."
	}
	mat := usn.Match([]byte(vuser.Uname))
	if mat == false {
		vuser.Errors["Uname"] = "Please enter a valid username."
	}
	re := regexp.MustCompile(".+@.+\\..+")
	m := re.Match([]byte(vuser.Email))
	if m == false {
		vuser.Errors["Email"] = "Please enter a valid email address."
	}
	return len(vuser.Errors) == 0
}

// VPassword is the structure used for the user password form validation.
type VPassword struct {
	Password   string
	RePassword string
	Errors     map[string]string
}

// ValidatePassword is used to validate the register form.
func (pass *VPassword) ValidatePassword() bool {
	pass.Errors = make(map[string]string)

	if strings.TrimSpace(pass.Password) == "" {
		pass.Errors["Password"] = "Please enter a valid password."
	}
	if strings.TrimSpace(pass.RePassword) == "" {
		pass.Errors["RePassword"] = "Please enter a valid password."
	}
	if pass.Password != pass.RePassword {
		pass.Errors["RePassword"] = "Passwords do not match."
	}
	return len(pass.Errors) == 0
}

// ForgotPassword is the structure used for the forgot password form validation.
type ForgotPassword struct {
	Email  string
	Errors map[string]string
}

// ValidateForgotPassword is used to validate the forgot password form.
func (fp *ForgotPassword) ValidateForgotPassword() bool {
	fp.Errors = make(map[string]string)

	re := regexp.MustCompile(".+@.+\\..+")
	m := re.Match([]byte(fp.Email))
	if m == false {
		fp.Errors["Email"] = "Please enter a valid email address."
	}

	return len(fp.Errors) == 0
}

// VNewsletter is the structure used for the newsletter form validation.
type VNewsletter struct {
	NEmail string
	Errors map[string]string
}

// ValidateNewsletter is used to validate the newsletter form.
func (new *VNewsletter) ValidateNewsletter() bool {
	new.Errors = make(map[string]string)

	re := regexp.MustCompile(".+@.+\\..+")
	m := re.Match([]byte(new.NEmail))
	if m == false {
		new.Errors["Email"] = "Please enter a valid email address."
	}

	return len(new.Errors) == 0
}

// VPosts is the structure used for the post form validation.
type VPosts struct {
	Title  string
	Body   string
	Errors map[string]string
}

// ValidatePost is used to validate the post form.
func (new *VPosts) ValidatePost() bool {
	new.Errors = make(map[string]string)

	new.Title = strings.Replace(new.Title, "'", "", -1)
	new.Title = strings.Replace(new.Title, ",", "", -1)
	usn := regexp.MustCompile("[0-9A-Za-z\\s\\.\\-\\?!]+")
	matched := usn.Match([]byte(new.Title))
	if matched == false {
		new.Errors["Title"] = "Please enter a valid post Title."
	}

	new.Body = strings.Replace(new.Body, "'", "", -1)
	new.Body = strings.Replace(new.Body, ",", "", -1)
	match := usn.Match([]byte(new.Body))
	if match == false {
		new.Errors["Body"] = "Please enter a valid post body."
	}

	return len(new.Errors) == 0
}
