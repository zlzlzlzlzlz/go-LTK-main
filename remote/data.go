package remote

import (
	"goltk/data"
	"time"
)

type cardReceive struct {
	Id    data.PID
	Cards []data.CID
}

type cardMoveRec struct {
	Src, Dst data.PID
	Cards    []data.CID
}

type dropAbleRec struct {
	Cards   []data.CID
	DropNum uint8
}

type gameStateInf struct {
	State     data.GameState
	T         time.Duration
	CurPlayer data.PID
}

type usecardInf struct {
	User    data.PID
	Card    data.CID
	Targets []data.PID
}

type useSkillInf struct {
	User   data.PID
	Skill  data.SID
	Target []data.PID
	Args   []byte
}

type useTmpCardInf struct {
	User    data.PID
	Cname   data.CardName
	Dec     data.Decor
	Num     data.CNum
	TmpType data.TmpCardType
	Targets []data.PID
}

type setHpInf struct {
	Pid    data.PID
	Hp     data.HP
	Hptype data.SetHpType
}

type gsidCardInf struct {
	Id    data.GSID
	Cards []data.CID
}

type msgType = uint8

const (
	rqsUseCard    msgType = iota //请求用牌
	rqsDropcard                  //请求弃牌
	rqsUseSkill                  //请求使用技能
	rqsTarget                    //请求可用目标
	rqsSelRole                   //发送选择的角色
	sendCard                     //发牌
	dropCard                     //广播弃牌
	useCard                      //广播用牌
	moveCard                     //广播移动牌
	useSkill                     //广播使用技能
	rspUseSkill                  //回应使用技能
	useTmpCard                   //广播使用临时卡牌
	rspTarget                    //回应可用目标
	gsidCard                     //发送gsid Card
	sendPid                      //发送pid
	playerList                   //发送玩家列表
	useAbleCard                  //发送可用卡牌
	useAbleSkill                 //发送可用技能
	dropAbleCard                 //发送可丢弃卡牌
	sendTurnOwner                //发送当前回合拥有者
	sendSkillSel                 //问玩家要不要用技能
	availableRole                //发送可选角色
	setHp                        //设置血量
	setGameState                 //设置游戏状态
	closeSignal
)
