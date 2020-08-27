package main

/*
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
*/
