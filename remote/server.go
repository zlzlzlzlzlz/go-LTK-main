package remote

import (
	"encoding/gob"
	"goltk/data"
	"time"
)

type Server struct {
	rmtServer   *conn
	client      clientI
	closeSignal chan struct{}
}

func NewServer(conn *conn, client clientI) *Server {
	return &Server{rmtServer: conn, client: client, closeSignal: make(chan struct{})}
}

func (s *Server) Run() {
	var decBuff readBuf
	dec := gob.NewDecoder(&decBuff)
	enc := gob.NewEncoder(s.rmtServer)
start:
	select {
	case <-s.closeSignal:
		s.rmtServer.Write([]byte{closeSignal})
		s.rmtServer.send()
		s.rmtServer.Close()
		return
	case inf := <-s.client.GetUseCardInf():
		s.rmtServer.Write([]byte{rqsUseCard})
		enc.Encode(inf)
		s.rmtServer.send()
	case inf := <-s.client.GetDropCardInf():
		s.rmtServer.Write([]byte{rqsDropcard})
		enc.Encode(inf)
		s.rmtServer.send()
	case inf := <-s.client.GetUseSkillInf():
		s.rmtServer.Write([]byte{rqsUseSkill})
		enc.Encode(inf)
		s.rmtServer.send()
	case inf := <-s.client.GetTargetQuest():
		s.rmtServer.Write([]byte{rqsTarget})
		enc.Encode(inf)
		s.rmtServer.send()
	case inf := <-s.client.GetRole():
		s.rmtServer.Write([]byte{rqsSelRole})
		enc.Encode(inf)
		s.rmtServer.send()
	case msg := <-s.rmtServer.readCh:
		decBuff.load(msg[1:])
		switch msg[0] {
		case sendCard:
			var inf cardReceive
			errHandler(dec.Decode(&inf))
			s.client.SendCard(inf.Id, inf.Cards...)
		case dropCard:
			var inf cardReceive
			errHandler(dec.Decode(&inf))
			s.client.RemoveCard(inf.Id, inf.Cards...)
		case moveCard:
			var inf cardMoveRec
			errHandler(dec.Decode(&inf))
			s.client.MoveCard(inf.Src, inf.Dst, inf.Cards...)
		case useCard:
			var inf usecardInf
			errHandler(dec.Decode(&inf))
			s.client.UseCard(inf.User, inf.Card, inf.Targets...)
		case useTmpCard:
			var inf useTmpCardInf
			errHandler(dec.Decode(&inf))
			s.client.UseTmpCard(inf.User, inf.Cname, inf.Dec, inf.Num, inf.TmpType, inf.Targets...)
		case useSkill:
			var inf useSkillInf
			errHandler(dec.Decode(&inf))
			s.client.UseSkill(inf.User, inf.Skill, inf.Target, inf.Args...)
		case rspUseSkill:
			var inf data.UseSkillRsp
			errHandler(dec.Decode(&inf))
			s.client.SendUseSkillRsp(inf)
		case rspTarget:
			var inf data.AvailableTargetInf
			errHandler(dec.Decode(&inf))
			s.client.SendAvailableTarget(inf)
		case gsidCard:
			var inf gsidCardInf
			errHandler(dec.Decode(&inf))
			s.client.SendGSCards(inf.Id, inf.Cards...)
		case sendPid:
			var inf data.PID
			errHandler(dec.Decode(&inf))
			s.client.SetPid(inf)
		case playerList:
			var inf []data.PlayerInf
			errHandler(dec.Decode(&inf))
			s.client.SendPlayerInf(inf)
		case useAbleCard:
			var inf []data.CID
			errHandler(dec.Decode(&inf))
			s.client.SendUseAbleCards(inf)
		case useAbleSkill:
			var inf []data.SID
			errHandler(dec.Decode(&inf))
			s.client.SendUseAbleSkill(inf)
		case dropAbleCard:
			var inf dropAbleRec
			errHandler(dec.Decode(&inf))
			s.client.SendDropAbleCard(inf.Cards, inf.DropNum)
		case sendTurnOwner:
			var inf data.PID
			errHandler(dec.Decode(&inf))
			s.client.SetTurnOwner(inf)
		case sendSkillSel:
			var inf data.SID
			errHandler(dec.Decode(&inf))
			s.client.SendSkillSelect(inf)
		case availableRole:
			var inf []data.Role
			errHandler(dec.Decode(&inf))
			s.client.SendAvailableRole(inf...)
		case setHp:
			var inf setHpInf
			errHandler(dec.Decode(&inf))
			s.client.SetHP(inf.Pid, inf.Hp, inf.Hptype)
		case setGameState:
			var inf gameStateInf
			errHandler(dec.Decode(&inf))
			s.client.SetGameState(inf.State, inf.T, inf.CurPlayer)
		case closeSignal:
			s.client.Close()
			s.rmtServer.Close()
			return
		}
	}
	goto start
}

func (s *Server) Close() {
	close(s.closeSignal)
	s.client.Close()
}

type clientI interface {
	SetPid(data.PID)                               //设置玩家PID
	SendCard(data.PID, ...data.CID)                //向玩家手牌堆发牌
	RemoveCard(data.PID, ...data.CID)              //从玩家区域中移除卡牌
	MoveCard(src, dst data.PID, cards ...data.CID) //玩家间移动卡牌
	UseCard(data.PID, data.CID, ...data.PID)       //向玩家发送使用卡牌的指令
	UseTmpCard(data.PID, data.CardName, data.Decor, data.CNum, data.TmpCardType, ...data.PID)
	SendUseAbleCards([]data.CID)                          //发送玩家可用的卡牌列表
	SendUseAbleSkill([]data.SID)                          //发送可用的主动技
	SendDropAbleCard([]data.CID, uint8)                   //发送可以丢弃的卡牌列表
	SendSkillSelect(data.SID)                             //向玩家发送选择用不用技能的信号
	UseSkill(data.PID, data.SID, []data.PID, ...byte)     //发送使用技能的指令
	SendUseSkillRsp(data.UseSkillRsp)                     //发送使用主动技的回应
	SendAvailableTarget(data.AvailableTargetInf)          //向玩家发送可用目标信息
	SendPlayerInf([]data.PlayerInf)                       //向玩家发送玩家列表
	SendGSCards(data.GSID, ...data.CID)                   //发送游戏技能的卡牌
	SetHP(data.PID, data.HP, data.SetHpType)              //设置玩家HP
	SetGameState(data.GameState, time.Duration, data.PID) //设置游戏状态
	SetTurnOwner(data.PID)                                //设置当前回合拥有者
	GetUseCardInf() <-chan data.UseCardInf                //获取玩家用卡信息
	GetDropCardInf() <-chan data.DropCardInf              //获取玩家弃卡信息
	GetUseSkillInf() <-chan data.UseSkillInf              //获取玩家使用技能信息
	SendAvailableRole(...data.Role)                       //发送可选角色列表
	GetRole() <-chan data.Role                            //获取玩家选择角色
	GetTargetQuest() <-chan data.CID                      //获取玩家获取卡片可用目标的请求
	Close()
}
