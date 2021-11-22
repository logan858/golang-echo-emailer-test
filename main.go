package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	gomail "gopkg.in/mail.v2"
)

var db *sql.DB
var err error

func main() {
	fmt.Println("blibli")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/order/:name", paramsFunc)
	e.GET("/", hello)
	e.Logger.Fatal(e.Start(":1323"))
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "hello func hit")
}

func paramsFunc(c echo.Context) error {
	name := c.Param("name")

	emailerIP := os.Getenv("EMAILERIP")
	emailer := os.Getenv("EMAILER")
	databaseConn := os.Getenv("DB")
	db, err = sql.Open("sqlserver", fmt.Sprintf("sqlserver://%s", databaseConn))
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	fmt.Println("\n Connected!")
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(15 * time.Second)

	ctx := context.Background()
	err2 := db.PingContext(ctx)
	if err2 != nil {
		log.Fatal("Error pinging db: " + err2.Error())
	}

	fmt.Print(ctx)
	// get/post/whatever we want to do in a db, with name var here.
	var result struct {
		id string
	}
	err = db.QueryRowContext(ctx, `SELECT a.id
			FROM ORDER_HDR a
			WHERE a.numb = @name`, sql.Named("name", name)).Scan(&result.id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n", "id: ", name, "id results: ", result.id)

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
