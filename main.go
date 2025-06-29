package main

import (
	"fmt"
	"concurentGrep/workList"
	"concurentGrep/worker"
	"os"
	"path/filepath"
	"sync"
)

// function to get all the files  in a particular directory
func getAllFiles(wl *worklist.Worklist, path string){
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Error:",err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir(){
			// if next path is found to be directory the path and directory
			// name is joined for full path and the files in sub directory 
			// is recursively traversed
			nextPath := filepath.Join(path, entry.Name())
			getAllFiles(wl, nextPath)
		} else {
			wl.Add(worklist.NewJob(filepath.Join(path, entry.Name())))
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <search_term> <directory>")
		return
	}
	var workerWg sync.WaitGroup
	wl := worklist.New(100)
	results := make(chan worker.Result, 100)
	numWorkers := 10

	workerWg.Add(1)

	// get all files
	go func(){
		defer workerWg.Done()
		getAllFiles(&wl, os.Args[2])
		wl.Finalize(numWorkers)
	}()
	
	// find matches
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go func(){
			defer workerWg.Done()
			for{
				workEntry := wl.Next()
				if workEntry.Path != ""{
					workerResult := worker.FindInFile(workEntry.Path,os.Args[1])
					if workerResult != nil {
						for _, r := range workerResult.Inner {
							results <- r
						}
					}
				} else {
					// When the path is empty, this indicates that there 
					// are no more jobs available, so quit
					return
				}
			}
		}()
	}
	blockWorkerWg := make(chan struct{})
	go func(){
		workerWg.Wait()
		// close channel
		close(blockWorkerWg)
	}()

	var displayWg sync.WaitGroup
	displayWg.Add(1)

go func() {
		for {
			select {
			case r := <-results :
				fmt.Printf("%s:%d: %s \n",r.Path,r.LineNum,r.Line)

			case <- blockWorkerWg:
			// make sure channel is empty before aborting
			if len(results) == 0 {
					displayWg.Done()
					return
				}
			}
		}
	}()
displayWg.Wait()
}
