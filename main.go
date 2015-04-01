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
	"github.com/kyokomi/gomajan/player"
	"github.com/kyokomi/gomajan/taku"
)

type MainWindow struct {
	gxui.Window
	theme  gxui.Theme
	driver gxui.Driver
}

func NewMainWindow(driver gxui.Driver) MainWindow {
	var mainWindow MainWindow
	theme := dark.CreateTheme(driver)
	window := theme.CreateWindow(800, 800, "gxjan")

	mainWindow.Window = window
	mainWindow.theme = theme
	mainWindow.driver = driver

	return mainWindow
}

var _taku *taku.Taku
var _yamaImage [4][2][17]*PaiImage

func appMain(driver gxui.Driver) {

	window := NewMainWindow(driver)

	_taku = taku.NewTaku()

	// TODO: Splitterにしてみた
	rootLayer := window.Theme().CreateSplitterLayout()
//	rootLayer.SetOrientation(gxui.TopToBottom)
//	rootLayer.SetVerticalAlignment(gxui.AlignTop)

	// 手牌
	{
		tehaiLayer := window.Theme().CreateLinearLayout()
		tehaiLayer.SetDirection(gxui.TopToBottom)
		tehaiLayer.SetVerticalAlignment(gxui.AlignTop)
		tehaiLayer.SetMargin(math.CreateSpacing(20))

		for _, player := range _taku.Players {
			container := window.Theme().CreateLinearLayout()
			container.SetDirection(gxui.LeftToRight)
			container.SetHorizontalAlignment(gxui.AlignCenter)

			for _, hai := range player.Tiles() {
				if hai.Pai.Type() == pai.NoneType || hai.Val <= 0 {
					continue
				}

				for i := 0; i < hai.Val; i++ {
					window.nextPaiImage(container, &player, hai.Pai)
				}
			}
			container.SetMargin(math.CreateSpacing(10))
			tehaiLayer.AddChild(container)
		}
		rootLayer.AddChild(tehaiLayer)
		rootLayer.SetChildWeight(tehaiLayer, 0.4)
	}

	// 山
	{
		yamaLayer := window.Theme().CreateLinearLayout()
		yamaLayer.SetDirection(gxui.TopToBottom)
		yamaLayer.SetVerticalAlignment(gxui.AlignTop)
		yamaLayer.SetMargin(math.CreateSpacing(20))

		for i, 山2 := range _taku.Yama {
			container := window.Theme().CreateLinearLayout()
			container.SetDirection(gxui.TopToBottom)
			container.SetVerticalAlignment(gxui.AlignTop)
			for j, 山 := range 山2 {
				container2 := window.Theme().CreateLinearLayout()
				container2.SetDirection(gxui.LeftToRight)
				container2.SetHorizontalAlignment(gxui.AlignCenter)
				for k, 牌 := range 山 {
					if 牌.Type() == pai.NoneType {
						continue
					}

					imagePai := window.createPaiImage(牌)
					if _taku.YamaMask[i][j][k] != 0 {
						imagePai.SetVisible(false)
					}
					_yamaImage[i][j][k] = &imagePai
					container2.AddChild(imagePai)
				}
				container2.SetMargin(math.CreateSpacing(1))
				container.AddChild(container2)
			}
			container.SetMargin(math.CreateSpacing(4))
			yamaLayer.AddChild(container)
		}
		rootLayer.AddChild(yamaLayer)
		rootLayer.SetChildWeight(yamaLayer, 0.6)
	}

	//	label := window.Theme().CreateLabel()
	//	label.SetText("hogehogehoge")
	//	label.SetMargin(math.CreateSpacing(200))
	//	container.AddChild(label)
	window.AddChild(rootLayer)
	window.OnClose(driver.Terminate)
	gxui.EventLoop(driver)
}

func (m MainWindow) Theme() gxui.Theme {
	return m.theme
}

type PaiImage struct {
	gxui.Image
	Pai pai.MJP
}

func (m MainWindow) createPaiImage(p pai.MJP) PaiImage {
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

	return PaiImage{Image: pict, Pai: p}
}

var ueshita = 1 // ちょんちょん後の下段からなので1

func nextFunc(playerID int) pai.MJP {
	if ueshita == 2 {
		_taku.Add列And山越しFunc(1)
		ueshita = 0
	}

	p := _taku.Next(ueshita, playerID)

	// 引いた牌を非表示
	_yamaImage[_taku.PlayYamaIndex][ueshita][_taku.PlayRetuIndex].SetVisible(false)

	ueshita++
	return p
}

func (m MainWindow) nextPaiImage(container gxui.LinearLayout, player *player.Player, p pai.MJP) {
	paiImage := m.createPaiImage(p)
	paiImage.OnClick(func(e gxui.MouseEvent) {
		player.PaiDec(paiImage.Pai)
		container.RemoveChild(paiImage)

		nextPai := nextFunc(player.PlayerID())
		player.PaiInc(paiImage.Pai)
		m.nextPaiImage(container, player, nextPai)
	})

	for i := 0; i < container.ChildCount(); i++ {
		if container.ChildAt(i).(PaiImage).Pai > paiImage.Pai {
			container.AddChildAt(i, paiImage)
			return
		}
	}
	container.AddChild(paiImage)
}

func main() {
	flag.Parse()
	gl.StartDriver(appMain)
}
