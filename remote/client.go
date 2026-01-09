package remote

import (
	"encoding/gob"
	"goltk/data"
	"time"
)

var reConnHub *Hub

type Client struct {
	rmtClient          *conn //远程客户端的连接
	useCardInf         chan data.UseCardInf
	dropCardInf        chan data.DropCardInf
	useSkillInf        chan data.UseSkillInf
	cardReceiver       chan cardReceive             //接收发牌信息的chan
	removeReceiver     chan cardReceive             //接收弃牌信息的chan
	moveCardRec        chan cardMoveRec             //接收移动卡牌信息的chan
	useReceiver        chan usecardInf              //接受用牌信息的chan
	useTmpRec          chan useTmpCardInf           //接收使用临时卡信息的chan
	availableTargetrec chan data.AvailableTargetInf //接收可用目标信息的chan
	gsidCardReceiver   chan gsidCardInf             //游戏技能的卡牌接收器
	pidReceiver        chan data.PID                //接收主视角pid的chan
	playerInfRec       chan []data.PlayerInf        //接受玩家列表信息的chan
	useAbleReceiver    chan []data.CID              //可用的卡接收器
	useAbleSkillRec    chan []data.SID              //可用的主动技接收器
	useSkillRspRec     chan data.UseSkillRsp        //用主动技能的回应的接收器
	dropAbleRec        chan dropAbleRec             //可丢弃的卡的接收器
	turnOwnerRec       chan data.PID                //当前回合拥有者接收器
	skillSelectRec     chan data.SID                //问玩家要不要用技能的chan
	useSkillRec        chan useSkillInf             //接收使用技能信息的chan
	targetQuest        chan data.CID                //向服务端询问卡片可用目标
	availableRoleRec   chan []data.Role             //接收可选角色列表
	roleInf            chan data.Role               //向服务端发送选择的角色
	setHpReceiver      chan setHpInf
	gameStateRec       chan gameStateInf
	closeSignal        chan struct{}
	pid                data.PID
	reConnFn           func(data.PID, func())
}

func NewClient(rmtClient *conn) *Client {
	c := Client{
		rmtClient:          rmtClient,
		useCardInf:         make(chan data.UseCardInf, 1),
		dropCardInf:        make(chan data.DropCardInf, 1),
		useSkillInf:        make(chan data.UseSkillInf, 1),
		cardReceiver:       make(chan cardReceive, 1),
		removeReceiver:     make(chan cardReceive, 1),
		moveCardRec:        make(chan cardMoveRec, 1),
		useReceiver:        make(chan usecardInf, 1),
		useTmpRec:          make(chan useTmpCardInf, 1),
		availableTargetrec: make(chan data.AvailableTargetInf, 1),
		pidReceiver:        make(chan data.PID, 1),
		playerInfRec:       make(chan []data.PlayerInf, 1),
		useAbleReceiver:    make(chan []data.CID, 1),
		useAbleSkillRec:    make(chan []data.SID, 1),
		dropAbleRec:        make(chan dropAbleRec, 1),
		turnOwnerRec:       make(chan data.PID, 1),
		skillSelectRec:     make(chan data.SID, 1),
		useSkillRec:        make(chan useSkillInf, 1),
		useSkillRspRec:     make(chan data.UseSkillRsp, 1),
		gsidCardReceiver:   make(chan gsidCardInf, 1),
		targetQuest:        make(chan data.CID, 1),
		availableRoleRec:   make(chan []data.Role, 1),
		roleInf:            make(chan data.Role, 1),
		setHpReceiver:      make(chan setHpInf, 1),
		gameStateRec:       make(chan gameStateInf, 1),
		closeSignal:        make(chan struct{}),
	}
	rmtClient.SetOnErr(func(err error) { c.handleDisConn() })
	return &c
}

func (c *Client) Listen() {
	var decBuff readBuf
	dec := gob.NewDecoder(&decBuff)
	enc := gob.NewEncoder(c.rmtClient)
start:
	select {
	case msg := <-c.rmtClient.readCh:
		decBuff.load(msg[1:])
		switch msg[0] {
		case rqsUseCard:
			var inf data.UseCardInf
			errHandler(dec.Decode(&inf))
			c.useCardInf <- inf
		case rqsDropcard:
			var inf data.DropCardInf
			errHandler(dec.Decode(&inf))
			c.dropCardInf <- inf
		case rqsUseSkill:
			var inf data.UseSkillInf
			errHandler(dec.Decode(&inf))
			c.useSkillInf <- inf
		case rqsTarget:
			var inf data.CID
			errHandler(dec.Decode(&inf))
			c.targetQuest <- inf
		case rqsSelRole:
			var inf data.Role
			errHandler(dec.Decode(&inf))
			c.roleInf <- inf
		case closeSignal:
			// c.rmtClient.blackHold = true
			// c.rmtClient.Close()
			c.handleDisConn()
		}
	case inf := <-c.cardReceiver:
		c.rmtClient.Write([]byte{sendCard})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.removeReceiver:
		c.rmtClient.Write([]byte{dropCard})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.moveCardRec:
		c.rmtClient.Write([]byte{moveCard})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.useReceiver:
		c.rmtClient.Write([]byte{useCard})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.useTmpRec:
		c.rmtClient.Write([]byte{useTmpCard})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.useSkillRec:
		c.rmtClient.Write([]byte{useSkill})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.useSkillRspRec:
		c.rmtClient.Write([]byte{rspUseSkill})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.availableTargetrec:
		c.rmtClient.Write([]byte{rspTarget})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.gsidCardReceiver:
		c.rmtClient.Write([]byte{gsidCard})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.pidReceiver:
		c.pid = inf
		c.rmtClient.Write([]byte{sendPid})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.playerInfRec:
		c.rmtClient.Write([]byte{playerList})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.useAbleReceiver:
		c.rmtClient.Write([]byte{useAbleCard})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.useAbleSkillRec:
		c.rmtClient.Write([]byte{useAbleSkill})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.dropAbleRec:
		c.rmtClient.Write([]byte{dropAbleCard})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.turnOwnerRec:
		c.rmtClient.Write([]byte{sendTurnOwner})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.skillSelectRec:
		c.rmtClient.Write([]byte{sendSkillSel})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.availableRoleRec:
		c.rmtClient.Write([]byte{availableRole})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.setHpReceiver:
		c.rmtClient.Write([]byte{setHp})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case inf := <-c.gameStateRec:
		c.rmtClient.Write([]byte{setGameState})
		errHandler(enc.Encode(inf))
		c.rmtClient.send()
	case <-c.closeSignal:
		c.rmtClient.Write([]byte{closeSignal})
		c.rmtClient.send()
		c.rmtClient.Close()
		return
	}
	goto start
}

func (c *Client) GetUseCardInf() <-chan data.UseCardInf {
	return c.useCardInf
}

func (c *Client) GetDropCardInf() <-chan data.DropCardInf {
	return c.dropCardInf
}

func (c *Client) GetUseSkillInf() <-chan data.UseSkillInf {
	return c.useSkillInf
}

func (c *Client) SetPid(id data.PID) {
	c.pidReceiver <- id
}

func (c *Client) SendAvailableRole(roles ...data.Role) {
	c.availableRoleRec <- roles
}

func (c *Client) GetRole() <-chan data.Role {
	return c.roleInf
}

func (c *Client) SendCard(id data.PID, cards ...data.CID) {
	c.cardReceiver <- cardReceive{Id: id, Cards: cards}
}

func (c *Client) RemoveCard(id data.PID, cards ...data.CID) {
	c.removeReceiver <- cardReceive{Id: id, Cards: cards}
}

func (c *Client) MoveCard(src, dst data.PID, cards ...data.CID) {
	c.moveCardRec <- cardMoveRec{Src: src, Dst: dst, Cards: cards}
}

func (c *Client) UseCard(user data.PID, card data.CID, targets ...data.PID) {
	c.useReceiver <- usecardInf{User: user, Card: card, Targets: targets}
}

func (c *Client) UseTmpCard(user data.PID, name data.CardName, dec data.Decor, num data.CNum,
	tmpType data.TmpCardType, target ...data.PID) {
	c.useTmpRec <- useTmpCardInf{User: user, Cname: name, Dec: dec, Num: num, TmpType: tmpType, Targets: target}
}

func (c *Client) SendUseAbleCards(cards []data.CID) {
	c.useAbleReceiver <- cards
}

func (c *Client) SendUseAbleSkill(skills []data.SID) {
	c.useAbleSkillRec <- skills
}

func (c *Client) SendDropAbleCard(cards []data.CID, dropNum uint8) {
	c.dropAbleRec <- dropAbleRec{Cards: cards, DropNum: dropNum}
}

func (c *Client) SendSkillSelect(sid data.SID) {
	c.skillSelectRec <- sid
}

func (c *Client) SendUseSkillRsp(rsp data.UseSkillRsp) {
	c.useSkillRspRec <- rsp
}

func (c *Client) UseSkill(user data.PID, skill data.SID, target []data.PID, args ...byte) {
	c.useSkillRec <- useSkillInf{User: user, Skill: skill, Target: target, Args: args}
}

func (c *Client) SendAvailableTarget(inf data.AvailableTargetInf) {
	c.availableTargetrec <- inf
}

func (c *Client) SendPlayerInf(inf []data.PlayerInf) {
	c.playerInfRec <- inf
}

func (c *Client) SendGSCards(id data.GSID, cards ...data.CID) {
	c.gsidCardReceiver <- gsidCardInf{Id: id, Cards: cards}
}

func (c *Client) SetHP(pid data.PID, hp data.HP, dmgtype data.SetHpType) {
	c.setHpReceiver <- setHpInf{Pid: pid, Hp: hp, Hptype: dmgtype}
}

func (c *Client) SetGameState(state data.GameState, t time.Duration, curPlayer data.PID) {
	c.gameStateRec <- gameStateInf{State: state, T: t, CurPlayer: curPlayer}
}

func (c *Client) SetTurnOwner(pid data.PID) {
	c.turnOwnerRec <- pid
}

func (c *Client) GetTargetQuest() <-chan data.CID {
	return c.targetQuest
}

func (c *Client) GetClientType() data.PlayerType {
	return data.RemotePlayer
}

func (c *Client) Close() {
	close(c.closeSignal)
}

func (c *Client) SetReConnFn(fn func(data.PID, func())) {
	c.reConnFn = fn
}

func (c *Client) handleDisConn() {
	c.rmtClient.blackHold = true
	c.rmtClient.Close()
	waitForReConn := func() {
		select {
		case <-c.closeSignal:
			reConnHub.Close()
		case conn := <-reConnHub.cons:
			reConnHub.Close()
			c.reConnFn(c.pid, func() {
				c.Close()
				<-time.After(1 * time.Millisecond)
				c.closeSignal = make(chan struct{})
				c.rmtClient = NewConn(conn)
				c.rmtClient.SetOnErr(func(err error) { c.handleDisConn() })
				go c.Listen()
			})
		}
	}
	if reConnHub == nil {
		reConnHub = NewHub(1)
		go waitForReConn()
	} else {
		go func() {
			select {
			case <-reConnHub.closeSignal:
				reConnHub = NewHub(1)
				go waitForReConn()
			case <-c.closeSignal:
			}
		}()
	}
}
