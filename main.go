package main

import (
	"log"
	"unsafe"

	"image"
	"image/draw"
	_ "image/png"
	"os"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// We will move everything to its own module later on
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

// Material
type material struct {
	vertexSource   string
	fragmentSource string
	activeProgram  *Shader
	uniforms       []string
	origFragSource string // Source stored with no flags enabled
}

func newMaterial(vsSource, fsSource string) *material {
	return &material{vsSource, fsSource, nil, []string{}, fsSource}
}

func (m *material) setKeywords(keywords []string) {
	// Simplying here because we are gonna assume we arent gonna change params
	// much if all at compile time so we can take the preformance hit of
	// recompling programs.
	m.fragmentSource = m.origFragSource
	for _, keyword := range keywords {
		m.fragmentSource = "#define " + keyword + "\n" + m.fragmentSource
	}

	m.activeProgram = MakeShaders(m.vertexSource, m.fragmentSource)
	// TODO uniforms
}

func (m *material) bind() {
	m.activeProgram.Use()
}

type Shader struct {
	ID uint32
}

func MakeShaders(vertexCode, fragmentCode string) *Shader {
	// Compile the shaders
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	shaderSource, freeVertex := gl.Strs(vertexCode + "\x00")
	defer freeVertex()
	gl.ShaderSource(vertexShader, 1, shaderSource, nil)
	gl.CompileShader(vertexShader)
	checkCompileErrors(vertexShader, "VERTEX")

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	shaderSource, freeFragment := gl.Strs(fragmentCode + "\x00")
	defer freeFragment()
	gl.ShaderSource(fragmentShader, 1, shaderSource, nil)
	gl.CompileShader(fragmentShader)
	checkCompileErrors(fragmentShader, "FRAGMENT")

	// Create a shader program
	ID := gl.CreateProgram()
	gl.AttachShader(ID, vertexShader)
	gl.AttachShader(ID, fragmentShader)
	gl.LinkProgram(ID)

	checkCompileErrors(ID, "PROGRAM")

	// Delete shaders
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return &Shader{ID: ID}
}

func checkCompileErrors(shader uint32, shaderType string) {
	var success int32
	var infoLog [1024]byte

	var status uint32 = gl.COMPILE_STATUS
	stageMessage := "Shader_Compilation_error"
	errorFunc := gl.GetShaderInfoLog
	getIV := gl.GetShaderiv
	if shaderType == "PROGRAM" {
		status = gl.LINK_STATUS
		stageMessage = "Program_link_error"
		errorFunc = gl.GetProgramInfoLog
		getIV = gl.GetProgramiv
	}

	getIV(shader, status, &success)
	if success != 1 {
		test := &success
		errorFunc(shader, 1024, test, (*uint8)(unsafe.Pointer(&infoLog)))
		log.Fatalln(stageMessage + shaderType + "|" + string(infoLog[:1024]) + "|")
	}
}

func (s Shader) Use() {
	gl.UseProgram(s.ID)
}

// Create framebuffers

// Load in dithering texture
type texture struct {
	texture uint32
	width   int32
	height  int32
}

func (t *texture) attach(id uint32) uint32 {
	gl.ActiveTexture(gl.TEXTURE0 + id)
	gl.BindTexture(gl.TEXTURE_2D, t.texture)
	return id
}

func createTexture(path string) *texture {
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	data := image.NewRGBA(img.Bounds())
	if data.Stride != data.Rect.Size().X*4 {
		panic("Unsupported stride")
	}
	draw.Draw(data, data.Bounds(), img, image.Point{0, 0}, draw.Src)

	width, height := int32(data.Rect.Size().X), int32(data.Rect.Size().Y)
	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		width,
		height,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(data.Pix))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	// Set texture parameters for wrapping
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER,
		gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return &texture{textureID, width, height}
}

// Render function

// Step function

// Run simulation
func main() {
	// Create window
	window := initGLFW("Fluid sim", 500, 500)

	// Create shaders
	// TODO

	// Create material
	displayMaterial := newMaterial("place", "place")
	log.Println(displayMaterial)

	// Load in dithering texture
	ditheringTexture := createTexture("LDR_LLL1_0.png")
	log.Println(ditheringTexture)

	for !window.ShouldClose() {
		gl.ClearColor(0.3, 0.5, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// update parms

		// Wait / deal with dt

		// Step func

		// Render

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
