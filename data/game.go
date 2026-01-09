package data

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type GameState uint8

const (
	InitState          GameState = iota //初始化阶段
	PrepareState                        //准备阶段
	JudgedState                         //判定阶段
	SendCardState                       //发牌环节
	UseCardState                        //用牌环节
	DropCardState                       //弃牌环节
	EndState                            //结束环节
	DodgeState                          //闪阶段
	SetHpState                          //修改血量阶段
	DyingState                          //濒死阶段
	DieState                            //死掉阶段
	WXKJState                           //无懈可击阶段
	DuelState                           //决斗阶段
	NMRQState                           //南蛮入侵阶段
	WJQFState                           //万箭齐发阶段
	TYJYState                           //桃园结义阶段
	WGFDState                           //五谷丰登阶段
	BurnShowState                       //火攻展示阶段
	BurnDropState                       //火攻弃置阶段
	SSQYState                           //顺手牵羊阶段
	GHCQState                           //过河拆桥阶段
	JDSRState                           //借刀杀人阶段
	MakeJudgeState                      //进行判定
	SkillSelectState                    //技能选择阶段
	SkillJudgeState                     //技能判定阶段
	DropSelfAllCards                    //丢弃自己牌(包括装备)阶段
	QlYYDState                          //追杀阶段
	CXSGJState                          //雌雄双股剑
	SkillDropState                      //通用弃手牌阶段
	QLGState                            //麒麟弓
	DropOtherCardState                  //丢别人牌阶段
	DropSelfHandCard                    //弃自己手牌
	PoJunState                          //破军
	WenJiState                          //问计
	EnYuanState                         //恩怨
	YingHunState                        //英魂
	GuanXingState                       //观星
	QinYinState                         //琴音
	LiyuState                           //利驭
	FanKuiState                         //反馈
	ChangeRoleState                     //换人阶段
	FengPoState                         //凤魄
	LiangZhuState                       //良助
	GSFState                            //贯石斧
	QueDiState                          //却敌
	WinState                            //胜利
	LoseState                           //失败
	JiQiaoState                         //机巧
	LuanWuState                         //乱武
	YanYuState                          //燕语
	GetSkillState                       //自选技能阶段
	ReConnState                         //玩家重新连接
)

func (state GameState) String() string {
	switch state {
	case InitState:
		return "init"
	case PrepareState:
		return "prepare"
	case JudgedState:
		return "judge"
	case SendCardState:
		return "sendCard"
	case UseCardState:
		return "useCard"
	case DropCardState:
		return "dropCard"
	case EndState:
		return "end"
	case DodgeState:
		return "dodge"
	case DyingState:
		return "dying"
	case DieState:
		return "die"
	case WXKJState:
		return "wxkj"
	case DuelState:
		return "duel"
	case NMRQState:
		return "nmrq"
	case WJQFState:
		return "wjqf"
	case TYJYState:
		return "tyjy"
	case WGFDState:
		return "wgfd"
	case BurnShowState:
		return "burnShow"
	case BurnDropState:
		return "burnDrop"
	case SetHpState:
		return "setHP"
	case SSQYState:
		return "ssqy"
	case GHCQState:
		return "ghcq"
	case JDSRState:
		return "jdsr"
	case MakeJudgeState:
		return "makeJudge"
	case SkillSelectState:
		return "skillSelect"
	case SkillJudgeState:
		return "skillJudge"
	case DropSelfAllCards:
		return "dropSelfAll"
	case QlYYDState:
		return "ctnAtk"
	case QLGState:
		return "qlg"
	case CXSGJState:
		return "cxsgj"
	case DropOtherCardState:
		return "dropOtherCard"
	case DropSelfHandCard:
		return "dropSelfHandCard"
	case PoJunState:
		return "pojun"
	case EnYuanState:
		return "enyuan"
	case GuanXingState:
		return "guanxing"
	case FengPoState:
		return "fengpo"
	case LiangZhuState:
		return "liangzu"
	case GSFState:
		return "gsf"
	case QueDiState:
		return "quedi"
	case WinState:
		return "win"
	case LoseState:
		return "lose"
	case LuanWuState:
		return "luanwu"
	case GetSkillState:
		return "getSkill"
	case ReConnState:
		return "reConn"
	default:
		return ""
	}
}

func (state GameState) Name() string {
	switch state {
	case InitState:
		return "初始化"
	case PrepareState:
		return "准备"
	case JudgedState:
		return "判定"
	case SendCardState:
		return "发牌"
	case UseCardState:
		return "出牌"
	case DropCardState:
		return "弃牌"
	case EndState:
		return "结束"
	case DodgeState:
		return "闪"
	case DyingState:
		return "濒死"
	case DieState:
		return "死亡"
	case WXKJState:
		return "无懈可击"
	case DuelState:
		return "决斗"
	case NMRQState:
		return "南蛮入侵"
	case WJQFState:
		return "万箭齐发"
	case TYJYState:
		return "桃园结义"
	case WGFDState:
		return "五谷丰登"
	case BurnShowState:
		return "火攻展示"
	case BurnDropState:
		return "火攻弃置"
	case SetHpState:
		return "血量更改"
	case SSQYState:
		return "顺手牵羊"
	case GHCQState:
		return "过河拆桥"
	case JDSRState:
		return "借刀杀人"
	case MakeJudgeState:
		return "判定"
	case SkillSelectState:
		return "技能选择"
	case SkillJudgeState:
		return "技能判定"
	case DropSelfAllCards:
		return "贯石斧"
	case QlYYDState:
		return "追杀"
	case QLGState:
		return "麒麟弓"
	case CXSGJState:
		return "雌雄双股剑"
	case DropOtherCardState:
		return "丢别人牌"
	case DropSelfHandCard:
		return "丢自己手牌"
	case PoJunState:
		return "破军"
	case EnYuanState:
		return "恩怨"
	case GuanXingState:
		return "观星"
	case LiangZhuState:
		return "良助"
	case GSFState:
		return "贯石斧"
	case QueDiState:
		return "却敌"
	case WinState:
		return "胜利"
	case LoseState:
		return "失败"
	case LuanWuState:
		return "乱武"
	case GetSkillState:
		return "选技能"
	case ReConnState:
		return "重新连接"
	}
	return ""
}

var stateTextMap map[string]string

var stateTextMapOther map[string]string

func (state GameState) GetText() string {
	if stateTextMap == nil {
		stateTextMap = map[string]string{}
		fd, err := os.Open("assets/stateText.json")
		if err != nil {
			log.Fatal(err)
		}
		b, err := io.ReadAll(fd)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(b, &stateTextMap)
		if err != nil {
			log.Fatal(err)
		}
	}
	return stateTextMap[state.String()]
}

func (state GameState) GetTextOther() string {
	if stateTextMapOther == nil {
		stateTextMapOther = map[string]string{}
		fd, err := os.Open("assets/stateTextOther.json")
		if err != nil {
			log.Fatal(err)
		}
		b, err := io.ReadAll(fd)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(b, &stateTextMapOther)
		if err != nil {
			log.Fatal(err)
		}
	}
	return stateTextMapOther[state.String()]
}

type GSID uint8

const (
	GSIDWGFD          GSID = iota //用于五谷丰登阶段的id
	GSIDBurn                      //用于火攻阶段的id
	GSIDSSQY                      //用于顺手牵羊的id
	GSIDGHCQ                      //过河拆桥id
	GSIDPerJudge                  //发送要判定的卡牌
	GSIDJudge                     //发送判定结果
	GSIDPerSkillJudge             //发送要判定的技能
	GSIDSkillJudge                //发送判定结果
	GSIDQLG                       //麒麟弓id
	GSIDDropOtrCard               //丢别人牌
)

type SpecialPID = PID

const SpecialPIDGame SpecialPID = -128

type GameMode uint8

const (
	QGZXEasyMode GameMode = iota //驱鬼逐邪模式
	QGZXNormalMode
	QGZXHardMode
	QGZXVeryHardMode
	QGZXDoubleMode
	QGZXFreeMode
	NianShouMode //年兽
)
