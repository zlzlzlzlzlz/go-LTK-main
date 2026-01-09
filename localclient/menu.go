package localclient

import (
	"goltk/app"
	"goltk/bot"
	"goltk/data"
	"goltk/front"
	"goltk/remote"
	"goltk/server"
	"image/color"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

type menu struct {
	bgImg   *ebiten.Image
	btnList []menuButtonI
}

func (m *menu) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(m.bgImg, op)
	for i := 0; i < len(m.btnList); i++ {
		m.btnList[i].Draw(screen)
	}
}

func (m *menu) Update(a *app.App) {
	for i := 0; i < len(m.btnList); i++ {
		m.btnList[i].Update(a)
	}
}

func NewMianMenu(a *app.App) *menu {
	mainMenu := &menu{bgImg: getImg("assets/menu/menubg/" + strconv.Itoa(rand.Intn(9)) + ".jpg"),
		btnList: []menuButtonI{newMenuButton(200, 100, getImg("assets/menu/button/0.png"), func(a *app.App) {
			a.Bgm.Pause()
			c := NewGames(a)
			s := server.NewGame(data.QGZXEasyMode)
			a.Server = s
			s.AddClient(c)
			a.CurMenu = c
			for i := 0; i < 2; i++ {
				b := bot.NewBot([]bool{true, true, true, false})
				s.AddClient(b)
				go b.Run()
			}
			b := bot.NewBot([]bool{false, false, false, true})
			s.AddClient(b)
			go b.Run()
			go a.Server.Run()
		}),
			newMenuButton(520, 100, getImg("assets/menu/button/boss.png"), func(a *app.App) { a.CurMenu = newBossRoom() }),
			newMenuButton(200, 600, getImg("assets/menu/button/roleDetail.png"), func(a *app.App) { a.CurMenu = newRoleDetailSelMenu() }),
			newMenuButton(780, 80, getImg("assets/menu/button/zyxl.png"), func(a *app.App) { a.CurMenu = newTestMenu() }),
			newMenuButton(780, 400, getImg("assets/menu/button/qgzx.png"), func(a *app.App) { a.CurMenu = newQGZXRoom() }),
		}}
	inputBox := newInputBox(210, 490)
	var confirm *menuSearchBtn
	confirm = newMenuSearchBtn(640, 480, func(a *app.App) {
		a.Bgm.Pause()
		c := NewGames(a)
		conn, err := remote.DialServer(inputBox.text.Str)
		if err != nil {
			confirm.promptText.SetPos(300, 600)
			confirm.promptText.SetText("[white]无法连接至服务器")
			return
		}
		s := remote.NewServer(remote.NewConn(conn), c)
		a.CurMenu = c
		a.Server = s
		go s.Run()
	})
	mainMenu.btnList = append(mainMenu.btnList, inputBox, confirm)
	return mainMenu
}

type hostRoomMenu struct {
	menu
	ready   bool
	selList [8]data.PlayerType
}

func newHostRoomMenu() *hostRoomMenu {
	menu := &hostRoomMenu{menu: menu{bgImg: getImg("assets/menu/menubg/" + strconv.Itoa(rand.Intn(9)) + ".jpg")}}
	getOnclick := func(line int, b *selectButton, pType data.PlayerType) func(*app.App) {
		return func(a *app.App) {
			if pType == data.LocalPlayer {
				for i, p := range menu.selList {
					if p == data.LocalPlayer {
						menu.btnList[i*4+int(p)].(*selectButton).setSelect(false)
						menu.btnList[i*4].(*selectButton).setSelect(true)
						menu.selList[i] = data.NoPlayer
					}
				}
			}
			for i := line * 4; i < line*4+4; i++ {
				menu.btnList[i].(*selectButton).setSelect(false)
			}
			b.setSelect(true)
			menu.selList[line] = pType
		}
	}
	for line := range len(menu.selList) {
		for i := 0; i < 4; i++ {
			pType := data.PlayerType(i)
			btn := newSelectButton(float64(i*100), float64(line*40), pType.Name(), nil)
			if i == 0 {
				btn.highLight = true
			}
			btn.onclick = getOnclick(line, btn, pType)
			menu.btnList = append(menu.btnList, btn)
		}
	}
	menu.btnList = append(menu.btnList, newMenuConfirmBtn(500, 500, func(a *app.App) {
		remoteCount := 0
		for _, p := range menu.selList {
			if p == data.RemotePlayer {
				remoteCount++
			}
		}
		hub := remote.NewHub(remoteCount)
		s := server.NewGame(data.QGZXNormalMode)
		a.Server = s
		for _, p := range menu.selList {
			switch p {
			case data.LocalPlayer:
				c := NewGames(a)
				s.AddClient(c)
				a.CurMenu = c
			case data.RemotePlayer:
				c := remote.NewClient(remote.NewConn(hub.GetConn()))
				go c.Listen()
				s.AddClient(c)
			case data.BotPlayer:
				c := bot.NewBot([]bool{false, false, false, false})
				go c.Run()
				s.AddClient(c)
			default:
				continue
			}
		}
		a.Bgm.Pause()
		go a.Server.Run()
	}))
	return menu
}

func (m *hostRoomMenu) Update(a *app.App) {
	if !m.ready {
		for i := 0; i < len(m.btnList); i++ {
			m.btnList[i].Update(a)
		}
		return
	}
}

type qgzxRoom struct {
	menu
	playerTypes [3]data.PlayerType
	gameMode    data.GameMode
	ready       bool
	index       int
	hub         *remote.Hub
	s           *server.Games
	c           *Games
	quitBtn     menuButtonI
	promptText  *front.TextItem2
}

func newQGZXRoom() *qgzxRoom {
	m := &qgzxRoom{menu: menu{bgImg: getImg("assets/menu/menubg/qgzxBg.png")},
		playerTypes: [...]data.PlayerType{data.LocalPlayer, data.RemotePlayer, data.RemotePlayer}, gameMode: data.QGZXEasyMode,
		promptText: front.NewTextItem2("", 300, 620, 24, 1, 24)}
	m.quitBtn = newBackBtn(1260, 20, func(a *app.App) {
		if m.hub != nil {
			m.hub.Close()
		}
		a.CurMenu = NewMianMenu(a)
	})
	getOnClick := func(btn *menuTextBtn, index int) func(*app.App) {
		return func(g *app.App) {
			if m.playerTypes[index] == data.BotPlayer {
				m.playerTypes[index] = data.LocalPlayer
			} else {
				m.playerTypes[index]++
			}
			btn.text.SetText("[yellow]" + m.playerTypes[index].Name())
		}
	}
	p1Btn := newMenuTextBtn(150, 430, m.playerTypes[0].Name(), nil)
	p1Btn.onclick = getOnClick(p1Btn, 0)
	p2Btn := newMenuTextBtn(150, 500, m.playerTypes[1].Name(), nil)
	p2Btn.onclick = getOnClick(p2Btn, 1)
	p3Btn := newMenuTextBtn(150, 570, m.playerTypes[2].Name(), nil)
	p3Btn.onclick = getOnClick(p3Btn, 2)
	confirmBtn := newMenuButton(1100, 600, btnImgList[15], func(a *app.App) {
		localCount := 0
		remoteCount := 0
		for i := 0; i < len(m.playerTypes); i++ {
			switch m.playerTypes[i] {
			case data.LocalPlayer:
				localCount++
			case data.RemotePlayer:
				remoteCount++
			}
		}
		if localCount != 1 {
			m.promptText.SetText("[yellow]本地玩家[white]必须有且只有一个")
			return
		}
		if remoteCount > 0 {
			m.hub = remote.NewHub(remoteCount)
		}
		m.s = server.NewGame(m.gameMode)
		m.ready = true
	})
	var easy, normal, hard, veryHard, double, free *modeSelectBtn
	easy = newModeSelectBtn(155, 95, getImg("assets/menu/button/easy.png"), func(a *app.App) {
		m.gameMode = data.QGZXEasyMode
		easy.selected = true
		iterators([]*modeSelectBtn{normal, hard, veryHard, double, free}, func(b *modeSelectBtn) { b.selected = false })
	})
	easy.selected = true
	normal = newModeSelectBtn(333, 93, getImg("assets/menu/button/normal.png"), func(a *app.App) {
		m.gameMode = data.QGZXNormalMode
		normal.selected = true
		iterators([]*modeSelectBtn{easy, hard, veryHard, double, free}, func(b *modeSelectBtn) { b.selected = false })
	})
	hard = newModeSelectBtn(514, 95, getImg("assets/menu/button/hard.png"), func(a *app.App) {
		m.gameMode = data.QGZXHardMode
		hard.selected = true
		iterators([]*modeSelectBtn{easy, normal, veryHard, double, free}, func(b *modeSelectBtn) { b.selected = false })
	})
	veryHard = newModeSelectBtn(692, 95, getImg("assets/menu/button/veryHard.png"), func(a *app.App) {
		m.gameMode = data.QGZXVeryHardMode
		veryHard.selected = true
		iterators([]*modeSelectBtn{easy, normal, hard, double, free}, func(b *modeSelectBtn) { b.selected = false })
	})
	double = newModeSelectBtn(872, 95, getImg("assets/menu/button/double.png"), func(a *app.App) {
		m.gameMode = data.QGZXDoubleMode
		double.selected = true
		iterators([]*modeSelectBtn{easy, normal, hard, veryHard, free}, func(b *modeSelectBtn) { b.selected = false })
	})
	free = newModeSelectBtn(1052, 95, getImg("assets/menu/button/free.png"), func(a *app.App) {
		m.gameMode = data.QGZXFreeMode
		free.selected = true
		iterators([]*modeSelectBtn{easy, normal, hard, veryHard, double}, func(b *modeSelectBtn) { b.selected = false })
	})
	m.btnList = []menuButtonI{easy, normal, hard, veryHard, double, free, confirmBtn, p1Btn, p2Btn, p3Btn}
	return m
}

func (m *qgzxRoom) Draw(screen *ebiten.Image) {
	m.menu.Draw(screen)
	m.promptText.Draw(screen)
	m.quitBtn.Draw(screen)
}

func (m *qgzxRoom) Update(a *app.App) {
	m.quitBtn.Update(a)
	if !m.ready {
		for i := 0; i < len(m.btnList); i++ {
			m.btnList[i].Update(a)
		}
		return
	}
	switch m.playerTypes[m.index] {
	case data.LocalPlayer:
		c := NewGames(a)
		m.s.AddClient(c)
		m.c = c
		m.index++
	case data.RemotePlayer:
		select {
		case con := <-m.hub.GetConnChan():
			rmt := remote.NewClient(remote.NewConn(con))
			rmt.SetReConnFn(m.s.ClientReConn)
			go rmt.Listen()
			m.s.AddClient(rmt)
			m.index++
		default:
			return
		}
	case data.BotPlayer:
		var teamMate []bool
		if m.gameMode == data.QGZXEasyMode || m.gameMode == data.QGZXDoubleMode || m.gameMode == data.QGZXNormalMode {
			teamMate = []bool{true, true, true, false}
		} else if m.gameMode == data.QGZXFreeMode {
			if m.index == 2 {
				teamMate = []bool{false, false, true}
			} else {
				teamMate = []bool{true, true, false}
			}
		} else {
			teamMate = []bool{true, true, true, false, false, false}
		}
		b := bot.NewBot(teamMate)
		go b.Run()
		m.s.AddClient(b)
		m.index++
	}
	m.promptText.SetText("[white]正在等待玩家[blue] " + strconv.Itoa(m.index+1) + "[white] 的加入")
	if m.index != len(m.playerTypes) {
		return
	}
	switch m.gameMode {
	case data.QGZXEasyMode:
		b := bot.NewBot([]bool{false, false, false, true})
		m.s.AddClient(b)
		go b.Run()
	case data.QGZXNormalMode:
		b := bot.NewBot([]bool{false, false, false, true})
		m.s.AddClient(b)
		go b.Run()
	case data.QGZXDoubleMode:
		b := bot.NewBot([]bool{false, false, false, true})
		m.s.AddClient(b)
		go b.Run()
	case data.QGZXFreeMode:
	default:
		for i := 0; i < 3; i++ {
			b := bot.NewBot([]bool{false, false, false, true, true, true})
			m.s.AddClient(b)
			go b.Run()
		}
	}
	a.CurMenu = m.c
	a.Server = m.s
	a.Bgm.Pause()
	go m.s.Run()
}

type testMenu struct {
	menu
	text1 *front.TextItem
	text2 *front.TextItem2
}

func newTestMenu() *testMenu {
	m := &testMenu{menu: menu{bgImg: getImg("assets/menu/menubg/" + strconv.Itoa(rand.Intn(9)) + ".jpg")},
		text1: front.NewTextItem("白色黑色蓝色红色绿色", 100, 100, 24, false, 32, color.White),
		text2: front.NewTextItem2("[yellow]白色[white]黄色", 100, 200, 18, 0, 18),
	}
	return m
}

func (m *testMenu) Draw(screen *ebiten.Image) {
	m.menu.Draw(screen)
	m.text2.Draw(screen)
	m.text1.Draw(screen)
}

func newRoleDetailSelMenu() *menu {
	m := &menu{
		bgImg: getImg("assets/menu/menubg/" + strconv.Itoa(rand.Intn(9)) + ".jpg"),
	}
	m.btnList = append(m.btnList, newBackBtn(1260, 690,
		func(a *app.App) { a.CurMenu = NewMianMenu(a) }))
	roleList := data.RoleList
	x, y := -110.0, 20.0
	for i := 0; i < len(roleList); i++ {
		if i%10 == 0 && i != 0 {
			y += 130
			x = 20
		} else {
			x += 130
		}
		b := newRoleDetailBtn(x, y, roleList[i], nil)
		b.onclick = func(a *app.App) { a.CurMenu = newRoleDetailMenu(b.role) }
		m.btnList = append(m.btnList, b)
	}
	return m
}

type roleDetailMenu struct {
	menu
	role data.Role
}

func newRoleDetailMenu(role data.Role) *roleDetailMenu {
	m := &roleDetailMenu{role: role}
	m.btnList = append(m.btnList, newBackBtn(1180, 630,
		func(a *app.App) { a.CurMenu = newRoleDetailSelMenu() }))
	return m
}

func (m *roleDetailMenu) Draw(screen *ebiten.Image) {
	text := ""
	for _, skill := range m.role.SkillList {
		const maxLen = 20 //每行的最大长度
		str := []rune(skill.Name() + "：" + skill.Text())
		for i := maxLen; i < len(str); i += maxLen {
			str = append(str[:i], append([]rune("\n"), str[i:]...)...)
		}
		text += "[yellow]" + skill.Name() + "：[white]" + string(str[len([]rune(skill.Name()))+1:])
		if len(str) != 0 && str[len(str)-1] != rune('\n') {
			text += "\n\n"
		}
	}
	skillText := front.NewTextItem2(text, 560, 150, 29, 1.8, 24)
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(getImg("assets/menu/menubg/roleDetail.png"), op)
	//画选将信息
	//计算卡面坐标并绘制
	op.GeoM.Reset()
	op.GeoM.Scale(1.4, 1.4)
	img := getPlayerImg("assets/role/charImg/" + m.role.Name + ".png")
	op.GeoM.Translate(220, 302-float64(img.Bounds().Size().Y/2))
	screen.DrawImage(img, op)
	//绘制名字
	roleNameBg := getImg("assets/menu/roleNameBg.png")
	op.GeoM.Reset()
	op.GeoM.Translate(186, 215)
	screen.DrawImage(roleNameBg, op)
	front.DrawText(screen, m.role.DspName, 205, 230, 30, true)
	skillText.Draw(screen)
	for i := 0; i < len(m.btnList); i++ {
		m.btnList[i].Draw(screen)
	}
}

type bossRoom struct {
	menu
	playerTypes [3]data.PlayerType
	gameMode    data.GameMode
	ready       bool
	index       int
	hub         *remote.Hub
	s           *server.Games
	c           *Games
	quitBtn     menuButtonI
	promptText  *front.TextItem2
}

func newBossRoom() *bossRoom {
	m := &bossRoom{menu: menu{bgImg: getImg("assets/menu/menubg/boss.png")},
		playerTypes: [...]data.PlayerType{data.LocalPlayer, data.RemotePlayer, data.RemotePlayer}, gameMode: data.NianShouMode,
		promptText: front.NewTextItem2("", 300, 620, 24, 1, 24)}
	m.quitBtn = newBackBtn(1260, 20, func(a *app.App) {
		if m.hub != nil {
			m.hub.Close()
		}
		a.CurMenu = NewMianMenu(a)
	})
	getOnClick := func(btn *menuTextBtn, index int) func(*app.App) {
		return func(g *app.App) {
			if m.playerTypes[index] == data.BotPlayer {
				m.playerTypes[index] = data.LocalPlayer
			} else {
				m.playerTypes[index]++
			}
			btn.text.SetText("[yellow]" + m.playerTypes[index].Name())
		}
	}
	p1Btn := newMenuTextBtn(10, 500, m.playerTypes[0].Name(), nil)
	p1Btn.onclick = getOnClick(p1Btn, 0)
	p2Btn := newMenuTextBtn(10, 570, m.playerTypes[1].Name(), nil)
	p2Btn.onclick = getOnClick(p2Btn, 1)
	p3Btn := newMenuTextBtn(10, 640, m.playerTypes[2].Name(), nil)
	p3Btn.onclick = getOnClick(p3Btn, 2)
	confirmBtn := newMenuButton(1100, 600, btnImgList[15], func(a *app.App) {
		localCount := 0
		remoteCount := 0
		for i := 0; i < len(m.playerTypes); i++ {
			switch m.playerTypes[i] {
			case data.LocalPlayer:
				localCount++
			case data.RemotePlayer:
				remoteCount++
			}
		}
		if localCount != 1 {
			m.promptText.SetText("[yellow]本地玩家[white]必须有且只有一个")
			return
		}
		if remoteCount > 0 {
			m.hub = remote.NewHub(remoteCount)
		}
		m.s = server.NewGame(m.gameMode)
		m.ready = true
	})
	var nianshou *modeSelectBtn
	nianshou = newModeSelectBtn(220, 100, getImg("assets/menu/button/nianshou.png"), func(a *app.App) {
		m.gameMode = data.NianShouMode
		nianshou.selected = true
		iterators([]*modeSelectBtn{}, func(b *modeSelectBtn) { b.selected = false })
	})
	nianshou.selected = true
	// normal = newModeSelectBtn(333, 93, getImg("assets/menu/button/normal.png"), func(a *app.App) {
	// 	m.gameMode = data.QGZXNormalMode
	// 	normal.selected = true
	// 	iterators([]*modeSelectBtn{easy, hard, veryHard, double, free}, func(b *modeSelectBtn) { b.selected = false })
	// })
	// hard = newModeSelectBtn(514, 95, getImg("assets/menu/button/hard.png"), func(a *app.App) {
	// 	m.gameMode = data.QGZXHardMode
	// 	hard.selected = true
	// 	iterators([]*modeSelectBtn{easy, normal, veryHard, double, free}, func(b *modeSelectBtn) { b.selected = false })
	// })
	// veryHard = newModeSelectBtn(692, 95, getImg("assets/menu/button/veryhard.png"), func(a *app.App) {
	// 	m.gameMode = data.QGZXVeryHardMode
	// 	veryHard.selected = true
	// 	iterators([]*modeSelectBtn{easy, normal, hard, double, free}, func(b *modeSelectBtn) { b.selected = false })
	// })
	// double = newModeSelectBtn(872, 95, getImg("assets/menu/button/double.png"), func(a *app.App) {
	// 	m.gameMode = data.QGZXDoubleMode
	// 	double.selected = true
	// 	iterators([]*modeSelectBtn{easy, normal, hard, veryHard, free}, func(b *modeSelectBtn) { b.selected = false })
	// })
	// free = newModeSelectBtn(1052, 95, getImg("assets/menu/button/free.png"), func(a *app.App) {
	// 	m.gameMode = data.QGZXFreeMode
	// 	free.selected = true
	// 	iterators([]*modeSelectBtn{easy, normal, hard, veryHard, double}, func(b *modeSelectBtn) { b.selected = false })
	// })
	m.btnList = []menuButtonI{nianshou, confirmBtn, p1Btn, p2Btn, p3Btn}
	return m
}

func (m *bossRoom) Draw(screen *ebiten.Image) {
	m.menu.Draw(screen)
	m.promptText.Draw(screen)
	m.quitBtn.Draw(screen)
}

func (m *bossRoom) Update(a *app.App) {
	m.quitBtn.Update(a)
	if !m.ready {
		for i := 0; i < len(m.btnList); i++ {
			m.btnList[i].Update(a)
		}
		return
	}
	switch m.playerTypes[m.index] {
	case data.LocalPlayer:
		c := NewGames(a)
		m.s.AddClient(c)
		m.c = c
		m.index++
	case data.RemotePlayer:
		select {
		case con := <-m.hub.GetConnChan():
			rmt := remote.NewClient(remote.NewConn(con))
			go rmt.Listen()
			m.s.AddClient(rmt)
			m.index++
		default:
			return
		}
	case data.BotPlayer:
		var teamMate []bool
		teamMate = []bool{true, true, true, false}
		b := bot.NewBot(teamMate)
		go b.Run()
		m.s.AddClient(b)
		m.index++
	}
	m.promptText.SetText("[white]正在等待玩家[blue] " + strconv.Itoa(m.index+1) + "[white] 的加入")
	if m.index != len(m.playerTypes) {
		return
	}
	// switch m.gameMode {
	// case data.QGZXEasyMode:
	b := bot.NewBot([]bool{false, false, false, true})
	m.s.AddClient(b)
	go b.Run()
	// case data.QGZXNormalMode:
	// 	b := bot.NewBot([]bool{false, false, false, true})
	// 	m.s.AddClient(b)
	// 	go b.Run()
	// case data.QGZXDoubleMode:
	// 	b := bot.NewBot([]bool{false, false, false, true})
	// 	m.s.AddClient(b)
	// 	go b.Run()
	// case data.QGZXFreeMode:
	// default:
	// 	for i := 0; i < 3; i++ {
	// 		b := bot.NewBot([]bool{false, false, false, true, true, true})
	// 		m.s.AddClient(b)
	// 		go b.Run()
	// 	}
	// }
	a.CurMenu = m.c
	a.Server = m.s
	a.Bgm.Pause()
	go m.s.Run()
}
