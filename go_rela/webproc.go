// main.go
package main

import (
	"log"
    "html/template"
	"net/http"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3/s3manager"
    "github.com/aws/aws-sdk-go/service/s3"
    "fmt"
    "os"
    "io"
    "time"
)

type urlData struct {
    Content string
}

var urldata = urlData {}
var global_back_file = ""
func ShowUrl(w http.ResponseWriter, r *http.Request) {
    if urldata.Content == "" {
        return
    }
    if r.Method == "POST" {
        renderHTML(w, "output.html", urldata)
    } else {
    }
}
func GetUrl() {
    if global_back_file == "" {
        return 
    }
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-west-2")},
    )

    svc := s3.New(sess)

    req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
    Bucket: aws.String("pkuhx1"),
    Key:    aws.String(global_back_file),
    })
    urlStr, err := req.Presign(15 * time.Minute)

    if err != nil {
        fmt.Println("Failed to sign request", err)
    }
    fmt.Println(urlStr)
    urldata.Content = urlStr
}
func UploadToS3(filename string) {

    bucket := "pkuhx1"
    file, err := os.Open(filename)
    if err != nil {
        exitErrorf("Unable to open file %q, %v", err)
    }

    defer file.Close()
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-west-2")},
    )

    uploader := s3manager.NewUploader(sess)
    _, err = uploader.Upload(&s3manager.UploadInput{
        Bucket: aws.String(bucket),
        Key: aws.String(filename),
        Body: file,
   })

   if err != nil {
       exitErrorf("Unable to upload %q to %q, %v", filename, bucket, err)
   }

   fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)
}

func exitErrorf(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
func renderHTML(w http.ResponseWriter, file string, data interface{}) {
    t, err := template.New(file).ParseFiles("views/" + file)
    checkErr(err)
    t.Execute(w, data)
}

//func index(w http.ResponseWriter, r *http.Request) {
//	renderHTML(w, "uploadfile.html", "no data")
//}

func page(w http.ResponseWriter, r *http.Request) {
    renderHTML(w, "uploadfile.html", "nodata")
	if r.Method == "POST" {
        fmt.Println("Starting upload...")
		r.ParseMultipartForm(1 << 32)
		file, handler, err := r.FormFile("fileUpload")
		if err != nil {
			log.Fatalln(err)
			return
		}
        defer file.Close()
        fmt.Println(handler.Filename)
        backname := "./save/"+handler.Filename + "_back.dat";
        f, err1 := os.OpenFile(backname, os.O_WRONLY | os.O_CREATE, 0666)
        checkErr(err1)
        defer f.Close()
        defer file.Close()
        io.Copy(f, file)
        UploadToS3(backname)
        os.Remove(backname)
        global_back_file = backname
        fmt.Println("End upload")
        GetUrl()
	} else {
      //  fmt.Println("Not Post method");
	}
}

func main() {
	// http.HandleFunc("/", index)            
	http.HandleFunc("/", page)

	http.HandleFunc("/output", ShowUrl)
	err := http.ListenAndServe(":9090", nil) 
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
