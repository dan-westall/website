package main

import (
	"github.com/gin-gonic/gin"
	"github.com/boltdb/bolt"
	"net/http"
	"path/filepath"
	"os"
	"log"
	"time"
	"fmt"
)

type Login struct {
	User     string `form:"user" json:"user" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Weather struct {
		WeatherReport
	} ` binding:"required"`
}

type WeatherReport struct {
	Temperature string `json:"temp" binding:"required"`
	Luminosity  string `json:"luminosity" binding:"required"`
}

func main() {
	db, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}
	//close DB connection on termination
	defer db.Close()

	router := gin.Default()
	router.LoadHTMLGlob(filepath.Join(os.Getenv("GOPATH"), "src/website/views/**/*"))
	router.Static("/css", "src/website/assets/css")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Hello World",
		})
	})

	router.POST("/caravan-weather", func(c *gin.Context) {
		var json Login
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// No these passwords are not used in production.
		if json.User != "foo" || json.Password != "bar" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		err = addLuminosity(db, json.Weather.Luminosity)
		if err != nil {
			log.Fatal(err)
		}

		err = addTemperature(db, json.Weather.Temperature)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{"status": "Result"})
	})

	router.Run(":8080")
}

func setupDB() (*bolt.DB, error) {

	db, err := bolt.Open("caravan_weather.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		root, err := tx.CreateBucketIfNotExists([]byte("LUMINOSITY"))
		if err != nil {
			return fmt.Errorf("could not create bucket: %v", err)
		}

		_, err = root.CreateBucketIfNotExists([]byte("TEMPERATURE"))
		if err != nil {
			return fmt.Errorf("could not create bucket: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}

	fmt.Println("DB Setup Done")

	return db, nil
}

func addLuminosity(db *bolt.DB, lumens string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("LUMINOSITY"))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(time.Now().Format(time.RFC3339)), []byte(lumens))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func addTemperature(db *bolt.DB, celsius string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("TEMPERATURE"))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(time.Now().Format(time.RFC3339)), []byte(celsius))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	return nil
}
