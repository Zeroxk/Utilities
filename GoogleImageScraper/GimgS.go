package main

import (
        "fmt"
        "log"
        "net/http"
        "io/ioutil"
        "encoding/json"
        //"os"
)

type Image struct {
    results map[string]interface{}
}

type RespData struct {
    ResponseData map[string]interface{}
}

func main() {
    resp, err := http.Get("https://ajax.googleapis.com/ajax/services/search/images?v=1.0&safe=active&q=Red");
    
    if(err != nil) {
        log.Fatal(err)
    }

    defer resp.Body.Close();

    body, err := ioutil.ReadAll(resp.Body)
    
    b, err := json.Marshal(body)
    //fmt.Println(string(body));

    var rpD RespData
    err = json.Unmarshal(b, &rpD);

    fmt.Println(rpD)


}

