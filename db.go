package main

import (
	"fmt"
	"log"
	"encoding/json"
	"github.com/boltdb/bolt"
	"os/exec"
	"os"
	"strings"
	"strconv"
)

func setupDB() (*bolt.DB, error) {
	db, err := bolt.Open("swapi.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		// Set up a database
		_, err = root.CreateBucketIfNotExists([]byte("Starship"))
		_, err = root.CreateBucketIfNotExists([]byte("Planets"))
		_, err = root.CreateBucketIfNotExists([]byte("Species"))
		_, err = root.CreateBucketIfNotExists([]byte("People"))
		_, err = root.CreateBucketIfNotExists([]byte("Film"))
		_, err = root.CreateBucketIfNotExists([]byte("Vehicle"))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}
	fmt.Println("DB Setup Done")
	return db, nil
}

// Add content to the database
func addToBucket(db *bolt.DB, jsonStr string, id string, bucketName string) error {

	planetBytes := []byte(jsonStr)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("DB")).Bucket([]byte(bucketName)).Put([]byte(id), []byte(planetBytes))
		if err != nil {
			return fmt.Errorf("could not insert %s: %v", bucketName, err)
		}

		return nil
	})
	fmt.Println("Added " + bucketName)
	return err
}

// Add function
func testdb() {
	db, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	const (
		planet = 1
		species = 2
		vehicle = 3
		starship = 4
		people = 5
		film  = 6
	)
	var nameText = map[int]string{
		planet:  "planets",
		species: "species",
		vehicle: "vehicles",
		starship: "starships",
		people: "people",
		film:  "films",
	}
	var tabText = map[int]string{
		planet:  "Planets",
		species: "Species",
		vehicle: "Vehicle",
		starship: "Starship",
		people: "People",
		film:  "Film",
	}
	var curl []byte
	var cmd *exec.Cmd
	for i := 1; i < 100; i++ {
		for k := 1; k <= 6; k++ {
			ID := strconv.Itoa(i)
			cmd = exec.Command("curl", ("https://swapi.co/api/" + nameText[k] + "/" + ID + "/"))
			if curl, err = cmd.Output(); err != nil {
			    fmt.Println(err)
			    os.Exit(1)
			}

			JSONStr := strings.TrimRight(string(curl), "\n")
			fmt.Println(JSONStr)
			fmt.Println(ID)

			err = addToBucket(db, JSONStr, ID, tabText[k])
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("Planets"))
		b.ForEach(func(k, v []byte) error {
			fmt.Println(string(k), string(v),"\n")
			return nil
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("Species"))
		b.ForEach(func(k, v []byte) error {
			fmt.Println(string(k), string(v))
			return nil
		})
		return nil
	})

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("Planets"))
		v := string(b.Get([]byte("3")))
		fmt.Printf(v)
		var planet Planet
		err := json.Unmarshal([]byte(v), &planet)
		if err != nil {
			return fmt.Errorf("could not Unmarshal json string: %v", err)
		}
		fmt.Printf("\nThe name of the planet 3 is: %s\n", planet.Name)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

}