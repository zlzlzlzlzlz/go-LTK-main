package server

import "goltk/data"

type effect uint8

const (
	noEffect           effect = iota
	unLimit                   //无限次数
	ignorArmor                //无视防具
	unResponsive              //不可响应
	noDistance                //无视距离
	limitCard                 //限制出卡
	dropCard                  //弃置卡
	getPlayerCard             //从目标摸卡
	letPlayerFlip             //使玩家翻面
	continueAtk               //闪可追杀
	bgzEffect                 //八卦阵效果
	rwdEffect                 //仁王盾效果
	tengjiaEffect             //藤甲效果
	byszEffect                //白银狮子效果
	cxsgjEffect               //雌雄双股剑效果
	gsfEffect                 //贯石斧效果
	gddEffect                 //古锭刀效果
	hbjEffect                 //寒冰剑效果
	zqysEffect                //朱雀羽扇效果
	fthjEffect                //方天画戟效果
	qlgEffect                 //麒麟弓效果
	virtualDodgeEffect        //虚拟闪效果
	dstDffEffect              //距离减小效果
	dstUpEffect               //距离增加效果
	jiaoZiEffect              //骄姿，受到伤害或造成伤害
	liMuEffect                //立牧
	fenYinEffect              //奋音
	jiLiEffect                //蒺藜
	poJunEffect               //破军
	tunJinagEffect            //屯江
	langXiEffect              //狼袭
	fuJiEffect                //伏骑
	wenJiEffect               //问计
	lieGongEffect             //烈弓
	tuSheEffect               //图射
	tieJiEffect               //铁骑
	guiMeiEffect              //鬼魅
	diDongEffect              //地动
	guiHuoEffect              //鬼火
	shanBengEffect            //山崩
	mingBaoEffect             //冥爆
	huiHanEffect              //挥泪
	beiMinEffect              //悲鸣
	guiXinEffect              //归心
	enYuanEffect              //恩怨
	luoLeiEffect              //落雷
	zhangbaEffect             //丈八
	qilinEffect               //麒麟
	yanyueEffect              //偃月
	shiChouEffect             //誓仇
	qingGangEffect            //青钢
	yingHunEffect             //英魂
	xiaoShouEffect            //枭首
	kongChengEffect           //空城
	guanXingEffect            //观星
	qinYineffect              //琴音
	zhuQueEffect              //朱雀
	guDingEffect              //古锭
	wuShuangEffect            //无双
	fangTianEffect            //方天
	jueQingeffect             //绝情
	shangShiEffect            //殇逝
	jinQuEffect               //进趋
	qizhiEffect               //奇制
	liyuEffect                //利驭
	zhenGuEffect              //镇骨
	ziShuEffect               //自书
	yingYuanEffect            //应援
	qianXiEffect              //潜袭
	zhuiJiEffect              //追击
	baZhenEffect              //八阵
	newBaZhenEffect           //孔明八阵
	duJinEffect               //独进
	wanShaEffect              //完杀
	weiMuEffect               //帷幕
	guijiEffect               //诡计
	shequeEffect              //射却
	taipingEffect             //太平
	suomingEffect             //索命
	xixingEffect              //吸星
	baolianEffect
	renwangEffect
	manjiaEffect
	lianyuEffect
	zuijiuEffect
	qiangzhengEffect
	niepanEffect
	modaoEffect
	mojianEffect
	yushouEffect
	jueceEffect
	fankuiEffect
	danshueffect
	moYanEffect
	jijiuEffect
	xiaojiEffect
	qingnangEffect
	liangzhuEffect
	chenglueEffect
	cunmuEffect
	shicaiEffect
	wukuEffect
	sanchenEffect
	fengpoEffect
	huxiaoEffect
	wujiEffect
	miewuEffect
	quanjiEffect
	longyinEffect
	shajueEffect
	xionhuoEffect
	baoliEffect
	yisuanEffect
	quediEffect
	choujueEffect
	zhuiFengEffect
	jizhiEffect
	qicaiEffect
	changJiEffect
	jiqiaoEffect
	kejiEffect
	zhihengEffect
	tongyeEffect
	jianyingEffect
	shibeiEffect
	qiaoshiEffect
	yanyuEffect
	caichonEffect
	yingwuEffect
	lueyingEffect
	shouxiEffect
	shoudunEffect
	mojiaEffect
	longxuanEffect
	extraAtkEffect
	liexiEffect
	qinglongEffect
	shenshouEffect
)

var equipEffectMap = map[data.CardName]effect{
	data.HBJ:     hbjEffect,
	data.BGZ:     bgzEffect,
	data.TengJia: tengjiaEffect,
	data.BYSZ:    byszEffect,
	data.QLYYD:   continueAtk,
	data.QGJ:     ignorArmor,
	data.CXSGJ:   cxsgjEffect,
	data.RWD:     rwdEffect,
	data.ZGLN:    unLimit,
	data.GDD:     gddEffect,
	data.ZQYS:    zqysEffect,
	data.GSF:     gsfEffect,
	data.FTHJ:    fthjEffect,
	data.QLG:     qlgEffect,
}

var skillEffectMap = map[data.SID]effect{
	data.PaoXiaoSkill: unLimit, data.MaShuSkill: dstDffEffect,
	data.JiaoZiSkill: jiaoZiEffect, data.FenYinSkill: fenYinEffect,
	data.JiLiSkill: jiLiEffect, data.PoJunSkill: poJunEffect,
	data.TunJiangSkill: tunJinagEffect, data.LangXiSkill: langXiEffect,
	data.FuJiSkill: fuJiEffect, data.WenJiSkill: wenJiEffect,
	data.LieGongSkill: lieGongEffect, data.TuSheSkill: tuSheEffect,
	data.LiMuSkill: liMuEffect, data.TieJiSkill: tieJiEffect,
	data.GuiMeiSkill: guiMeiEffect, data.DiDongSkill: diDongEffect,
	data.ShanBengSkill: shanBengEffect, data.GuiXinSkill: guiXinEffect,
	data.BeiMingSkill: beiMinEffect, data.LuoLeiSkill: luoLeiEffect,
	data.FeiYingSkill: dstUpEffect, data.HuiLeiSkill: huiHanEffect,
	data.GuiHuoSkill: guiHuoEffect, data.MingBaoSkill: mingBaoEffect,
	data.EnYuanSkill: enYuanEffect, data.ZhangBaSkill: zhangbaEffect,
	data.ShiChouSkill: shiChouEffect, data.QingGangSkill: qingGangEffect,
	data.YingHunSkill: yingHunEffect, data.XiaoShouSkill: xiaoShouEffect,
	data.YanYueSkill: yanyueEffect, data.QiLinSkill: qilinEffect,
	data.GuanXingSkill: guanXingEffect, data.KongChengSkill: kongChengEffect,
	data.QinYinSkill: qinYineffect, data.ZhuQueSkill: zhuQueEffect,
	data.GuDingSkill: guDingEffect, data.WuShuangSkill: wuShuangEffect,
	data.FangTianSkill: fangTianEffect, data.JueQingSkill: jueQingeffect,
	data.JinQuSkill: jinQuEffect, data.QiZhiSkill: qizhiEffect,
	data.LiYuSkill: liyuEffect, data.ZhenGuSkill: zhenGuEffect,
	data.ZiShuSkill: ziShuEffect, data.YingYuanSkill: yingYuanEffect,
	data.QianXiSkill: qianXiEffect, data.ZhuiJiSkill: zhuiJiEffect,
	data.ShangShiSkill: shangShiEffect, data.BaZhenSkill: baZhenEffect,
	data.DuJinskill: duJinEffect, data.WanShaSkill: wanShaEffect,
	data.WeiMuSkill: weiMuEffect, data.LongXuanSkill: longxuanEffect,
	data.GuiJiSkill: guijiEffect, data.JiZhiSkill: jizhiEffect,
	data.TaiPingSkill: taipingEffect, data.SuoMingSkill: suomingEffect,
	data.XiXingSkill: xixingEffect, data.BaoLianSkill: baolianEffect,
	data.RenWangSkill: renwangEffect, data.ManJiaSkill: manjiaEffect,
	data.LianYuSkill: lianyuEffect, data.ZuiJiuSkill: zuijiuEffect,
	data.QiangZhengSkill: qiangzhengEffect, data.NiePanSkill: niepanEffect,
	data.ModaoSkill: modaoEffect, data.MoJianSkill: mojianEffect,
	data.YuShouSkill: yushouEffect, data.NewBaZhenSkill: newBaZhenEffect,
	data.JueCeSkill: jueceEffect, data.FanKuiSkill: fankuiEffect,
	data.DanShuSkill: danshueffect, data.MoYanSkill: moYanEffect,
	data.JiJiuSkill: jijiuEffect, data.QingNangSkill: qingnangEffect,
	data.XiaoJiSkill: xiaojiEffect, data.LiangZhuSkill: liangzhuEffect,
	data.ChengLueSkill: chenglueEffect, data.CunMuSkill: cunmuEffect,
	data.ShiCaiSkill: shicaiEffect, data.WuKuSkill: wukuEffect,
	data.SanChenSkill: sanchenEffect, data.FengPoSkill: fengpoEffect,
	data.HuXiaoSkill: huxiaoEffect, data.WuJiSkill: wujiEffect,
	data.MieWuSkill: miewuEffect, data.QuanJiSkill: quanjiEffect,
	data.LongYinSkill: longyinEffect, data.ShaJueSkill: shajueEffect,
	data.XionHuoSkill: xionhuoEffect, data.YiSuanSkill: yisuanEffect,
	data.QueDiSkill: quediEffect, data.ChouJueSkill: choujueEffect,
	data.ZhuiFengSkill: zhuiFengEffect, data.QiCaiSkill: qicaiEffect,
	data.JiQiaoSkill: jiqiaoEffect, data.KeJiSkill: kejiEffect,
	data.ZhiHengSkill: zhihengEffect, data.TongYeSkill: tongyeEffect,
	data.JianYingSkill: jianyingEffect, data.ShiBeiSkill: shibeiEffect,
	data.QiaoShiSkill: qiaoshiEffect, data.YanYuSkill: yanyuEffect,
	data.CaiChon: caichonEffect, data.YingWuSkill: yingwuEffect,
	data.LueYingSkill: lueyingEffect, data.ShouXiSkill: shouxiEffect,
	data.MoJiaSkill: mojiaEffect, data.LiexiSkill: liexiEffect,
	data.QinglongSkill: qinglongEffect, data.ShenShouSkill: shenshouEffect,
}
