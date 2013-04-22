package main

import (
        "fmt"
        "log"
        "net/http"
        "io/ioutil"
        "encoding/json"
		//"reflect"
        //"os"
)

type RespData struct {
    ResponseData Results
}

type Results struct {
	Results []Image
}

type Image struct {
    Url string
}

func main() {
    resp, err := http.Get("https://ajax.googleapis.com/ajax/services/search/images?v=1.0&safe=active&q=Red");
    
    if(err != nil) {
        log.Fatal(err)
    }

    body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close();

    var rpD RespData
    err = json.Unmarshal(body, &rpD)
	fmt.Println(rpD)
	
	//TODO: Write images from url to disk

}

