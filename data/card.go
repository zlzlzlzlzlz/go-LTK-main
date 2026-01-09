package data

import "strconv"

// 卡片ID
type CID uint8

// 卡片数字
type CNum uint8

// 卡片花色
type Decor uint8

const (
	NoDec      Decor = iota //无色
	SpadeDec                //黑桃
	HeartDec                //红心
	ClubDec                 //梅花
	DiamondDec              //方片
	NoCol                   //用于在临时牌上显示无色
	RedCol                  //用于在临时牌上显示红色
	BlackCol                //用于在临时牌上显示黑色
)

func (d Decor) String() string {
	switch d {
	case SpadeDec:
		return "spade"
	case HeartDec:
		return "heart"
	case ClubDec:
		return "club"
	case DiamondDec:
		return "diamond"
	default:
		return ""
	}
}

// 花色是否为红色
func (d Decor) IsRed() bool {
	return d == HeartDec || d == DiamondDec || d == RedCol
}

// 花色是否为黑色
func (d Decor) ISBlack() bool {
	return d == SpadeDec || d == ClubDec || d == BlackCol
}

type Distence int8

var WeaponDstMap = map[CardName]Distence{
	CXSGJ: 2, GDD: 2, QLYYD: 3, HBJ: 2, GSF: 3, ZQYS: 4,
	ZGLN: 1, QGJ: 2, ZBSM: 3, FTHJ: 4, QLG: 5,
}

// 卡牌类别
type CardType uint8

const (
	BaseCardType      CardType = iota //基本卡
	TipsCardType                      //锦囊卡
	DealyTipsCardType                 //延时锦囊牌
	WeaponCardType
	ArmorCardType
	HorseUpCardType   //+1马类型
	HorseDownCardType //-1马类型
)

// 临时卡类型
type TmpCardType uint8

const (
	NotTmpCard  TmpCardType = iota //不是临时卡
	VirtualCard                    //虚拟卡
	ConvertCard                    //转化卡
)

type Card struct {
	ID       CID      //卡片ID
	CardType CardType //卡片类型
	Dec      Decor    //卡片花色
	Num      CNum     //卡片数字
	Name     CardName //卡片名字
}

type CardName uint8

const (
	NoName       CardName = iota
	Attack                //杀
	FireAttack            //火杀
	LightnAttack          //雷杀
	Dodge                 //闪
	Drunk                 //酒
	Peach                 //桃
	SSQY                  //顺手牵羊
	GHCQ                  //过河拆桥
	TYJY                  //桃园结义
	WJQF                  //万箭齐发
	NMRQ                  //南蛮入侵
	Duel                  //决斗
	WZSY                  //无中生有
	JDSR                  //借刀杀人
	WGFD                  //五谷丰登
	TSLH                  //铁索连环
	WXKJ                  //无懈可击
	Burn                  //火攻
	LBSS                  //乐不思蜀
	BLCD                  //兵粮寸断
	Lightning             //闪电
	ZQYS                  //朱雀羽扇
	FTHJ                  //方天画戟
	GDD                   //古锭刀
	HBJ                   //寒冰剑
	GSF                   //贯石斧
	CXSGJ                 //雌雄双股剑
	ZBSM                  //丈八蛇矛
	QLYYD                 //青龙偃月刀
	ZGLN                  //诸葛连弩
	QLG                   //麒麟弓
	QGJ                   //青釭剑
	BGZ                   //八卦阵
	RWD                   //仁王盾
	TengJia               //藤甲
	BYSZ                  //白银狮子
	JueYing               //绝影
	DiLU                  //的卢
	ZiJin                 //紫荆
	ChiTu                 //赤兔
	ZHFD                  //爪黄飞电
	HuaLiu                //骅骝
	DaWan                 //大宛

)

func (c CardName) String() string {
	switch c {
	case Attack:
		return "attack"
	case FireAttack:
		return "fireattack"
	case LightnAttack:
		return "lightnattack"
	case Dodge:
		return "dodge"
	case Drunk:
		return "drunk"
	case Peach:
		return "peach"
	case SSQY:
		return "ssqy"
	case GHCQ:
		return "ghcq"
	case TYJY:
		return "tyjy"
	case WJQF:
		return "wjqf"
	case NMRQ:
		return "nmrq"
	case Duel:
		return "juedou"
	case WZSY:
		return "wzsy"
	case JDSR:
		return "jdsr"
	case WGFD:
		return "wgfd"
	case TSLH:
		return "tslh"
	case WXKJ:
		return "wxkj"
	case Burn:
		return "huogon"
	case LBSS:
		return "lbss"
	case BLCD:
		return "blcd"
	case Lightning:
		return "lightning"
	case ZQYS:
		return "zqys"
	case FTHJ:
		return "fthj"
	case GDD:
		return "gdd"
	case HBJ:
		return "hbj"
	case GSF:
		return "gsf"
	case CXSGJ:
		return "cxsgj"
	case ZBSM:
		return "zbsm"
	case QLYYD:
		return "qlyyd"
	case ZGLN:
		return "zgln"
	case QLG:
		return "qlg"
	case QGJ:
		return "qgj"
	case BGZ:
		return "bgz"
	case RWD:
		return "rwd"
	case TengJia:
		return "tengjia"
	case BYSZ:
		return "bysz"
	case JueYing:
		return "jueying"
	case DiLU:
		return "dilu"
	case ZiJin:
		return "zijin"
	case ChiTu:
		return "chitu"
	case ZHFD:
		return "zhfd"
	case DaWan:
		return "dawan"
	case HuaLiu:
		return "hualiu"
	}
	panic("CardName=" + strconv.Itoa(int(c)) + "不存在")
}

func (c CardName) ChnName() string {
	switch c {
	case Attack:
		return "杀"
	case FireAttack:
		return "火杀"
	case LightnAttack:
		return "雷杀"
	case Dodge:
		return "闪"
	case Drunk:
		return "酒"
	case Peach:
		return "桃"
	case SSQY:
		return "顺手牵羊"
	case GHCQ:
		return "过河拆桥"
	case TYJY:
		return "桃园结义"
	case WJQF:
		return "万箭齐发"
	case NMRQ:
		return "南蛮入侵"
	case Duel:
		return "决斗"
	case WZSY:
		return "无中生有"
	case JDSR:
		return "借刀杀人"
	case WGFD:
		return "五谷丰登"
	case TSLH:
		return "铁索连环"
	case WXKJ:
		return "无懈可击"
	case Burn:
		return "火攻"
	case LBSS:
		return "乐不思蜀"
	case BLCD:
		return "兵粮寸断"
	case Lightning:
		return "闪电"
	case ZQYS:
		return "朱雀羽扇"
	case FTHJ:
		return "方天画戟"
	case GDD:
		return "古锭刀"
	case HBJ:
		return "寒冰剑"
	case GSF:
		return "贯石斧"
	case CXSGJ:
		return "雌雄双股剑"
	case ZBSM:
		return "丈八蛇矛"
	case QLYYD:
		return "青龙偃月刀"
	case ZGLN:
		return "诸葛连弩"
	case QLG:
		return "麒麟弓"
	case QGJ:
		return "青釭剑"
	case BGZ:
		return "八卦阵"
	case RWD:
		return "仁王盾"
	case TengJia:
		return "藤甲"
	case BYSZ:
		return "白银狮子"
	case JueYing:
		return "绝影"
	case DiLU:
		return "的卢"
	case ZiJin:
		return "紫骍"
	case ChiTu:
		return "赤兔"
	case ZHFD:
		return "爪黄飞电"
	case DaWan:
		return "大宛"
	case HuaLiu:
		return "骅骝"
	}
	panic("CardName=" + strconv.Itoa(int(c)) + "不存在")
}

var baseCardNameList = [...]CardName{Attack, FireAttack, LightnAttack, Dodge, Drunk, Peach}

var tipsCardNameList = [...]CardName{SSQY, GHCQ, TYJY, WJQF, NMRQ, Duel, WZSY, JDSR, WGFD, TSLH, WXKJ, Burn}

var dealyTipsCardNameList = [...]CardName{LBSS, BLCD, Lightning}

var weaponCardNameList = [...]CardName{ZQYS, FTHJ, GDD, HBJ, GSF, CXSGJ, ZBSM, QLYYD, ZGLN, QLG, QGJ}

var armorCardNameList = [...]CardName{BGZ, RWD, TengJia, BYSZ}

var horseUpCardNameList = [...]CardName{JueYing, DiLU, ZHFD, HuaLiu}

var horseDownCardNameList = [...]CardName{DaWan, ChiTu, ZiJin}

func isListHasName(list []CardName, name CardName) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == name {
			return true
		}
	}
	return false
}

func NewCard(name CardName, dec Decor, num CNum) Card {
	c := Card{Name: name, Dec: dec, Num: num}
	if isListHasName(baseCardNameList[:], name) {
		c.CardType = BaseCardType
		return c
	}
	if isListHasName(tipsCardNameList[:], name) {
		c.CardType = TipsCardType
		return c
	}
	if isListHasName(dealyTipsCardNameList[:], name) {
		c.CardType = DealyTipsCardType
		return c
	}
	if isListHasName(weaponCardNameList[:], name) {
		c.CardType = WeaponCardType
		return c
	}
	if isListHasName(armorCardNameList[:], name) {
		c.CardType = ArmorCardType
		return c
	}
	if isListHasName(horseUpCardNameList[:], name) {
		c.CardType = HorseUpCardType
		return c
	}
	if isListHasName(horseDownCardNameList[:], name) {
		c.CardType = HorseDownCardType
		return c
	}
	panic("无法为卡名为" + name.String() + "的卡找到匹配的类型")
}

// 获取卡片列表
func GetCards() (cards []Card) {
	//添加红心桃
	for i := CNum(3); i <= 9; i++ {
		cards = append(cards, NewCard(Peach, HeartDec, i))
	}
	cards = append(cards, NewCard(Peach, HeartDec, 2))
	cards = append(cards, NewCard(Peach, HeartDec, 12))
	//添加方片桃
	cards = append(cards, NewCard(Peach, DiamondDec, 2))
	cards = append(cards, NewCard(Peach, DiamondDec, 3))
	cards = append(cards, NewCard(Peach, DiamondDec, 12))
	//12桃
	//添加黑桃酒
	cards = append(cards, NewCard(Drunk, SpadeDec, 3))
	cards = append(cards, NewCard(Drunk, SpadeDec, 9))
	//添加梅花酒
	cards = append(cards, NewCard(Drunk, ClubDec, 3))
	cards = append(cards, NewCard(Drunk, ClubDec, 9))
	//添加方片酒
	cards = append(cards, NewCard(Drunk, DiamondDec, 9))
	//5酒
	//添加红心闪
	cards = append(cards, NewCard(Dodge, HeartDec, 2))
	cards = append(cards, NewCard(Dodge, HeartDec, 2))
	cards = append(cards, NewCard(Dodge, HeartDec, 8))
	cards = append(cards, NewCard(Dodge, HeartDec, 9))
	cards = append(cards, NewCard(Dodge, HeartDec, 11))
	cards = append(cards, NewCard(Dodge, HeartDec, 12))
	cards = append(cards, NewCard(Dodge, HeartDec, 13))
	//添加方片闪
	for i := CNum(2); i <= 11; i++ {
		cards = append(cards, NewCard(Dodge, DiamondDec, i))
	}
	cards = append(cards, NewCard(Dodge, DiamondDec, 2))
	cards = append(cards, NewCard(Dodge, DiamondDec, 6))
	cards = append(cards, NewCard(Dodge, DiamondDec, 7))
	cards = append(cards, NewCard(Dodge, DiamondDec, 8))
	cards = append(cards, NewCard(Dodge, DiamondDec, 10))
	cards = append(cards, NewCard(Dodge, DiamondDec, 11))
	cards = append(cards, NewCard(Dodge, DiamondDec, 11))
	//24闪
	//添加黑桃杀
	cards = append(cards, NewCard(Attack, SpadeDec, 7))
	for i := CNum(8); i <= 10; i++ {
		cards = append(cards, NewCard(Attack, SpadeDec, i))
		cards = append(cards, NewCard(Attack, SpadeDec, i))
	}
	//添加红心杀
	cards = append(cards, NewCard(Attack, HeartDec, 10))
	cards = append(cards, NewCard(Attack, HeartDec, 10))
	cards = append(cards, NewCard(Attack, HeartDec, 11))
	//添加梅花杀
	for i := CNum(2); i <= 11; i++ {
		cards = append(cards, NewCard(Attack, ClubDec, i))
	}
	cards = append(cards, NewCard(Attack, ClubDec, 8))
	cards = append(cards, NewCard(Attack, ClubDec, 9))
	cards = append(cards, NewCard(Attack, ClubDec, 10))
	cards = append(cards, NewCard(Attack, ClubDec, 11))
	//添加方片杀
	cards = append(cards, NewCard(Attack, DiamondDec, 13))
	for i := CNum(6); i <= 10; i++ {
		cards = append(cards, NewCard(Attack, DiamondDec, i))
	} //30张杀
	//火杀
	cards = append(cards, NewCard(FireAttack, DiamondDec, 4))
	cards = append(cards, NewCard(FireAttack, DiamondDec, 5))
	cards = append(cards, NewCard(FireAttack, HeartDec, 10))
	cards = append(cards, NewCard(FireAttack, HeartDec, 4))
	cards = append(cards, NewCard(FireAttack, HeartDec, 7))
	//雷杀
	for i := CNum(5); i <= 8; i++ {
		cards = append(cards, NewCard(LightnAttack, ClubDec, i))
	}
	for i := CNum(4); i <= 8; i++ {
		cards = append(cards, NewCard(LightnAttack, SpadeDec, i))
	}
	//添加顺手牵羊
	cards = append(cards, NewCard(SSQY, SpadeDec, 11))
	for i := CNum(3); i <= 4; i++ {
		cards = append(cards, NewCard(SSQY, SpadeDec, i))
		cards = append(cards, NewCard(SSQY, DiamondDec, i))
	} //5顺手牵羊
	//添加过河拆桥
	cards = append(cards, NewCard(GHCQ, SpadeDec, 12))
	cards = append(cards, NewCard(GHCQ, DiamondDec, 12))
	for i := CNum(3); i <= 4; i++ {
		cards = append(cards, NewCard(GHCQ, SpadeDec, i))
		cards = append(cards, NewCard(GHCQ, DiamondDec, i))
	} //6过河拆桥
	// //添加桃园结义
	cards = append(cards, NewCard(TYJY, HeartDec, 1)) //1桃园结义
	// //添加万箭齐发
	cards = append(cards, NewCard(WJQF, HeartDec, 1)) //1万箭齐发 第83
	// //添加南蛮入侵
	cards = append(cards, NewCard(NMRQ, SpadeDec, 7))
	cards = append(cards, NewCard(NMRQ, SpadeDec, 13))
	cards = append(cards, NewCard(NMRQ, ClubDec, 7)) //3南蛮入侵
	// //添加决斗
	cards = append(cards, NewCard(Duel, SpadeDec, 1))
	cards = append(cards, NewCard(Duel, ClubDec, 1))
	cards = append(cards, NewCard(Duel, DiamondDec, 1)) //3决斗
	//添加无中生有
	cards = append(cards, NewCard(WZSY, HeartDec, 11))
	for i := CNum(7); i <= 9; i++ {
		cards = append(cards, NewCard(WZSY, HeartDec, i))
	} //4无中生有
	//添加借刀杀人
	for i := CNum(12); i <= 13; i++ {
		cards = append(cards, NewCard(JDSR, ClubDec, i))
	} //2借刀杀人
	//添加五谷丰登
	for i := CNum(3); i <= 4; i++ {
		cards = append(cards, NewCard(WGFD, HeartDec, i))
	}
	//添加铁索连环
	cards = append(cards, NewCard(TSLH, SpadeDec, 11))
	cards = append(cards, NewCard(TSLH, SpadeDec, 12))
	for i := CNum(10); i <= 13; i++ {
		cards = append(cards, NewCard(TSLH, ClubDec, i))
	}
	//添加无懈可击
	cards = append(cards, NewCard(WXKJ, SpadeDec, 11))
	cards = append(cards, NewCard(WXKJ, SpadeDec, 13))
	cards = append(cards, NewCard(WXKJ, ClubDec, 12))
	cards = append(cards, NewCard(WXKJ, ClubDec, 13))
	cards = append(cards, NewCard(WXKJ, DiamondDec, 12))
	cards = append(cards, NewCard(WXKJ, HeartDec, 1))
	cards = append(cards, NewCard(WXKJ, HeartDec, 13))
	//添加火攻
	for i := CNum(2); i <= 3; i++ {
		cards = append(cards, NewCard(Burn, HeartDec, i))
	}
	cards = append(cards, NewCard(Burn, DiamondDec, 12))
	//乐不思蜀
	cards = append(cards, NewCard(LBSS, HeartDec, 6))
	cards = append(cards, NewCard(LBSS, ClubDec, 6))
	cards = append(cards, NewCard(LBSS, SpadeDec, 6))
	//兵粮寸断
	cards = append(cards, NewCard(BLCD, ClubDec, 4))
	cards = append(cards, NewCard(BLCD, SpadeDec, 10))
	//闪电
	cards = append(cards, NewCard(Lightning, HeartDec, 12))
	cards = append(cards, NewCard(Lightning, SpadeDec, 1))
	//八卦阵
	cards = append(cards, NewCard(BGZ, SpadeDec, 2))
	cards = append(cards, NewCard(BGZ, ClubDec, 2))
	//添加仁王盾
	cards = append(cards, NewCard(RWD, ClubDec, 2))
	//藤甲
	cards = append(cards, NewCard(TengJia, ClubDec, 2))
	cards = append(cards, NewCard(TengJia, SpadeDec, 2))
	///白银狮子
	cards = append(cards, NewCard(BYSZ, ClubDec, 1))
	//诸葛连弩
	cards = append(cards, NewCard(ZGLN, ClubDec, 1))
	cards = append(cards, NewCard(ZGLN, DiamondDec, 1))
	//古锭刀
	cards = append(cards, NewCard(GDD, SpadeDec, 1))
	//寒冰剑
	cards = append(cards, NewCard(HBJ, SpadeDec, 2))
	//方天画戟
	cards = append(cards, NewCard(FTHJ, DiamondDec, 12))
	//贯石斧
	cards = append(cards, NewCard(GSF, DiamondDec, 5))
	//青釭剑
	cards = append(cards, NewCard(QGJ, SpadeDec, 6))
	//丈八蛇矛
	cards = append(cards, NewCard(ZBSM, SpadeDec, 12))
	//雌雄双股剑
	cards = append(cards, NewCard(CXSGJ, SpadeDec, 2))
	//麒麟弓
	cards = append(cards, NewCard(QLG, HeartDec, 5))
	//青龙偃月刀
	cards = append(cards, NewCard(QLYYD, SpadeDec, 5))
	//朱雀羽扇
	cards = append(cards, NewCard(ZQYS, DiamondDec, 1))
	//大宛
	cards = append(cards, NewCard(DaWan, SpadeDec, 13))
	//紫骍
	cards = append(cards, NewCard(ZiJin, DiamondDec, 13))
	//赤兔
	cards = append(cards, NewCard(ChiTu, HeartDec, 5))
	//爪黄飞电
	cards = append(cards, NewCard(ZHFD, HeartDec, 13))
	//的卢
	cards = append(cards, NewCard(DiLU, ClubDec, 5))
	//绝影
	cards = append(cards, NewCard(JueYing, SpadeDec, 5))
	//骅骝
	cards = append(cards, NewCard(HuaLiu, DiamondDec, 13))
	// // 添加测试用卡牌
	// for i := 0; i <= 40; i++ {
	// 	cards = append(cards, NewCard(LBSS, DiamondDec, 3))
	// 	cards = append(cards, NewCard(GHCQ, DiamondDec, 3))
	// }
	// //为卡片分配ID
	for i := 0; i < len(cards); i++ {
		cards[i].ID = CID(i + 1)
	}
	return
}
