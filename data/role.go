package data

type Role struct {
	Name      string   //姓名
	DspName   string   //显示名字
	MaxHP     HP       //最大hp
	Female    bool     //是否是女性
	Side      RoleSide //玩家阵营
	SkillList []SID    //技能列表
	Dragon    uint8
}

// 玩家阵营
type RoleSide uint8

const (
	Qun   RoleSide = iota //群雄
	Wei                   //魏国
	Shu                   //蜀国
	Wu                    //吴国
	God                   //神
	Ye                    //野
	Ghost                 //鬼
)

func GetRoleList() []Role {
	out := make([]Role, len(RoleList))
	copy(out, RoleList)
	return out
}

var RoleList = []Role{
	{Name: "liubei", DspName: "刘备", MaxHP: 4, Side: Shu, SkillList: []SID{JieYingSkill, RenDeSkill}, Dragon: 3},
	{Name: "kongming", DspName: "孔明", MaxHP: 3, Side: Shu, SkillList: []SID{GuanXingSkill, KongChengSkill, NewBaZhenSkill}, Dragon: 1},
	{Name: "guanyu", DspName: "关羽", MaxHP: 5, Side: Shu, SkillList: []SID{YanYueSkill, WuShengSkill}, Dragon: 1},
	{Name: "zhangfei", DspName: "张飞", MaxHP: 4, Side: Shu, SkillList: []SID{PaoXiaoSkill, ZhangBaSkill}, Dragon: 1},
	{Name: "zhaoyun", DspName: "赵云", MaxHP: 4, Side: Shu, SkillList: []SID{QingGangSkill, LongDanSkill}, Dragon: 1},
	{Name: "huangzhong", DspName: "黄忠", MaxHP: 4, Side: Shu, SkillList: []SID{LieGongSkill, QiLinSkill}, Dragon: 1},
	{Name: "machao", DspName: "马超", MaxHP: 4, Side: Shu, SkillList: []SID{TieJiSkill, MaShuSkill, ShiChouSkill}, Dragon: 1},
	{Name: "lvbu", DspName: "吕布", MaxHP: 5, Side: Qun, SkillList: []SID{WuShuangSkill, LiYuSkill, FangTianSkill}, Dragon: 2},
	{Name: "ssx", DspName: "孙尚香", MaxHP: 3, Female: true, Side: Wu, SkillList: []SID{XiaoJiSkill, LiangZhuSkill, JieYinSkill}, Dragon: 3},
	{Name: "xuyou421", DspName: "许攸", MaxHP: 3, Side: Qun, SkillList: []SID{ShiCaiSkill, CunMuSkill, ChengLueSkill}, Dragon: 2},
	{Name: "huatuo", DspName: "华佗", MaxHP: 3, Side: Qun, SkillList: []SID{JiJiuSkill, QingNangSkill}, Dragon: 3},
	{Name: "xusheng491", DspName: "徐盛", MaxHP: 4, Side: Wu, SkillList: []SID{PoJunSkill}, Dragon: 2},
	{Name: "gyp115", DspName: "关银屏", MaxHP: 3, Female: true, Side: Shu, SkillList: []SID{WuJiSkill, HuXiaoSkill, XueHenSkill}, Dragon: 2},
	{Name: "myl414", DspName: "马云禄", MaxHP: 4, Female: true, Side: Shu, SkillList: []SID{FengPoSkill, MaShuSkill}, Dragon: 2},
	{Name: "maliang128", DspName: "马良", MaxHP: 3, Side: Shu, SkillList: []SID{ZiShuSkill, YingYuanSkill}, Dragon: 2},
	{Name: "lingcao71", DspName: "凌操", MaxHP: 4, Side: Wu, SkillList: []SID{DuJinskill}, Dragon: 3},
	{Name: "liuqi423", DspName: "刘琦", MaxHP: 3, Side: Qun, SkillList: []SID{WenJiSkill, TunJiangSkill}, Dragon: 2},
	{Name: "liuyan406", DspName: "刘焉", MaxHP: 3, Side: Qun, SkillList: []SID{TuSheSkill, LiMuSkill}, Dragon: 1},
	{Name: "jiaxu48", DspName: "贾诩", MaxHP: 3, Side: Qun, SkillList: []SID{WanShaSkill, WeiMuSkill, LuanWuSkill}, Dragon: 3},
	{Name: "haozhao466", DspName: "郝昭", MaxHP: 4, Side: Wei, SkillList: []SID{ZhenGuSkill}, Dragon: 2},
	{Name: "caocao205", DspName: "曹操", MaxHP: 3, Side: Wei, SkillList: []SID{GuiXinSkill, FeiYingSkill}, Dragon: 3},
	{Name: "madai167", DspName: "马岱", MaxHP: 4, Side: Shu, SkillList: []SID{QianXiSkill, ZhuiJiSkill, MaShuSkill}, Dragon: 2},
	{Name: "zch161", DspName: "张春华", MaxHP: 4, Female: true, Side: Wei, SkillList: []SID{JueQingSkill, ShangShiSkill}, Dragon: 3},
	{Name: "quyi374", DspName: "麹义", MaxHP: 4, Side: Qun, SkillList: []SID{FuJiSkill, JiaoZiSkill}, Dragon: 2},
	{Name: "liuzan132", DspName: "留赞", MaxHP: 4, Side: Wu, SkillList: []SID{FenYinSkill}, Dragon: 2},
	{Name: "wangji362", DspName: "王基", MaxHP: 3, Side: Wei, SkillList: []SID{QiZhiSkill, JinQuSkill}, Dragon: 2},
	{Name: "zhouyu203", DspName: "周瑜", MaxHP: 4, Side: Wu, SkillList: []SID{QinYinSkill, ZhuQueSkill, YeYanSkill}, Dragon: 2},
	{Name: "shamoke475", DspName: "沙摩柯", MaxHP: 4, Side: Shu, SkillList: []SID{JiLiSkill}, Dragon: 2},
	{Name: "sunjian44", DspName: "孙坚", MaxHP: 5, Side: Wu, SkillList: []SID{YingHunSkill, GuDingSkill}, Dragon: 2},
	{Name: "lijue426", DspName: "李傕", MaxHP: 4, Side: Qun, SkillList: []SID{LangXiSkill, YiSuanSkill}, Dragon: 3},
	{Name: "zhonhui173", DspName: "钟会", MaxHP: 4, Side: Ye, SkillList: []SID{QuanJiSkill, PaiYiSkill}, Dragon: 1},
	{Name: "duyu552", DspName: "杜预", MaxHP: 4, Side: Qun, SkillList: []SID{WuKuSkill, SanChenSkill, MieWuSkill}, Dragon: 1},
	{Name: "xuron428", DspName: "徐荣", MaxHP: 4, Side: Qun, SkillList: []SID{ShaJueSkill, XionHuoSkill}, Dragon: 2},
	{Name: "wenyuan580", DspName: "文鸯", MaxHP: 4, Side: Wei, SkillList: []SID{QueDiSkill, ChouJueSkill, ZhuiFengSkill}, Dragon: 1},
	{Name: "hyy647", DspName: "黄月英", MaxHP: 3, Side: Shu, Female: true, SkillList: []SID{JiZhiSkill, QiCaiSkill, JiQiaoSkill}, Dragon: 2},
	{Name: "lvmeng", DspName: "吕蒙", MaxHP: 4, Side: Wu, SkillList: []SID{KeJiSkill}, Dragon: 2},
	{Name: "sunquan", DspName: "孙权", MaxHP: 4, Side: Wu, SkillList: []SID{TongYeSkill, ZhiHengSkill}, Dragon: 3},
	{Name: "jushou604", DspName: "沮授", MaxHP: 3, Side: Qun, SkillList: []SID{ShiBeiSkill, JianYingSkill}, Dragon: 3},
	{Name: "xhs712", DspName: "夏侯氏", MaxHP: 3, Side: Shu, Female: true, SkillList: []SID{QiaoShiSkill, YanYuSkill}, Dragon: 3},
	{Name: "liucheng609", DspName: "刘赪", MaxHP: 3, Side: Qun, SkillList: []SID{LueYingSkill, YingWuSkill}, Dragon: 2},
}

var GhostList = []Role{
	{Name: "ghost/chi", DspName: "魑", MaxHP: 5, Side: Ghost, SkillList: []SID{GuiMeiSkill, DiDongSkill, ShanBengSkill}},
	{Name: "ghost/mei", DspName: "魅", MaxHP: 5, Side: Ghost, SkillList: []SID{GuiMeiSkill, BeiMingSkill, EnYuanSkill}},
	{Name: "ghost/wang", DspName: "魍", MaxHP: 5, Side: Ghost, SkillList: []SID{GuiMeiSkill, LuoLeiSkill, HuiLeiSkill}},
	{Name: "ghost/liang", DspName: "魉", MaxHP: 5, Side: Ghost, SkillList: []SID{GuiMeiSkill, GuiHuoSkill, MingBaoSkill}},
	{Name: "ghost/niutou", DspName: "牛头", MaxHP: 7, Side: Ghost, SkillList: []SID{BaoLianSkill, NiePanSkill, ManJiaSkill, XiaoShouSkill}},
	{Name: "ghost/mamian", DspName: "马面", MaxHP: 6, Side: Ghost, SkillList: []SID{GuiJiSkill, JueCeSkill, LianYuSkill, FanKuiSkill}},
	{Name: "ghost/heiwuchang", DspName: "黑无常", MaxHP: 9, Side: Ghost, SkillList: []SID{GuiJiSkill, XiXingSkill, SuoMingSkill, TaiPingSkill}},
	{Name: "ghost/baiwuchang", DspName: "白无常", MaxHP: 9, Side: Ghost, SkillList: []SID{BaoLianSkill, ZuiJiuSkill, JueCeSkill, QiangZhengSkill}},
	{Name: "ghost/yecha", DspName: "夜叉", MaxHP: 11, Side: Ghost, SkillList: []SID{MoJianSkill, ModaoSkill, BaZhenSkill, DanShuSkill}},
	{Name: "ghost/luocha", DspName: "罗刹", MaxHP: 12, Female: true, Side: Ghost, SkillList: []SID{RenWangSkill, ModaoSkill, YuShouSkill, MoYanSkill}},
}

var GroupGhostList = []Role{
	{Name: "ghost/chimei", DspName: "魑&魅", MaxHP: 5, Side: Ghost, SkillList: []SID{GuiMeiSkill, DiDongSkill, ShanBengSkill, BeiMingSkill, EnYuanSkill}},
	{Name: "ghost/wangliang", DspName: "魍&魉", MaxHP: 5, Side: Ghost, SkillList: []SID{GuiMeiSkill, LuoLeiSkill, HuiLeiSkill, GuiHuoSkill, MingBaoSkill}},
	{Name: "ghost/niutoumamian", DspName: "牛头&马面", MaxHP: 7, Side: Ghost, SkillList: []SID{BaoLianSkill, NiePanSkill, ManJiaSkill, XiaoShouSkill, GuiJiSkill, JueCeSkill, LianYuSkill, FanKuiSkill}},
	{Name: "ghost/heibaiwuchang", DspName: "黑&白无常", MaxHP: 9, Side: Ghost, SkillList: []SID{GuiJiSkill, XiXingSkill, SuoMingSkill, TaiPingSkill, BaoLianSkill, ZuiJiuSkill, JueCeSkill, QiangZhengSkill}},
	{Name: "ghost/yechaluocha", DspName: "夜叉&罗刹", MaxHP: 11, Side: Ghost, SkillList: []SID{MoJianSkill, ModaoSkill, BaZhenSkill, DanShuSkill, RenWangSkill, YuShouSkill, MoYanSkill}},
}

var MonsterList = []Role{
	{Name: "monster/nianshou", DspName: "年兽", MaxHP: 40, Side: Qun, SkillList: []SID{YuShouSkill, ShouXiSkill, MoJiaSkill}},
	{Name: "caichon", DspName: "菜虫", MaxHP: 66, Side: Wu, SkillList: []SID{CaiChon}},
	{Name: "monster/qinglong", DspName: "青龙", MaxHP: 20, Side: Qun, SkillList: []SID{LongXuanSkill, LiexiSkill, QinglongSkill, ShenShouSkill}},
}
