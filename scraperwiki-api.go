package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
)

type Scrapers struct {
	Owner  []string
	Editor []string
}
type InfoDict struct {
	Username    string
	ProfileName string
	CodeRoles   map[string][]string
	DateJoined  string
}

type UserInfoList struct {
	Information []InfoDict
}

func getInfo(username string) (InfoDict, error) {
	address := fmt.Sprintf("https://api.scraperwiki.com/api/1.0/scraper/getuserinfo?format=jsondict&username=%s", username)
	cresp, err := http.Get(address)
	if err != nil {
		return InfoDict{}, err
	}
	defer cresp.Body.Close()

	dec := json.NewDecoder(cresp.Body)

	var items []InfoDict
	if err := dec.Decode(&items); err != nil {
		return InfoDict{}, err
	}

	if len(items) == 0 {
		return InfoDict{}, errors.New("User not found")
	}

	return items[0], nil
}

func getDB(name string, output_folder string) error {

	address := fmt.Sprintf("https://scraperwiki.com/scrapers/export_sqlite/%s/", name)

	resp, err := http.Head(address)
	fmt.Printf("    File is reportedly %s bytes\n", resp.Header.Get("Content-Length"))
	defer resp.Body.Close()

	length, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 0)
	if length == 0 {
		fmt.Println("    Skipping download, no data")
		return nil
	}

	output_file := path.Join(output_folder, name+".sqlite")

	// Check if the file already exists and how large it is 
	st, err := os.Stat(output_file)
	if err == nil {
		if st.Size() == length {
			fmt.Println("      Skipping download, already have data")
			return nil
		}
	}

	f, err := os.Create(output_file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	cresp, err := http.Get(address)
	defer cresp.Body.Close()

	n, err := io.Copy(f, cresp.Body)
	fmt.Printf("     Wrote %d bytes\n", n)
	return err
}

func getCode(name string, output_folder string) error {
	address := fmt.Sprintf("https://api.scraperwiki.com/api/1.0/scraper/getinfo?format=jsondict&name=%s&version=-1&quietfields=attachable_here%7Cattachables%7Ctags%7Clast_run%7chistory%7Cdatasummary%7Cuserroles%7Crunevents%7Clast_run", name)

	resp, err := http.Get(address)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var items []map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&items); err != nil {
		return err
	}

	languages := map[string]string{
		"python": ".py",
		"ruby":   ".rb",
		"php":    ".php",
		"html":   ".html",
	}

	code := fmt.Sprintf("%v", items[0]["code"])
	if len(code) == 0 {
		fmt.Println("      Skipping writing code as there is none")
		return nil
	}

	language := fmt.Sprintf("%v", items[0]["language"])
	output_file := path.Join(output_folder, name+languages[language])
	f, err := os.Create(output_file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	l, err := f.WriteString(code)
	if err != nil {
		panic("Failed to write code to file")
	}
	fmt.Printf("      Wrote %d bytes of code to %s\n", l, output_file)
	return nil
}