// Pi on the Dart Board - Go Version with concurrency
// This version sends individual x,y, pairs to the workers = communication heavy
// A.ORAM Nov 2017
// Version 4. Same as V3 but ask for the number of dartboards to use from user.

package main

// Imported packages

import (
	"fmt" // for console I/O
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"
)

var msg = " rand in farmer only"

type dartCoords struct {
	x, y float64
}

const Pi = 3.14159

//const numDartBoards = 1500
const dartsToThrow = 1e7

var nextXY dartCoords
var totalCount int
var dartsThrown int

var n int // used to dump byte counts from file IO

// File OK check
func checkFileOK(ecode error) { // type 'error' is known to Go
	if ecode != nil {
		panic(ecode) // crash out of the program!
	}
}

func getNumDartBoards() int {
	var num int
	fmt.Print("Enter number of dart boards to throw darts at: ")
	fmt.Scanln(&num)
	fmt.Println(num, "boards") //*** debug
	return num
}

//////////////////////////////////////////////////////////////////////////////////
//  Main program, create required channels, then start goroutines in parallel.
//////////////////////////////////////////////////////////////////////////////////

func main() {
	rand.Seed(999)                      // Seed the random number generator (constant seed makes a repeatable program)
	numDartBoards := getNumDartBoards() // get number of dart boards to use
	//numDartBoards := 8
	fmt.Println("Max cores=", runtime.GOMAXPROCS(0), "Number of dart boards=", numDartBoards) // Set the max number of cores whilst seeing what it currently is.

	// Open a file to write results to.
	PiResults, fError := os.OpenFile("PiResults.txt", os.O_APPEND|os.O_CREATE, 0666) // create file if not existing, and append to it. 666 = permission bits rw-/rw-/rw-
	checkFileOK(fError)
	defer PiResults.Close() // defer means do this last, just before surrounding block terminates (ie 'main')

	nextXY.x = rand.Float64() // generate first set of coords in range 0.0 - 1.0 (float)
	nextXY.y = rand.Float64()

	// Set up required channels
	xyValues := make([]chan dartCoords, numDartBoards) // Create arrays of channels
	hitResult := make([]chan int, numDartBoards)

	for i := range xyValues { // Now set them up.
		xyValues[i] = make(chan dartCoords)
		hitResult[i] = make(chan int)
	}

	// Now start the dart board workers in parallel.
	fmt.Println("\nStart Dart Board processes...")

	//******** BEGIN TIMING HERE ***************************************************
	startTimer := time.Now()

	for i := 0; i < numDartBoards; i++ {

		go DartBoard(i, xyValues[i], hitResult[i])
	}

	// farmer/dart thrower process /////////////////////////////////////////

	for dartsThrown = 0; dartsThrown < dartsToThrow; {
	myloop:
		for i := range xyValues {
			select {
			// DART BOARD i - use a non-blocking check by using default: case
			case count := <-hitResult[i]: // get 0 or 1 from dart board
				xyValues[i] <- nextXY                           // immediately throw another dart its way
				if dartsThrown++; dartsThrown >= dartsToThrow { // SOLUTION: increment tally and check for limit
					break myloop // and break out if exceeded
				}
				nextXY.x = rand.Float64() // and generate next dart coords
				nextXY.y = rand.Float64()

				totalCount += count
			default: // if dart board not ready check next...
			}
		}
	}

	//******** END TIMING HERE ***************************************************
	elapsedTime := time.Since(startTimer)

	fmt.Println("Total number of darts thrown =", dartsThrown)
	fmt.Println("Final hit count =", totalCount)

	// do PI calculations and display/save results...
	var PiApprox = (4.0 * float64(totalCount)) / float64(dartsThrown)
	var PiError = math.Abs((Pi - PiApprox) * (100.0 / Pi))

	fmt.Println("\n\nPi approx   =", PiApprox, " using", dartsThrown, "darts")
	fmt.Println("\nPi actually =", Pi, "  Error =", PiError, "%")

	log.Println("Elapsed time = ", elapsedTime) // Main difference 'log' gives you is the time and date it logs too... handy

	n, fError = PiResults.WriteString("\nNOTE: " + msg)
	n, fError = PiResults.WriteString(fmt.Sprintf("\n\nPi approx   = %f using %d darts", PiApprox, dartsThrown)) // Sprintf formats output as string
	n, fError = PiResults.WriteString(fmt.Sprintf("\n\nPi actually = %f Error = %f%%", Pi, PiError))             // to keep WriteString happy.
	n, fError = PiResults.WriteString(fmt.Sprintf("\n\nElapsed time = %v", elapsedTime))
	n, fError = PiResults.WriteString("\n________________________________________________________________________")
	checkFileOK(fError)

} // end of main /////////////////////////////////////////////////////////////////

//----------------------------------------------------------------------------------
// Each dart board figures out if the x,y coord is a hit or not and returns 1 or 0
//----------------------------------------------------------------------------------

func DartBoard(id int, xyVal <-chan dartCoords, hitOrMiss chan<- int) {
	//fmt.Println("Dart board", id, "ready!")
	hitOrMiss <- 0 // kick off the process by sending a zero
	for {          // do forever
		Dcoords := <-xyVal // receive a dart (ow!)
		//time.Sleep(time.Duration(rand.Intn(1)) * time.Microsecond) // artificially slow down
		hitOrMiss <- 1 - int(Dcoords.x*Dcoords.x+Dcoords.y*Dcoords.y)

	}
}
