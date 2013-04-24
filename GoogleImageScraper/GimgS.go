package main

import (
        "fmt"
        "log"
        "net/http"
        "io/ioutil"
        "encoding/json"
		"strings"
        "os"
)

const dir = "imgsGo" 

type RespData struct {
    ResponseData Results
}

type Results struct {
	Results []Image
}

type Image struct {
    Url string
}

func readURL(url string) []byte {
	resp, err := http.Get(url);
    
    if(err != nil) {
        log.Fatal(err)
    }

    body, err := ioutil.ReadAll(resp.Body)
	
	if(err != nil) {
        log.Fatal(err)
		resp.Body.Close();
    }
	defer resp.Body.Close();
	
	return body
}

func main() {

	var keyword string
    fmt.Scanf("%s", &keyword)
	fmt.Println("Your keyword is:",keyword,"\n")
	
    query := "https://ajax.googleapis.com/ajax/services/search/images?v=1.0&safe=active&q=" + keyword
	fmt.Println(query)
	
    contents := readURL(query)

	fmt.Println("Parsing JSON")
    var rpD RespData
    err := json.Unmarshal(contents, &rpD)
	
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done parsing JSON\n")
	
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
	
	fmt.Println("Changing working directory to", dir)
	err = os.Chdir(dir)
	
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Starting image download")
	
	res := rpD.ResponseData.Results
	for _,img := range(res) {
		name := img.Url[ (strings.LastIndex(img.Url, "/")) + 1:]
		fmt.Println(name)
		image := readURL(img.Url)
		
		err = ioutil.WriteFile(name, image, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	
	fmt.Println("Finished")
}

