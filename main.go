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

	"fmt"

	"io/ioutil"

	"github.com/kyokomi/gomajan/pai"
	"github.com/kyokomi/gomajan/player"
	"github.com/kyokomi/gomajan/taku"
)

type State int

const (
	InitState       State = 0
	PlayerTurnState State = 1
	CPUTurnState    State = 2
	EndState        State = 3
	PlayerEndState  State = 4
)

type MainWindow struct {
	gxui.Window
	theme  gxui.Theme
	driver gxui.Driver
}

func inflate(filePath string) []byte {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return b
}

func NewMainWindow(driver gxui.Driver) MainWindow {
	var mainWindow MainWindow
	theme := dark.CreateTheme(driver)

	// TODO: go-bindataかな
	defaultFont, _ := driver.CreateFont(inflate("data/misaki_gothic.ttf"), 24)
	theme.SetDefaultFont(defaultFont)
	window := theme.CreateWindow(1200, 800, "gxjan")

	mainWindow.Window = window
	mainWindow.theme = theme
	mainWindow.driver = driver

	return mainWindow
}

var _state State
var _taku *taku.Taku
var _yamaImage [4][2][17]*PaiImage

var _playerContainer [4]gxui.LinearLayout

func ChangeState(state State) {
	beforeState := _state
	_state = state

	if beforeState == PlayerTurnState && state == CPUTurnState {
		// playerTurn -> CPUTurn
		EventCPU(2)
		EventCPU(3)
		EventCPU(4)
		ChangeState(PlayerTurnState)
	} else if state == EndState {
		fmt.Println("終了")
		playerMessage("終了", "")
	} else if state == PlayerTurnState {
		playerMessage("Player", "の番")
	}
}

func GameEnd() bool {
	return _state == EndState || _state == PlayerEndState
}

func NowState() State {
	return _state
}

func EventCPU(playerID int) {
	// 1枚引いて
	mjPai := nextFunc(playerID)

	// 1枚捨てる
	p := Player(playerID)
	p.PaiInc(mjPai)
	_window.nextPaiImage(_playerContainer[playerID-1], playerID, mjPai)

	yakuCheck := p.NewYakuCheck(mjPai)
	if yakuCheck.Is和了() {
		fmt.Println("Enemy アガリ: ", yakuCheck.String())
		ChangeState(EndState)
		return
	}
	//fmt.Println("チェック: ", yakuCheck.String())

	for _, noko := range yakuCheck.MentsuCheck().Nokori() {
		if noko.Pai.Type() == pai.NoneType || noko.Val == 0 {
			continue
		}

		//fmt.Println("捨てた牌: ", noko.Pai)
		p.PaiDec(noko.Pai)

		for i := 0; i < _playerContainer[playerID-1].ChildCount(); i++ {
			paiImage := _playerContainer[playerID-1].ChildAt(i).(PaiImage)
			if paiImage.Pai == noko.Pai {
				_playerContainer[playerID-1].RemoveChildAt(i)
				_sutehaiLayers[playerID-1].AddChild(paiImage)
				break
			}
		}
		break
	}
}

var _infoLabel gxui.Label
var _window MainWindow
var _sutehaiLayers [4]gxui.LinearLayout

func playerMessage(name string, message string) {
	msg := fmt.Sprintf("東%d局 ドラ %s %d本場 %s %s", 1, _taku.Dora()[0], 0, name, message)
	setInfoLabel(msg)
}
func setInfoLabel(text string) {
	_infoLabel.SetText(text)
}

func appMain(driver gxui.Driver) {
	_window = NewMainWindow(driver)

	ChangeState(InitState)

	_taku = taku.NewTaku()

	// TODO: Splitterにしてみた
	rootLayer := _window.Theme().CreateSplitterLayout()
	//	rootLayer.SetOrientation(gxui.TopToBottom)
	//	rootLayer.SetVerticalAlignment(gxui.AlignTop)

	// TODO: メッセージ的なやつ出す
	infoLayer := _window.Theme().CreateSplitterLayout()
	infoLayer.SetOrientation(gxui.Horizontal)
	{
		layer := _window.Theme().CreateLinearLayout()
		layer.SetDirection(gxui.TopToBottom)
		layer.SetVerticalAlignment(gxui.AlignTop)
		layer.SetMargin(math.CreateSpacing(10))

		label := _window.Theme().CreateLabel()
		label.SetText("hogehogehoge")
		layer.AddChild(label)
		infoLayer.AddChild(layer)
		_infoLabel = label
	}
	rootLayer.AddChild(infoLayer)
	rootLayer.SetChildWeight(infoLayer, 0.07)

	// 手牌
	{
		playLayer := _window.Theme().CreateSplitterLayout()
		playLayer.SetOrientation(gxui.Horizontal)

		tehaiLayer := _window.Theme().CreateLinearLayout()
		tehaiLayer.SetDirection(gxui.TopToBottom)
		tehaiLayer.SetVerticalAlignment(gxui.AlignTop)
		tehaiLayer.SetMargin(math.CreateSpacing(10))

		sutehaiLayer := _window.Theme().CreateLinearLayout()
		sutehaiLayer.SetDirection(gxui.TopToBottom)
		sutehaiLayer.SetVerticalAlignment(gxui.AlignTop)
		sutehaiLayer.SetMargin(math.CreateSpacing(10))

		for _, player := range _taku.Players {
			container := _window.Theme().CreateLinearLayout()
			container.SetDirection(gxui.LeftToRight)
			container.SetHorizontalAlignment(gxui.AlignCenter)

			for _, hai := range player.Tiles() {
				if hai.Pai.Type() == pai.NoneType || hai.Val <= 0 {
					continue
				}

				for i := 0; i < hai.Val; i++ {
					_window.nextPaiImage(container, player.PlayerID(), hai.Pai)
				}
			}
			container.SetMargin(math.CreateSpacing(10))
			tehaiLayer.AddChild(container)
			_playerContainer[player.PlayerID()-1] = container

			suteContainer := _window.Theme().CreateLinearLayout()
			suteContainer.SetDirection(gxui.LeftToRight)
			suteContainer.SetHorizontalAlignment(gxui.AlignCenter)
			suteContainer.SetMargin(math.CreateSpacing(10))

			_sutehaiLayers[player.PlayerID()-1] = suteContainer
			sutehaiLayer.AddChild(suteContainer)
		}
		playLayer.AddChild(tehaiLayer)
		playLayer.SetChildWeight(tehaiLayer, 0.5)

		playLayer.AddChild(sutehaiLayer)
		playLayer.SetChildWeight(sutehaiLayer, 0.5)

		rootLayer.AddChild(playLayer)
		rootLayer.SetChildWeight(playLayer, 0.35)
	}

	// 山
	{
		yamaLayer := _window.Theme().CreateLinearLayout()
		yamaLayer.SetDirection(gxui.TopToBottom)
		yamaLayer.SetVerticalAlignment(gxui.AlignTop)
		yamaLayer.SetMargin(math.CreateSpacing(10))

		for i, 山2 := range _taku.Yama {
			container := _window.Theme().CreateLinearLayout()
			container.SetDirection(gxui.TopToBottom)
			container.SetVerticalAlignment(gxui.AlignTop)
			for j, 山 := range 山2 {
				container2 := _window.Theme().CreateLinearLayout()
				container2.SetDirection(gxui.LeftToRight)
				container2.SetHorizontalAlignment(gxui.AlignCenter)
				for k, 牌 := range 山 {
					if 牌.Type() == pai.NoneType {
						continue
					}

					imagePai := _window.createPaiImage(牌)
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

	_window.AddChild(rootLayer)
	_window.OnClose(driver.Terminate)

	ChangeState(PlayerTurnState)

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

func Player(playerID int) *player.Player {
	for _, p := range _taku.Players {
		if p.PlayerID() == playerID {
			return &p
		}
	}
	return nil
}

func (m MainWindow) nextPaiImage(container gxui.LinearLayout, playerID int, mjPai pai.MJP) {
	paiImage := m.createPaiImage(mjPai)

	fmt.Println("nextPaiImage playerID = ", playerID)
	// TODO: player1だけ操作可能にする
	if playerID == 1 {
		paiImage.OnClick(func(e gxui.MouseEvent) {
			if GameEnd() {
				return
			}
			ChangeState(CPUTurnState)

			fmt.Println("OnClick", playerID, paiImage.Pai)

			Player(playerID).PaiDec(paiImage.Pai)
			container.RemoveChild(paiImage)

			_sutehaiLayers[playerID-1].AddChild(paiImage)

			nextPai := nextFunc(playerID)
			Player(playerID).PaiInc(nextPai)
			m.nextPaiImage(container, playerID, nextPai)
		})
	}

	isAdded := false
	for i := 0; i < container.ChildCount(); i++ {
		if container.ChildAt(i).(PaiImage).Pai > paiImage.Pai {
			container.AddChildAt(i, paiImage)
			isAdded = true
			break
		}
	}

	if !isAdded {
		container.AddChild(paiImage)
	}

	// TODO: 明かり判定
	if playerID == 1 {
		p := Player(playerID)
		yakuCheck := p.NewYakuCheck(mjPai)
		if yakuCheck.Is和了() {
			fmt.Println("!!! Player アガリ: ", yakuCheck.String())
			playerMessage("Player", "ツモ" + fmt.Sprintf("%s", yakuCheck.Yakus()))
			ChangeState(PlayerEndState)
		} else {
			fmt.Println("!!! Player チェック: ", yakuCheck.String())
		}
	}
}

func main() {
	flag.Parse()
	gl.StartDriver(appMain)
}
