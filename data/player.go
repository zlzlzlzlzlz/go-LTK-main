package data

type PlayerType uint8

const (
	NoPlayer PlayerType = iota
	LocalPlayer
	RemotePlayer
	BotPlayer
)

func (p PlayerType) Name() string {
	switch p {
	case NoPlayer:
		return "无"
	case LocalPlayer:
		return "本地玩家"
	case RemotePlayer:
		return "远程玩家"
	case BotPlayer:
		return "人机玩家"
	}
	return ""
}

// 玩家ID
type PID int8

func (p PID) Name() string {
	switch p {
	case 0:
		return "一"
	case 1:
		return "二"
	case 2:
		return "三"
	case 3:
		return "四"
	case 4:
		return "五"
	case 5:
		return "六"
	case 6:
		return "七"
	case 7:
		return "八"
	default:
	}
	return ""
}

// 玩家HP
type HP int8

// 使用卡的信息
type UseCardInf struct {
	Skip       bool
	ID         CID
	TargetList []PID
}

// 使用技能的信息
type UseSkillInf struct {
	Skip       bool
	ID         SID
	TargetList []PID
	Cards      []CID
	Args       []byte
}

// 请求使用主动技能的回应
type UseSkillRsp struct {
	ID      SID
	Targets []PID
	Cards   []CID
	Args    []byte
}

// 可攻击目标信息
type AvailableTargetInf struct {
	TargetNum  uint8 //可攻击的目标数量
	TargetList []PID
}

// 玩家信息
type PlayerInf struct {
	Role Role
	PID  PID
}

// 弃卡信息
type DropCardInf []CID

type EquipSlot uint8

const (
	WeaponSlot    EquipSlot = iota //武器槽
	ArmorSlot                      //装甲槽
	HorseUpSlot                    //+1马槽
	HorseDownSlot                  //-1马槽
)

type JudgeSlot uint8

const (
	LBSSSlot      JudgeSlot = iota //乐不思蜀槽
	BLCDSlot                       //兵粮寸断槽
	LightningSlot                  //闪电槽
)

type SetHpType uint8

const (
	NormalDmg    SetHpType = iota //普通伤害
	FireDmg                       //火焰伤害
	LightningDmg                  //雷电伤害
	BleedingDmg                   //流血伤害
	Recover                       //恢复
	DffHPMax                      //扣血上限
	AddHpMax
)
