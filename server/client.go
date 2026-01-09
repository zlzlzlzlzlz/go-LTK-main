package server

import (
	"goltk/data"
	"time"
)

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
	GetClientType() data.PlayerType                       //获取客户端类型
	Close()                                               //关闭客户端
}
