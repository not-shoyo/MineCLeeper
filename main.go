package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
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

func main() {
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

	// Clear whole screen
	fmt.Print("\033[H\033[2J")

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

	// initializeMines(gameBoard, numMines)

	fmt.Print("\033[?25l")

	firstPress := true

	for {

		fmt.Printf("board after move(%v, %v): \n%v", cursorRow, cursorCol, formatBoard(gameBoard, cursorRow, cursorCol, pressedABomb))

		var b []byte = make([]byte, 1)
		os.Stdin.Read(b)

		trimmedInput := strings.TrimSpace(string(b))

		switch trimmedInput {
		case "w":
			if cursorRow > 0 {
				cursorRow -= 1
			}
		case "a":
			if cursorCol > 0 {
				cursorCol -= 1
			}
		case "s":
			if cursorRow < numRows-1 {
				cursorRow += 1
			}
		case "d":
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

				fmt.Print("\033[H\033[2J")
				fmt.Println("Oh no!! You hit a mine :(")
				fmt.Printf("%v", formatBoard(gameBoard, cursorRow, cursorCol, pressedABomb))
				fmt.Print("\033[?25h")
				exec.Command("stty", "-F", "/dev/tty", "echo").Run()
				os.Exit(0)
			} else {
				expandPossibleCells(gameBoard, cursorRow, cursorCol)
			}
		case "f":
			gameBoard[cursorRow][cursorCol].isFlagged = !gameBoard[cursorRow][cursorCol].isFlagged
		case "e", "q":
			fmt.Print("\033[?25h")
			exec.Command("stty", "-F", "/dev/tty", "echo").Run()
			os.Exit(0)
		default:
			// fmt.Println("None :(")
		}

		fmt.Printf("board after move(%v, %v): \n%v", cursorRow, cursorCol, formatBoard(gameBoard, cursorRow, cursorCol, pressedABomb))

		fmt.Print("\033[H\033[2J")

		if isFullBoardCleared(gameBoard) {
			fmt.Println("Congratulations!! You have sweeped the entire field without hitting a mine!")
			fmt.Printf("board:\n%v", formatBoard(gameBoard, cursorRow, cursorCol, pressedABomb))
			fmt.Print("\033[?25h")
			exec.Command("stty", "-F", "/dev/tty", "echo").Run()
			os.Exit(0)
		}

	}
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

func formatBoard(gameBoard [][]cell, cursorRow, cursorCol int, pressedABomb bool) string {
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
					printString += string('█') + string('█')
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
