package main

import (
	"log"
	"runtime"
	"unsafe"

	"image"
	"image/draw"
	_ "image/png"
	"os"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	runtime.LockOSThread()
}

const (
	width  = 500
	height = 500
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

const baseVertexShader = `
    #version 410 core
    precision highp float;

    layout (location = 0) in vec2 aPosition;

    out vec2 vUv;
    out vec2 vL;
    out vec2 vR;
    out vec2 vT;
    out vec2 vB;
    out vec2 texelSize;

    void main () {
        vUv = aPosition * 0.5 + 0.5;
        vL = vUv - vec2(texelSize.x, 0.0);
        vR = vUv + vec2(texelSize.x, 0.0);
        vT = vUv + vec2(0.0, texelSize.y);
        vB = vUv - vec2(0.0, texelSize.y);
        gl_Position = vec4(aPosition, 0.0, 1.0);
    }
`

const curlShader = `
    precision mediump float;
    precision mediump sampler2D;

    varying highp vec2 vUv;
    varying highp vec2 vL;
    varying highp vec2 vR;
    varying highp vec2 vT;
    varying highp vec2 vB;
    uniform sampler2D uVelocity;

    void main () {
        float L = texture2D(uVelocity, vL).y;
        float R = texture2D(uVelocity, vR).y;
        float T = texture2D(uVelocity, vT).x;
        float B = texture2D(uVelocity, vB).x;
        float vorticity = R - L - T + B;
        gl_FragColor = vec4(0.5 * vorticity, 0.0, 0.0, 1.0);
    }
`

const vorticityShader = `
    precision highp float;
    precision highp sampler2D;

    varying vec2 vUv;
    varying vec2 vL;
    varying vec2 vR;
    varying vec2 vT;
    varying vec2 vB;
    uniform sampler2D uVelocity;
    uniform sampler2D uCurl;
    uniform float curl;
    uniform float dt;

    void main () {
        float L = texture2D(uCurl, vL).x;
        float R = texture2D(uCurl, vR).x;
        float T = texture2D(uCurl, vT).x;
        float B = texture2D(uCurl, vB).x;
        float C = texture2D(uCurl, vUv).x;

        vec2 force = 0.5 * vec2(abs(T) - abs(B), abs(R) - abs(L));
        force /= length(force) + 0.0001;
        force *= curl * C;
        force.y *= -1.0;

        vec2 velocity = texture2D(uVelocity, vUv).xy;
        velocity += force * dt;
        velocity = min(max(velocity, -1000.0), 1000.0);
        gl_FragColor = vec4(velocity, 0.0, 1.0);
    }
`

const divergenceShader = `
    precision mediump float;
    precision mediump sampler2D;

    varying highp vec2 vUv;
    varying highp vec2 vL;
    varying highp vec2 vR;
    varying highp vec2 vT;
    varying highp vec2 vB;
    uniform sampler2D uVelocity;

    void main () {
        float L = texture2D(uVelocity, vL).x;
        float R = texture2D(uVelocity, vR).x;
        float T = texture2D(uVelocity, vT).y;
        float B = texture2D(uVelocity, vB).y;

        vec2 C = texture2D(uVelocity, vUv).xy;
        if (vL.x < 0.0) { L = -C.x; }
        if (vR.x > 1.0) { R = -C.x; }
        if (vT.y > 1.0) { T = -C.y; }
        if (vB.y < 0.0) { B = -C.y; }

        float div = 0.5 * (R - L + T - B);
        gl_FragColor = vec4(div, 0.0, 0.0, 1.0);
    }
`

const clearShader = `
    precision mediump float;
    precision mediump sampler2D;

    varying highp vec2 vUv;
    uniform sampler2D uTexture;
    uniform float value;

    void main () {
        gl_FragColor = value * texture2D(uTexture, vUv);
    }
`

const pressureShader = `
    precision mediump float;
    precision mediump sampler2D;

    varying highp vec2 vUv;
    varying highp vec2 vL;
    varying highp vec2 vR;
    varying highp vec2 vT;
    varying highp vec2 vB;
    uniform sampler2D uPressure;
    uniform sampler2D uDivergence;

    void main () {
        float L = texture2D(uPressure, vL).x;
        float R = texture2D(uPressure, vR).x;
        float T = texture2D(uPressure, vT).x;
        float B = texture2D(uPressure, vB).x;
        float C = texture2D(uPressure, vUv).x;
        float divergence = texture2D(uDivergence, vUv).x;
        float pressure = (L + R + B + T - divergence) * 0.25;
        gl_FragColor = vec4(pressure, 0.0, 0.0, 1.0);
    }
`

const gradientSubtractShader = `
    precision mediump float;
    precision mediump sampler2D;

    varying highp vec2 vUv;
    varying highp vec2 vL;
    varying highp vec2 vR;
    varying highp vec2 vT;
    varying highp vec2 vB;
    uniform sampler2D uPressure;
    uniform sampler2D uVelocity;

    void main () {
        float L = texture2D(uPressure, vL).x;
        float R = texture2D(uPressure, vR).x;
        float T = texture2D(uPressure, vT).x;
        float B = texture2D(uPressure, vB).x;
        vec2 velocity = texture2D(uVelocity, vUv).xy;
        velocity.xy -= vec2(R - L, T - B);
        gl_FragColor = vec4(velocity, 0.0, 1.0);
    }
`

const advectionShader = `
    precision highp float;
    precision highp sampler2D;

    varying vec2 vUv;
    uniform sampler2D uVelocity;
    uniform sampler2D uSource;
    uniform vec2 texelSize;
    uniform vec2 dyeTexelSize;
    uniform float dt;
    uniform float dissipation;

    vec4 bilerp (sampler2D sam, vec2 uv, vec2 tsize) {
        vec2 st = uv / tsize - 0.5;

        vec2 iuv = floor(st);
        vec2 fuv = fract(st);

        vec4 a = texture2D(sam, (iuv + vec2(0.5, 0.5)) * tsize);
        vec4 b = texture2D(sam, (iuv + vec2(1.5, 0.5)) * tsize);
        vec4 c = texture2D(sam, (iuv + vec2(0.5, 1.5)) * tsize);
        vec4 d = texture2D(sam, (iuv + vec2(1.5, 1.5)) * tsize);

        return mix(mix(a, b, fuv.x), mix(c, d, fuv.x), fuv.y);
    }

    void main () {
    #ifdef MANUAL_FILTERING
        vec2 coord = vUv - dt * bilerp(uVelocity, vUv, texelSize).xy * texelSize;
        vec4 result = bilerp(uSource, coord, dyeTexelSize);
    #else
        vec2 coord = vUv - dt * texture2D(uVelocity, vUv).xy * texelSize;
        vec4 result = texture2D(uSource, coord);
    #endif
        float decay = 1.0 + dissipation * dt;
        gl_FragColor = result / decay;
    }
`

const colorShader = `
    precision mediump float;

    uniform vec4 color;

    void main () {
        gl_FragColor = color;
    }
`

const displayShaderSource = `
    precision highp float;
    precision highp sampler2D;

    varying vec2 vUv;
    varying vec2 vL;
    varying vec2 vR;
    varying vec2 vT;
    varying vec2 vB;
    uniform sampler2D uTexture;
    uniform sampler2D uBloom;
    uniform sampler2D uSunrays;
    uniform sampler2D uDithering;
    uniform vec2 ditherScale;
    uniform vec2 texelSize;

    vec3 linearToGamma (vec3 color) {
        color = max(color, vec3(0));
        return max(1.055 * pow(color, vec3(0.416666667)) - 0.055, vec3(0));
    }

    void main () {
        vec3 c = texture2D(uTexture, vUv).rgb;

    #ifdef SHADING
        vec3 lc = texture2D(uTexture, vL).rgb;
        vec3 rc = texture2D(uTexture, vR).rgb;
        vec3 tc = texture2D(uTexture, vT).rgb;
        vec3 bc = texture2D(uTexture, vB).rgb;

        float dx = length(rc) - length(lc);
        float dy = length(tc) - length(bc);

        vec3 n = normalize(vec3(dx, dy, length(texelSize)));
        vec3 l = vec3(0.0, 0.0, 1.0);

        float diffuse = clamp(dot(n, l) + 0.7, 0.7, 1.0);
        c *= diffuse;
    #endif

    #ifdef BLOOM
        vec3 bloom = texture2D(uBloom, vUv).rgb;
    #endif

    #ifdef SUNRAYS
        float sunrays = texture2D(uSunrays, vUv).r;
        c *= sunrays;
    #ifdef BLOOM
        bloom *= sunrays;
    #endif
    #endif

    #ifdef BLOOM
        float noise = texture2D(uDithering, vUv * ditherScale).r;
        noise = noise * 2.0 - 1.0;
        bloom += noise / 255.0;
        bloom = linearToGamma(bloom);
        c += bloom;
    #endif

        float a = max(c.r, max(c.g, c.b));
        gl_FragColor = vec4(c, a);
    }
`

// Used in adding dye and motion to simulation
const splatShader = `
    precision highp float;
    precision highp sampler2D;

    varying vec2 vUv;
    uniform sampler2D uTarget;
    uniform float aspectRatio;
    uniform vec3 color;
    uniform vec2 point;
    uniform float radius;

    void main () {
        vec2 p = vUv - point.xy;
        p.x *= aspectRatio;
        vec3 splat = exp(-dot(p, p) / radius) * color;
        vec3 base = texture2D(uTarget, vUv).xyz;
        gl_FragColor = vec4(base + splat, 1.0);
    }
`

// Create framebuffers
type framebuffers struct {
	dye        *doubleFramebuffer
	velocity   *doubleFramebuffer
	divergence *framebuffer
	curl       *framebuffer
	pressure   *framebuffer
}

type framebuffer struct {
	texture    uint32
	fbo        uint32
	width      int
	height     int
	texelSizeX float32
	texelSizeY float32
}

func (f *framebuffer) attach(id uint32) uint32 {
	gl.ActiveTexture(gl.TEXTURE0 + id)
	gl.BindTexture(gl.TEXTURE_2D, f.texture)
	return id
}

type doubleFramebuffer struct {
	width      int
	height     int
	texelSizeX float32
	texelSizeY float32
	fbo1       *framebuffer
	fbo2       *framebuffer
}

func (df *doubleFramebuffer) read() *framebuffer {
	return df.fbo1
}

func (df *doubleFramebuffer) write(f *framebuffer) {
	df.fbo2 = f
}

func (df *doubleFramebuffer) swap() {
	temp := df.fbo1
	df.fbo1 = df.fbo2
	df.fbo2 = temp
}

func initFramebuffers() *framebuffers {
	simResX, simResY := getResolution(config.SIM_RESOLUTION)
	dyeResX, dyeResY := getResolution(config.DYE_RESOLUTION)

	//log.Println(simResX, simResY, dyeResX, dyeResY)

	// Assuming we support this?
	// gl.getExtension('EXT_color_buffer_float');
	// supportLinearFiltering = gl.getExtension('OES_texture_float_linear');
	texType := uint32(gl.HALF_FLOAT)
	rgbaInt, rgba := uint32(gl.RGBA16F), uint32(gl.RGBA)
	rgInt, rg := uint32(gl.RG16F), uint32(gl.RG)
	rInt, r := uint32(gl.R16F), uint32(gl.RED)
	filtering := int32(gl.LINEAR)

	gl.Disable(gl.BLEND)

	dye := createDoubleFBO(dyeResX, dyeResY, rgbaInt, rgba, texType, filtering)
	velocity := createDoubleFBO(simResX, simResY, rgInt, rg, texType, filtering)

	divergence := createFBO(simResX, simResY, rInt, r, texType, gl.NEAREST)
	curl := createFBO(simResX, simResY, rInt, r, texType, gl.NEAREST)
	pressure := createFBO(simResX, simResY, rInt, r, texType, gl.NEAREST)

	//divergence :=
	// TODO states

	return &framebuffers{dye, velocity, divergence, curl, pressure}
}

func createFBO(w, h int, internalFormat, format, texType uint32, param int32) *framebuffer {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, param)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, param)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, int32(internalFormat), int32(w), int32(h),
		0, format, texType, gl.Ptr(nil))
	// 	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
	//      gl.TEXTURE_2D, textureColorbuffer, 0)

	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
		gl.TEXTURE_2D, texture, 0)
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// TODO check
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(gl.CheckFramebufferStatus(gl.FRAMEBUFFER))
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	}
	return &framebuffer{texture, fbo, w, h, 1.0 / float32(w), 1.0 / float32(h)}
}

func createDoubleFBO(w, h int, internalFormat, format,
	texType uint32, param int32) *doubleFramebuffer {

	fbo1 := createFBO(w, h, internalFormat, format, texType, param)
	fbo2 := createFBO(w, h, internalFormat, format, texType, param)

	return &doubleFramebuffer{w, h, fbo1.texelSizeX, fbo2.texelSizeY, fbo1, fbo2}
}

func getResolution(resolution int) (int, int) {
	aspectRatio := float32(width) / float32(height)
	if aspectRatio < 1.0 {
		aspectRatio = 1.0 / aspectRatio
	}

	min := int(resolution)
	max := int(float32(resolution) * aspectRatio)

	if width > height {
		return max, height
	}
	return min, max
}

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
	window := initGLFW("Fluid sim", width, height)
	_ = window

	newMaterial(baseVertexShader, "hey").setKeywords(nil)

	//test()
	/*
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

		// Create frame buffers

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
	*/
}

// Program to test shader and texture functions
func test() {
	// Create window
	window := initGLFW("Fluid sim", width, height)

	// Create buffers
	vertices := []float32{
		//Positions      // Colors       // Texture coords
		0.5, 0.5, 0.0, 1.0, 0.0, 0.0, 1.0, 1.0, // Top right
		0.5, -0.5, 0.0, 0.0, 1.0, 0.0, 1.0, 0.0, // Bottom right
		-0.5, -0.5, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, // Bottom left
		-0.5, 0.5, 0.0, 1.0, 1.0, 0.0, 0.0, 1.0, // Top left
	}
	indices := []uint32{
		0, 1, 3, // First triangle
		1, 2, 3, // Second triangle
	}
	var VAO, VBO, EBO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.GenBuffers(1, &EBO)
	gl.BindVertexArray(VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices),
		gl.STATIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4,
		gl.Ptr(indices), gl.STATIC_DRAW)
	// Specify our position attributes
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Specify our color attributes
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 8*4,
		gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)
	// Texture coord attributes
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8*4,
		gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(2)
	// Unbind our vertex array so we don't mess with it later
	gl.BindVertexArray(0)

	// Create material
	displayMaterial := newMaterial(`
#version 410 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aColor;
layout (location = 2) in vec2 aTexCoord;

out vec3 ourColor;
out vec2 TexCoord;

void main()
{
	gl_Position = vec4(aPos, 1.0);
	ourColor = aColor;
 	TexCoord = vec2(aTexCoord.x, aTexCoord.y);
}`, `
#version 410 core
out vec4 FragColor;

in vec3 ourColor;
in vec2 TexCoord;

uniform sampler2D texture1;

void main() 
{
    FragColor = texture(texture1, TexCoord);
}`)
	displayMaterial.setKeywords([]string{})

	// Load in dithering texture
	ditheringTexture := createTexture("LDR_LLL1_0.png")

	// Create frame buffers
	// NOTE we do get a weird issue with the texture being rendered small in the
	// bottom left corner however I think this might be a view port issue.
	// When the window resizies this issue corrects itself.
	// It is a viewport issue.
	initFramebuffers()
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	for !window.ShouldClose() {
		gl.ClearColor(0.3, 0.5, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		//gl.BindTexture(gl.TEXTURE_2D, ditheringTexture.attach(0))

		ditheringTexture.attach(0)
		displayMaterial.bind()
		gl.BindVertexArray(VAO)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		gl.BindVertexArray(0)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
