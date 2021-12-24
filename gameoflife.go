package main

import (
	"GameOfLife-GL/glrender"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	height             = 600
	width              = 600
	threshhold         = 0.33
	frameRate          = 16
	vertexShaderSource = `
    #version 410 core
    in vec3 v;
    void main() {
        gl_Position = vec4(v, 1.0);
    }
` + "\x00"

	fragmentShaderSource = `
    #version 410 core
    out vec4 frag_colour;
    void main() {
        frag_colour = vec4(1, 1, 1, 1);
    }
` + "\x00"
)

var (
	board_rows int64 = 100
	board_cols int64 = 100
	//direction vectors (x,y pairs on the same index for all neighbors)
	dx     = [8]int{0, 1, 1, 1, 0, -1, -1, -1}
	dy     = [8]int{-1, -1, 0, 1, 1, 1, 0, -1}
	square = []float32{
		// lower-left triangle
		-0.5, 0.5, 0, // top
		-0.5, -0.5, 0, // left
		0.5, -0.5, 0, // right

		// upper-right triangle
		-0.5, 0.5, 0, // left
		0.5, 0.5, 0, // right
		0.5, -0.5, 0, // bottom
	}
)

type cell struct {
	drawable uint32
	state    bool
	next     bool
	x        int
	y        int
}

func init() {
	runtime.LockOSThread()
}

func main() {
	if len(os.Args) == 3 {
		board_cols, _ = strconv.ParseInt(os.Args[1], 10, 0)
		board_rows, _ = strconv.ParseInt(os.Args[2], 10, 0)
	} else if len(os.Args) != 1 {
		println("Usage:\tGameOfLife-GL <number of columns> <number of rows>\n\tpass no arguments for a default 100x100 board.")
		return
	}

	window := glrender.InitGlfw(width, height)
	defer window.Destroy()

	program := glrender.InitOpenGL(vertexShaderSource, fragmentShaderSource)
	board := generateBoard()

	for !window.ShouldClose() {
		t := time.Now()
		renderBoard(board, window, program)
		cellState(board)
		time.Sleep(time.Second/time.Duration(frameRate) - time.Since(t))
	}
}

func newCell(x, y int) *cell {
	v := make([]float32, len(square))
	copy(v, square)

	for i := 0; i < len(v); i++ {
		var position float32
		var size float32
		switch i % 3 {
		case 0:
			size = 1.0 / float32(board_cols)
			position = float32(x) * size
		case 1:
			size = 1.0 / float32(board_rows)
			position = float32(y) * size
			// skip z coord
		default:
			continue
		}

		if v[i] < 0 {
			v[i] = (position * 2) - 1
		} else {
			v[i] = ((position + size) * 2) - 1
		}
	}

	return &cell{
		drawable: glrender.MakeVao(v),
		x:        x,
		y:        y,
	}
}

// Create the entire board such that each cell is in a random state.
func generateBoard() [][]*cell {
	//var board [][]byte = make([][]byte, board_cols)
	board := make([][]*cell, board_cols)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < len(board); i++ {
		board[i] = make([]*cell, board_rows)

		for j := 0; j < len(board[0]); j++ {
			board[i][j] = newCell(i, j)
			board[i][j].state = rand.Float32() < threshhold
			board[i][j].next = board[i][j].state

		}
	}

	return board
}

func (c *cell) draw() {
	if !c.state {
		return
	}

	gl.BindVertexArray(c.drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
}

func renderBoard(board [][]*cell, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	for x := range board {
		for _, c := range board[x] {
			c.draw()
		}
	}

	glfw.PollEvents()
	window.SwapBuffers()
}

func cellState(b [][]*cell) {
	for x := range b {
		for y := range b[x] {
			sum := 0

			for i := 0; i < len(dx); i++ {
				nx := dx[i] + x
				ny := dy[i] + y

				if nx < 0 {
					nx = len(b) - 1
				} else if nx >= len(b) {
					nx = 0
				}

				if ny < 0 {
					ny = len(b[0]) - 1
				} else if ny >= len(b[0]) {
					ny = 0
				}

				if b[nx][ny].state {
					sum += 1
				}
			}

			if sum == 3 || (b[x][y].state && sum == 2) {
				b[x][y].next = true
			} else {
				b[x][y].next = false
			}
		}
	}

	updateBoard(b)
}

func updateBoard(b [][]*cell) {
	for x := range b {
		for y := range b[x] {
			b[x][y].state = b[x][y].next
		}
	}
}
