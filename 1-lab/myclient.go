package main

import (
	"fmt"
	// "io"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	dataSizes := []int{1, 1000, 1000000, 100000000 }
	urls := getUrlsFromNums("localhost", dataSizes)

	epochs := 3

	fmt.Printf(
		"===== Async Tests =====\n" +
		"requesting for %v bytes asynchoronously '%d' times\n\n",
	dataSizes, epochs)

	asyncTimes := testAsync(epochs, urls)
	fmt.Println("Time for each set", asyncTimes)
	var asyncTotal float64 = 0.0
	for _, time := range asyncTimes{
		asyncTotal += time
	}
	fmt.Printf("Average time : %f\n\n", asyncTotal/float64(epochs))


	fmt.Printf(
		"===== Sync Tests =====\n" +
		"requesting for %v bytes synchoronously '%d' times\n\n",
	dataSizes, epochs)

	syncTimes := testSync(epochs, urls)
	fmt.Println("Time for each set", syncTimes)
	var syncTotal float64 = 0.0
	for _, time := range syncTimes{
		syncTotal += time
	}
	fmt.Printf("Average time : %f\n\n", syncTotal/float64(epochs))
}

func getUrlsFromNums(domain string, bytes []int) []string{
		var urls = make([]string, len(bytes))
		for i, n := range bytes{
				urls[i] = fmt.Sprintf("http://%s:9999/genData?numBytes=%d", domain, n)
		}
		return urls
}

func testAsync(times int, urls []string) []float64{
		ch := make(chan string)
		//response time for each corresponding request set
		resTimes := make([]float64, times)

		for i := 0; i < times; i++{
				start := time.Now()

				for _, url := range urls{
						go fetch(url, ch)
				}
				for range urls{
						<- ch
				}

				resTimes[i] = time.Since(start).Seconds()
		}
		return resTimes
}

func testSync(times int, urls []string) []float64{
		ch := make(chan string)
		//response time for each corresponding request set
		resTimes := make([]float64, times)

		for i := 0; i < times; i++{
				start := time.Now()

				for _, url := range urls{
						go fetch(url, ch)
						<- ch
				}

				resTimes[i] = time.Since(start).Seconds()
		}
		close(ch)
		return resTimes
}

func fetch(url string, ch chan<- string) {

		resp, err := http.Get(url)
		if err != nil {
			ch <- fmt.Sprint(err) // send to channel ch
			return
		}

		_, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close() // don't leak resources
		if err != nil {
			ch <- fmt.Sprintf("while reading %s: %v", url, err)
			return
		}

    // time.Sleep(100 * time.Milisecond) // simulate processing
		ch <- ""
}
