package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	gomail "gopkg.in/mail.v2"
)

func main() {
	fmt.Println("blibli")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/:name", paramsFunc)
	e.GET("/", hello)
	e.Logger.Fatal(e.Start(":1323"))
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "hello func hit")
}

func paramsFunc(c echo.Context) error {
	name := c.Param("name")
	fmt.Println(name)
	emailerIP := os.Getenv("EMAILERIP")
	emailer := os.Getenv("EMAILER")

	// get/post/whatever we want to do in a db, with name var here.

	m := gomail.NewMessage()
	m.SetHeader("From", emailer)
	m.SetHeader("To", "lmajor@levelwear.com")
	m.SetHeader("Subejct", "bliblibli")
	m.SetBody("text/plain", name)

	d := gomail.NewDialer(emailerIP, 25, "", "")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		panic(err)
	}

	return c.String(http.StatusOK, name)
}
