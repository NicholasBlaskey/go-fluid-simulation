package main

import (
	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Create window
func initGLFW(windowTitle string, width, height int) *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(
		width, height, windowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	window.SetFramebufferSizeCallback(
		glfw.FramebufferSizeCallback(framebuffer_size_callback))
	window.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	return window
}

func framebuffer_size_callback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {

	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
}

// Set parameters (Perhaps better to var this but will keep same name for now)
// Actually move to its own module later on
var config = struct {
	SIM_RESOLUTION       int
	DYE_RESOLUTION       int
	CAPTURE_RESOLUTION   int
	DENSITY_DISSIPATION  int
	VELOCITY_DISSIPATION float32
	PRESSURE             float32
	PRESSURE_ITERATIONS  int
	CURL                 int
	SPLAT_RADIUS         float32
	SPLAT_FORCE          int
	SHADING              bool
	COLORFUL             bool
	COLOR_UPDATE_SPEED   int
	PAUSED               bool
	BACK_COLOR           mgl.Vec3
	TRANSPARENT          bool
	BLOOM                bool
	BLOOM_ITERATIONS     int
	BLOOM_RESOLUTION     int
	BLOOM_INTENSITY      float32
	BLOOM_THRESHOLD      float32
	BLOOM_SOFT_KNEE      float32
	SUNRAYS              bool
	SUNRAYS_RESOLUTION   int
	SUNRAYS_WEIGHT       float32
}{
	SIM_RESOLUTION:       128,
	DYE_RESOLUTION:       1024,
	CAPTURE_RESOLUTION:   512,
	DENSITY_DISSIPATION:  1,
	VELOCITY_DISSIPATION: 0.2,
	PRESSURE:             0.8,
	PRESSURE_ITERATIONS:  20,
	CURL:                 30,
	SPLAT_RADIUS:         0.25,
	SPLAT_FORCE:          6000,
	SHADING:              true,
	COLORFUL:             true,
	COLOR_UPDATE_SPEED:   10,
	PAUSED:               false,
	BACK_COLOR:           mgl.Vec3{0, 0, 0},
	TRANSPARENT:          false,
	BLOOM:                true,
	BLOOM_ITERATIONS:     8,
	BLOOM_RESOLUTION:     256,
	BLOOM_INTENSITY:      0.8,
	BLOOM_THRESHOLD:      0.6,
	BLOOM_SOFT_KNEE:      0.7,
	SUNRAYS:              true,
	SUNRAYS_RESOLUTION:   196,
	SUNRAYS_WEIGHT:       1.0,
}

// Pointer prototype?

// Material

// Create shaders (IGNORE keywords for now)

// Create framebuffers

// Load in dithering texture

// Render function

// Step function

// Run simulation
func main() {
	// Create window
	window := initGLFW("Fluid sim", 500, 500)

	for !window.ShouldClose() {
		gl.ClearColor(0.3, 0.5, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
