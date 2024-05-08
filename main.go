package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"
)

type cell struct {
	isBomb    bool
	isFlagged bool

	isCleared bool
	value     int
	sweeped   bool

	cellRow int
	cellCol int
}

func getUserInput(c chan string, quit <-chan bool) {
	for {
		select {
		case <-quit:
			// close(c)
			return
		default:
			var b []byte = make([]byte, 1)
			os.Stdin.Read(b)
			trimmedInput := strings.TrimSpace(string(b))
			c <- trimmedInput
		}
	}
}

func main() {
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

	// Clear whole screen
	fmt.Print("\033[H\033[2J")

	// Hide the cursor
	fmt.Print("\033[?25l")

	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	numRows, numCols, numMines := 8, 8, 10

	gameBoard := [][]cell{}

	for i := 0; i < numRows; i++ {
		tempGameRow := []cell{}
		for j := 0; j < numCols; j++ {
			tempGameRow = append(tempGameRow, cell{false, false, false, 0, false, i, j})
		}
		gameBoard = append(gameBoard, tempGameRow)
	}

	cursorRow, cursorCol := 0, 0
	pressedABomb := false
	numFlagged := 0

	// initializeMines(gameBoard, numMines)

	userInputChannel := make(chan string)
	quitUserInputChannel := make(chan bool)
	cursorOn := true

	go getUserInput(userInputChannel, quitUserInputChannel)

	firstPress := true
	for {
		select {
		case inp := <-userInputChannel:
			switch inp {
			case "w", "W":
				if cursorRow > 0 {
					cursorRow -= 1
				}
			case "a", "A":
				if cursorCol > 0 {
					cursorCol -= 1
				}
			case "s", "S":
				if cursorRow < numRows-1 {
					cursorRow += 1
				}
			case "d", "D":
				if cursorCol < numCols-1 {
					cursorCol += 1
				}
			case "":

				if firstPress {
					initializeMines(gameBoard, numMines, cursorRow, cursorCol)
					firstPress = false
				}

				if gameBoard[cursorRow][cursorCol].isBomb {
					pressedABomb = true

					quitUserInputChannel <- true
					fmt.Print("\033[H\033[2J")
					fmt.Println("Oh no!! You hit a mine :(")
					fmt.Printf("%v", display(formatBoard(gameBoard, cursorRow, cursorCol, pressedABomb, cursorOn), numMines, numFlagged, gameOverTextList))
					fmt.Print("\033[?25h")
					exec.Command("stty", "-F", "/dev/tty", "echo").Run()
					os.Exit(0)
				} else {
					expandPossibleCells(gameBoard, cursorRow, cursorCol)
				}
			case "f":
				gameBoard[cursorRow][cursorCol].isFlagged = !gameBoard[cursorRow][cursorCol].isFlagged
				if gameBoard[cursorRow][cursorCol].isFlagged {
					numFlagged += 1
				} else {
					numFlagged -= 1
				}
			case "e", "q":
				quitUserInputChannel <- true
				fmt.Print("\033[?25h")
				exec.Command("stty", "-F", "/dev/tty", "echo").Run()
				os.Exit(0)
			default:
				// fmt.Println("None :(")
			}
		default:
			cursorOn = !cursorOn
			time.Sleep(time.Millisecond * 100)
		}

		// Clear whole screen
		fmt.Print("\033[H\033[2J")

		if isFullBoardCleared(gameBoard) {
			quitUserInputChannel <- true
			fmt.Println("Congratulations!! You have sweeped the entire field without hitting a mine!")
			fmt.Printf("%v", display(formatBoard(gameBoard, cursorRow, cursorCol, pressedABomb, cursorOn), numMines, numFlagged, youWinTextList))
			fmt.Print("\033[?25h")
			exec.Command("stty", "-F", "/dev/tty", "echo").Run()
			os.Exit(0)
		} else {
			fmt.Printf("board: \n%v", display(formatBoard(gameBoard, cursorRow, cursorCol, pressedABomb, cursorOn), numMines, numFlagged, normalSpaceTextList))
		}

	}
}

var normalSpaceText string = "\n\n\n\n\n\n\n"

var gameOverText string = `
 ██████╗  █████╗ ███╗   ███╗███████╗     ██████╗ ██╗   ██╗███████╗██████╗ 
██╔════╝ ██╔══██╗████╗ ████║██╔════╝    ██╔═══██╗██║   ██║██╔════╝██╔══██╗
██║  ███╗███████║██╔████╔██║█████╗      ██║   ██║██║   ██║█████╗  ██████╔╝
██║   ██║██╔══██║██║╚██╔╝██║██╔══╝      ██║   ██║╚██╗ ██╔╝██╔══╝  ██╔══██╗
╚██████╔╝██║  ██║██║ ╚═╝ ██║███████╗    ╚██████╔╝ ╚████╔╝ ███████╗██║  ██║
 ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝     ╚═════╝   ╚═══╝  ╚══════╝╚═╝  ╚═╝
 `
var youWinText string = `
██╗   ██╗ ██████╗ ██╗   ██╗    ██╗    ██╗██╗███╗   ██╗██╗
╚██╗ ██╔╝██╔═══██╗██║   ██║    ██║    ██║██║████╗  ██║██║
 ╚████╔╝ ██║   ██║██║   ██║    ██║ █╗ ██║██║██╔██╗ ██║██║
  ╚██╔╝  ██║   ██║██║   ██║    ██║███╗██║██║██║╚██╗██║╚═╝
   ██║   ╚██████╔╝╚██████╔╝    ╚███╔███╔╝██║██║ ╚████║██╗
   ╚═╝    ╚═════╝  ╚═════╝      ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝
`

var normalSpaceTextList = strings.Split(normalSpaceText, "\n")
var gameOverTextList = strings.Split(gameOverText, "\n")
var youWinTextList = strings.Split(youWinText, "\n")

func centerString(s string, requiredLen int) string {
	currLen := utf8.RuneCountInString(s)
	paddingOnEachSide := (requiredLen - currLen) / 2
	paddingString := strings.Repeat(" ", paddingOnEachSide)

	return paddingString + s + paddingString
}

func display(formattedString string, numMines, numFlags int, centreTextList []string) string {
	differentLines := strings.Split(formattedString, "\n")
	numLines := len(differentLines) - 1
	halfWay, menuStartLine, menuEndLine := numLines/2, numLines/4, (numLines/2)+(numLines/4)

	// fmt.Println(numLines, halfWay, menuStartLine, menuEndLine, 10/4)

	// fmt.Println(utf8.RuneCountInString(centreTextList[2]))

	displayStr := ""
	for i, line := range differentLines {
		displayStr += line

		if i-1 > 0 && i < len(centreTextList) && i > 1 {
			displayStr += "\t" + centerString(centreTextList[i-1], 74)
		}

		if i >= menuStartLine && i <= menuEndLine {
			if i == menuStartLine {
				displayStr += "\t W,A,S,D: Move Up, Left, Down & Right"
			} else if i == halfWay-1 {
				displayStr += "\t F: Flag a cell"
			} else if i == halfWay {
				displayStr += "\t Flags Left: " + fmt.Sprint(numMines-numFlags) + "/" + fmt.Sprint(numMines)
			} else if i == menuEndLine {
				displayStr += "\t Q, E: Quit / Exit the Program"
			}
		}

		// fmt.Printf("len(gameOverTextList): %v\n", len(gameOverTextList))

		displayStr += "\n"
	}

	return displayStr
}

func isFullBoardCleared(gameBoard [][]cell) bool {
	for _, row := range gameBoard {
		for _, cell := range row {
			if !cell.sweeped && !cell.isBomb {
				return false
			}
		}
	}
	return true
}

func expandPossibleCells(gameBoard [][]cell, cursorRow, cursorCol int) {
	numRows, numCols := len(gameBoard), len(gameBoard[0])

	bfsQueue := []*cell{&gameBoard[cursorRow][cursorCol]}
	for len(bfsQueue) > 0 {

		currCell := bfsQueue[0]
		bfsQueue = bfsQueue[1:]
		currRow, currCol := currCell.cellRow, currCell.cellCol

		if gameBoard[currRow][currCol].isCleared {
			continue
		}

		gameBoard[currRow][currCol].isCleared = true
		gameBoard[currRow][currCol].sweeped = true

		tempPossibleAdditions := []*cell{}

		if currRow > 0 {
			if !gameBoard[currRow-1][currCol].isCleared && !gameBoard[currRow-1][currCol].isBomb {
				tempPossibleAdditions = append(tempPossibleAdditions, &gameBoard[currRow-1][currCol])
			} else if gameBoard[currRow-1][currCol].isBomb {
				currCell.value += 1
			}
		}
		if currCol > 0 {
			if !gameBoard[currRow][currCol-1].isCleared && !gameBoard[currRow][currCol-1].isBomb {
				tempPossibleAdditions = append(tempPossibleAdditions, &gameBoard[currRow][currCol-1])
			} else if gameBoard[currRow][currCol-1].isBomb {
				currCell.value += 1
			}
		}
		if currRow < numRows-1 {
			if !gameBoard[currRow+1][currCol].isCleared && !gameBoard[currRow+1][currCol].isBomb {
				tempPossibleAdditions = append(tempPossibleAdditions, &gameBoard[currRow+1][currCol])
			} else if gameBoard[currRow+1][currCol].isBomb {
				currCell.value += 1
			}
		}
		if currCol < numCols-1 {
			if !gameBoard[currRow][currCol+1].isCleared && !gameBoard[currRow][currCol+1].isBomb {
				tempPossibleAdditions = append(tempPossibleAdditions, &gameBoard[currRow][currCol+1])
			} else if gameBoard[currRow][currCol+1].isBomb {
				currCell.value += 1
			}
		}
		if currRow > 0 && currCol > 0 {
			if !gameBoard[currRow-1][currCol-1].isCleared && !gameBoard[currRow-1][currCol-1].isBomb {
				tempPossibleAdditions = append(tempPossibleAdditions, &gameBoard[currRow-1][currCol-1])
			} else if gameBoard[currRow-1][currCol-1].isBomb {
				currCell.value += 1
			}
		}
		if currRow > 0 && currCol < numCols-1 {
			if !gameBoard[currRow-1][currCol+1].isCleared && !gameBoard[currRow-1][currCol+1].isBomb {
				tempPossibleAdditions = append(tempPossibleAdditions, &gameBoard[currRow-1][currCol+1])
			} else if gameBoard[currRow-1][currCol+1].isBomb {
				currCell.value += 1
			}
		}
		if currRow < numRows-1 && currCol > 0 {
			if !gameBoard[currRow+1][currCol-1].isCleared && !gameBoard[currRow+1][currCol-1].isBomb {
				tempPossibleAdditions = append(tempPossibleAdditions, &gameBoard[currRow+1][currCol-1])
			} else if gameBoard[currRow+1][currCol-1].isBomb {
				currCell.value += 1
			}
		}
		if currRow < numRows-1 && currCol < numCols-1 {
			if !gameBoard[currRow+1][currCol+1].isCleared && !gameBoard[currRow+1][currCol+1].isBomb {
				tempPossibleAdditions = append(tempPossibleAdditions, &gameBoard[currRow+1][currCol+1])
			} else if gameBoard[currRow+1][currCol+1].isBomb {
				currCell.value += 1
			}
		}

		if currCell.value == 0 {
			bfsQueue = append(bfsQueue, tempPossibleAdditions...)
		}

		// fmt.Println(currCell.cellRow, currCell.cellCol, currCell.value)
		// fmt.Println(gameBoard[currRow][currCol].value)

	}
}

func initializeMines(gameBoard [][]cell, numMines int, selectedRow, selectedCol int) {
	numRows, numCols := len(gameBoard), len(gameBoard[0])
	for numMines > 0 {
		tempRow, tempCol := rand.Intn(numRows), rand.Intn(numCols)
		if !(tempRow == selectedRow && tempCol == selectedCol) && !gameBoard[tempRow][tempCol].isBomb {
			gameBoard[tempRow][tempCol].isBomb = true
			numMines--
		}
	}
	// fmt.Println(gameBoard)
}

func formatBoard(gameBoard [][]cell, cursorRow, cursorCol int, pressedABomb bool, cursorOn bool) string {
	numRows, numCols := len(gameBoard), len(gameBoard[0])

	printString := ""

	for i := 0; i < numCols+numCols/2; i++ {
		printString += string('━') + string('━')
	}
	printString += string('━') + "\n"

	for i := 0; i < numRows; i++ {
		printString += string('‖')
		for j := 0; j < numCols; j++ {
			if cursorRow == i && cursorCol == j {
				if pressedABomb {
					printString += string('❌')
				} else {
					if isFullBoardCleared(gameBoard) {
						printString += convertToDisplayCharacters(gameBoard[i][j], pressedABomb)
					} else if cursorOn {
						printString += string('█') + string('█')
					} else {
						printString += convertToDisplayCharacters(gameBoard[i][j], pressedABomb)
					}
				}
			} else {
				printString += convertToDisplayCharacters(gameBoard[i][j], pressedABomb)
			}

			if j == numCols-1 {
				printString += string('‖')
			} else {
				printString += "|"
			}
		}
		printString += "\n"
	}

	for i := 0; i < numCols+numCols/2; i++ {
		printString += string('━') + string('━')
	}
	printString += string('━') + "\n"

	return printString
}

func convertToDisplayCharacters(c cell, showBombs bool) string {

	if c.isFlagged {
		return string('🚩')
	}

	if c.isBomb {
		if showBombs {
			return string('💣')
		} else {
			return "  "
		}
	}

	if !c.isCleared {
		return string(' ') + " "
	}

	switch c.value {
	case -1:
		return string(' ') + " "
	case 0:
		return string('0') + " "
	case -2:
		return string('🚩')
	case 1, 2, 3, 4, 5, 6, 7, 8:
		return fmt.Sprint(c.value) + " "
	default:
		return string(rune(c.value)) + " "
	}
}
