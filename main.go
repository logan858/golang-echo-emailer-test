package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

	// get/post/whatever we want to do in a db, with name var here.
	var lineIDs string
	lineResults, err := db.QueryContext(ctx, `SELECT odtl.id AS lineIDs
			FROM ORDER_HDR a
			LEFT OUTER JOIN ORDER_DTL odtl ON odtl.orderID = a.id
			WHERE a.numb = @name`, sql.Named("name", name))
	if err != nil {
		log.Fatal(err)
	}
	defer lineResults.Close()
	var allLineIDs []string
	for lineResults.Next() {
		err := lineResults.Scan(&lineIDs)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(lineIDs)
		allLineIDs = append(allLineIDs, lineIDs)
	}
	err = lineResults.Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(allLineIDs)

	m := gomail.NewMessage()
	m.SetHeader("From", emailer)
	m.SetHeader("To", "lmajor@levelwear.com")
	m.SetHeader("Subject", "Order# "+name)
	m.SetBody("text/plain", "Line IDs: "+"\n"+strings.Join(allLineIDs, "\n"))
	d := gomail.NewDialer(emailerIP, 25, "", "")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		panic(err)
	}
	return c.String(http.StatusOK, name)
}
