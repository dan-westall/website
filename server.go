package main

import (
	"github.com/gin-gonic/gin"
	"github.com/boltdb/bolt"
	"net/http"
	"path/filepath"
	"encoding/json"
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

var auth string

func main() {


	db, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}
	//close DB connection on termination
	defer db.Close()

	// We need to build the current direct path
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	router := gin.Default()
	router.LoadHTMLGlob(filepath.Join(exPath, "/views/**/*"))
	router.Static("/css", "src/website/assets/css")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": filepath.Join(exPath, "/views/**/*"),
		})
	})

	router.GET("/caravan-weather", func(c *gin.Context) {
		err = db.View(func(tx *bolt.Tx) error {
			c := tx.Bucket([]byte("DB")).Bucket([]byte("W_ENTRIES")).Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				log.Printf("key=%s, value=%s\n", k, v)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	})

	router.POST("/caravan-weather", func(c *gin.Context) {
		var json Login
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// No these passwords are not used in production.
		if json.User != "user" || json.Password != auth {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		err = addResult(db, json.Weather.Temperature, json.Weather.Luminosity)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{"status": "Result"})
	})

	router.Run(":8080")
}

func setupDB() (*bolt.DB, error) {

	db, err := bolt.Open("caravan_weather_v2.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		root, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			return fmt.Errorf("could not create bucket: %v", err)
		}

		_, err = root.CreateBucketIfNotExists([]byte("W_ENTRIES"))
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

func addResult(db *bolt.DB, celsius string, lumen string ) error {

	entry := WeatherReport{Temperature: celsius, Luminosity: lumen}
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("could not marshal entry json: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err = tx.Bucket([]byte("DB")).Bucket([]byte("W_ENTRIES")).Put([]byte(time.Now().Format(time.RFC3339)), entryBytes)
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
