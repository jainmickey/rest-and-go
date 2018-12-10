package main

import (
	"errors"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photosets"
	"gopkg.in/masci/flickr.v2/photos"
	x2j "github.com/basgys/goxml2json"
	"fmt"
	"strings"
)

var envVars = map[string]string{}

func AddAndValidateEnvVars() {
	envVars["Port"] = os.Getenv("Port")
	envVars["FlickrPhotoSetId"] = os.Getenv("FlickrPhotoSetId")
	envVars["FlickrApiKey"] = os.Getenv("FlickrApiKey")
	envVars["FlickrSecretKey"] = os.Getenv("FlickrSecretKey")

	for k := range envVars {
		if envVars[k] == "" {
			log.Fatal(fmt.Sprintf("$%s must be set", k))
		}
	}
}

func ValidateAndGetQueryParam(param string, r *http.Request) (string, error) {
	paramValue := r.URL.Query().Get(param)
	if paramValue == "" {
		return paramValue, errors.New(fmt.Sprintf("Param %s is not provided", param))
	}
	return paramValue, nil
}

func FlickrPhotoDetailAPI(w http.ResponseWriter, r *http.Request) {
	photoId, err := ValidateAndGetQueryParam("photoId", r)
	w.Header().Add("Content-Type", "application/json")
	if err == nil {
		client := flickr.NewFlickrClient(envVars["FlickrApiKey"], envVars["FlickrSecretKey"])
		client.Init()
		client.EndpointUrl = flickr.API_ENDPOINT
		client.HTTPVerb = "POST"
		client.Args.Set("method", "flickr.photos.getInfo")
		client.Args.Set("photo_id", photoId)
		client.ApiSign()

		response := &photos.PhotoInfoResponse{}
		err := flickr.DoPost(client, response)
		if err != nil {
			fmt.Fprintln(w,"Error: ", err)
		} else {
			xml := strings.NewReader(response.Extra)
			jsn, err := x2j.Convert(xml)
			if err != nil {
				fmt.Fprintf(w, "Error in parsing: %s", err)
			} else {
				fmt.Fprintln(w, jsn)
			}
		}
	} else {
		errDict := map[string]string{"message": "Photo Id is not provided"}
		jsn, err := json.Marshal(errDict)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, string(jsn))
	}
}

func FlickrPhotoListAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	client := flickr.NewFlickrClient(envVars["FlickrApiKey"], envVars["FlickrSecretKey"])
	client.Init()
    client.Args.Set("method", "flickr.photosets.getPhotos")
	client.Args.Set("photoset_id", envVars["FlickrPhotoSetId"])
	client.Args.Set("extras",
		"url_k,url_h,url_o,url_m,owner_name,url_k,url_h,url_o,url_m,owner_name,views,date_upload")
	// sign the client for authentication and authorization
	client.ApiSign()
	response := &photosets.PhotosListResponse{}
	err := flickr.DoGet(client, response)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w,"Error: Random")
	} else {
		xml := strings.NewReader(response.Extra)
		jsn, err := x2j.Convert(xml)
		if err != nil {
			fmt.Fprintf(w, "Error in parsing: %s", err)
		} else {
			fmt.Fprintln(w, jsn)
		}
	}
}

func main() {
	AddAndValidateEnvVars()

	router := mux.NewRouter() // create routes
	router.HandleFunc("/api/flickr/detail", FlickrPhotoDetailAPI).Methods("GET")
	router.HandleFunc("/api/flickr", FlickrPhotoListAPI).Methods("GET")

	// These two lines are important if you're designing a front-end to utilise this API methods
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT"})

	// Launch server with CORS validations
	if err := http.ListenAndServe(":" + envVars["Port"], handlers.CORS(allowedOrigins,
		                          allowedMethods)(router)); err != nil {
		log.Fatal(err)
	}
}

