package graph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/gosuri/uiprogress"
	"github.com/piprate/json-gold/ld"

	"earthcube.org/Project418/gleaner/internal/common"
	minio "github.com/minio/minio-go"
	"github.com/spf13/viper"
)

// GraphNG is a new and improved RDF conversion
func GraphNG(mc *minio.Client, prefix string, v1 *viper.Viper) error {
	bucketname := "gleaner" //  "gleaner-summoned"
	// prefix := fmt.Sprintf("summoned/%s", suffix)

	// My go func controller vars
	semaphoreChan := make(chan struct{}, 10) // a blocking channel to keep concurrency under control (1 == single thread)
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	// Make a common proc and options to share with the upcoming go funcs
	proc, options := common.JLDProc(v1)

	// params for list objects calls
	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true

	// Spiffy progress line  (do I really want this?)
	uiprogress.Start()
	x := 0 // ugh..  why won't len(oc) work..   buffered channel issue I assume?
	for range mc.ListObjectsV2(bucketname, prefix, isRecursive, doneCh) {
		x = x + 1
	}
	count := x
	bar1 := uiprogress.AddBar(count).PrependElapsed().AppendCompleted()
	bar1.PrependFunc(func(b *uiprogress.Bar) string {
		return rightPad2Len("miller", " ", 15)
	})
	bar1.Fill = '-'
	bar1.Head = '>'
	bar1.Empty = ' '

	for object := range mc.ListObjectsV2(bucketname, prefix, isRecursive, doneCh) {
		wg.Add(1)
		go func(object minio.ObjectInfo) {
			semaphoreChan <- struct{}{}
			_, err := obj2RDF(bucketname, "milled", mc, object, proc, options)
			if err != nil {
				log.Println(err) // need to log to an "errors" log file
			}

			wg.Done() // tell the wait group that we be done
			// log.Printf("Doc: %s error: %v ", name, err) // why print the status??

			bar1.Incr()
			<-semaphoreChan
		}(object)
	}
	wg.Wait()  
   

	// uiprogress.Stop()

	// all done..  write the full graph to the object store
	// log.Printf("Building result graph from: %s/milled-dg/%s", bucketname, prefix)

	// log.Printf("Processed prefix: %s", prefix)
	millprefix := strings.ReplaceAll(prefix, "summoned", "milled")
	log.Printf("Building result graph from prefix: %s to: %s", prefix, millprefix)

	mcfg := v1.GetStringMapString("gleaner")
	err := pipeCopyNG(fmt.Sprintf("%s_graph.nq", mcfg["runid"]), "gleaner", millprefix, mc)
    if err != nil {
		log.Printf("Error on pipe copy: %s", err)
	} else {
		log.Println("Pipe copy for graph done")
	}

	return err
}

// func obj2RDF(fo io.Reader, key string, mc *minio.Client) (string, int64, error) {
func obj2RDF(bucketname, prefix string, mc *minio.Client, object minio.ObjectInfo, proc *ld.JsonLdProcessor, options *ld.JsonLdOptions) (string, error) {
	// object is an object reader
	fo, err := mc.GetObject(bucketname, object.Key, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	key := object.Key // replace if new function idea works..

	var b bytes.Buffer
	bw := bufio.NewWriter(&b)

	_, err = io.Copy(bw, fo)
	if err != nil {
		log.Println(err)
	}

	// TODO
	// Process the bytes in b to RDF (with randomized blank nodes)
	rdf, err := jld2nq(string(b.Bytes()), key, proc, options)
	if err != nil {
		return key, err
	}

	rdfubn := GlobalUniqueBNodes(rdf)

	milledkey := strings.ReplaceAll(key, ".jsonld", ".rdf")
	milledkey = strings.ReplaceAll(milledkey, "summoned/", "")

	// make an object with prefix like  scienceorg-dg/objectname.rdf  (where is had .jsonld before)
	objectName := fmt.Sprintf("%s/%s", prefix, milledkey)
	//contentType := "application/ld+json"
	usermeta := make(map[string]string) // what do I want to know?
	usermeta["origfile"] = key
	//		usermeta["url"] = urlloc
	//		usermeta["sha1"] = sha
	//		bucketName := "gleaner-summoned" //   fmt.Sprintf("gleaner-summoned/%s", k) // old was just k

	// Upload the file
	_, err = LoadToMinio(rdfubn, "gleaner", objectName, mc)
	if err != nil {
		return objectName, err
	}

	return objectName, nil
}

// jld2nq converts JSON-LD documents to NQuads
func jld2nq(jsonld, key string, proc *ld.JsonLdProcessor, options *ld.JsonLdOptions) (string, error) {

	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Printf("Error when transforming %s JSON-LD document to interface: %v", key, err)
		return "", err
	}

	nq, err := proc.ToRDF(myInterface, options)
	if err != nil {
		log.Printf("Error when transforming %s  JSON-LD document to RDF: %v", key, err)
		return "", err
	}

	return nq.(string), err
}

// sugar function for the ui bar
func rightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

func pipeCopyNG(name, bucket, prefix string, mc *minio.Client) error {
	log.Println("Start pipe reader / writer sequence")

	pr, pw := io.Pipe()     // TeeReader of use?
	lwg := sync.WaitGroup{} // work group for the pipe writes...
	lwg.Add(2)

	// params for list objects calls
	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true

	go func() {
		defer lwg.Done()
		defer pw.Close()
		for object := range mc.ListObjectsV2(bucket, prefix, isRecursive, doneCh) {
			fo, err := mc.GetObject(bucket, object.Key, minio.GetObjectOptions{})
			if err != nil {
				fmt.Println(err)
			}

			var b bytes.Buffer
			bw := bufio.NewWriter(&b)

			_, err = io.Copy(bw, fo)
			if err != nil {
				log.Println(err)
			}

			pw.Write(b.Bytes())
		}

	}()

	// log.Printf("%s_graph.nq", name)

	// go function to write to minio from pipe
	go func() {
		defer lwg.Done()
		_, err := mc.PutObject("gleaner", name, pr, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Println(err)
		}
	}()

	// Note: We can also make a file and pipe write to that, keep this code around in case
	// f, err := os.Create(fmt.Sprintf("%s_graph.nq", prefix))  // needs a f.Close() later
	// if err != nil {
	// 	log.Println(err)
	// }
	// go function to write to file from pipe
	// go func() {
	// 	defer lwg.Done()
	// 	if _, err := io.Copy(f, pr); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	lwg.Wait() // wait for the pipe read writes to finish
	pw.Close()
	pr.Close()

	return nil
}
