// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/dark"

	"image"
	"log"
	"os"
	"github.com/kyokomi/gomajan/pai"
	"github.com/kyokomi/gomajan/taku"
)

type MainWindow struct {
	gxui.Window
	theme gxui.Theme
	driver gxui.Driver
}

func NewMainWindow(driver gxui.Driver) MainWindow {
	var mainWindow MainWindow
	theme := dark.CreateTheme(driver)
	window := theme.CreateWindow(800, 600, "Polygon")

	mainWindow.Window = window
	mainWindow.theme = theme
	mainWindow.driver = driver

	return mainWindow
}

func appMain(driver gxui.Driver) {

	window := NewMainWindow(driver)

	taku := taku.NewTaku()

	containerBase := window.Theme().CreateLinearLayout()
	containerBase.SetDirection(gxui.TopToBottom)
	containerBase.SetVerticalAlignment(gxui.AlignTop)
	for _, player := range taku.Players {
		container := window.Theme().CreateLinearLayout()
		container.SetDirection(gxui.LeftToRight)
		container.SetHorizontalAlignment(gxui.AlignCenter)
		for _, hai := range player.Tiles() {
			if hai.Pai.Type() == pai.NoneType || hai.Val <= 0 {
				continue
			}
			for i := 0; i < hai.Val; i++ {
				container.AddChild(window.createPaiImage(hai.Pai))
			}
		}
		containerBase.AddChild(container)
	}

//	label := window.Theme().CreateLabel()
//	label.SetText("hogehogehoge")
//	label.SetMargin(math.CreateSpacing(200))
//	container.AddChild(label)
	window.AddChild(containerBase)
	window.OnClose(driver.Terminate)
	gxui.EventLoop(driver)
}

func (m MainWindow) Theme() gxui.Theme {
	return m.theme
}

func (m MainWindow) createPaiImage(p pai.MJP) gxui.Image {
	f, err := os.Open("data/pai-images/" + imageMappings[p])
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	im, _, err := image.Decode(f)
	if err != nil {
		log.Fatalln(err)
	}

	pict := m.theme.CreateImage()
	texture := m.driver.CreateTexture(im, 96)
	pict.SetTexture(texture)
	pict.SetExplicitSize(math.Size{32, 45})
	//	pict.SetMargin(math.CreateSpacing(4))

	return pict
}

func main() {
	flag.Parse()
	gl.StartDriver(appMain)
}
