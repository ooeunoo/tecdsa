// package main

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"
// 	"strings"
// )

// func main() {
// 	rootDir := "." // 현재 디렉토리를 루트로 설정
// 	outputFile := "one.txt"

// 	file, err := os.Create(outputFile)
// 	if err != nil {
// 		fmt.Printf("파일 생성 오류: %v\n", err)
// 		return
// 	}
// 	defer file.Close()

// 	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		if !info.IsDir() && strings.HasSuffix(path, ".go") {
// 			relPath, _ := filepath.Rel(rootDir, path)
// 			content, err := ioutil.ReadFile(path)
// 			if err != nil {
// 				return err
// 			}

// 			_, err = fmt.Fprintf(file, "// %s\n%s\n\n", relPath, string(content))
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		fmt.Printf("오류 발생: %v\n", err)
// 	} else {
// 		fmt.Printf("%s 파일이 생성되었습니다.\n", outputFile)
// 	}
// }

package main

import (
	"fmt"
	"net/http"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	url := "http://localhost:8080/key_gen"

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				fmt.Println("Failed to create request:", err)
				return
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Failed to send request:", err)
				return
			}
			defer resp.Body.Close()

			fmt.Printf("Response status for request %d: %s\n", i+1, resp.Status)
		}()
	}

	wg.Wait()
	fmt.Println("All requests completed.")
}
