### Building

Install needed deps
```
go get -u github.com/go-gl/mathgl/mgl32
go get -u github.com/go-gl/gl/v4.1-core/gl
go get -u github.com/go-gl/glfw/v3.3/glfw
```

```
sudo apt-get install evtest
```

Find your touchpad by doing
```
sudo evtest
```

Change the number 8 in the file main.go to whatever number your touchpad input is
```
go readTouchPad()
```

Then run
```
go build main.go && sudo ./main
```

It has to be run with sudo unfortunately due to using evtest to read touchpad absolute positions.

### Desktop version of this great website

https://paveldogreat.github.io/WebGL-Fluid-Simulation/  

https://github.com/PavelDoGreat/WebGL-Fluid-Simulation

Made by https://twitter.com/PavelDoGreat