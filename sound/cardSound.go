package sound

import (
	"goltk/data"
)

var cardSoundMale [data.Lightning][]byte

var cardSoundFemal [data.Lightning][]byte

var equipCardSound []byte

var clickCard = LoadMp3Audio("assets/card/audio/click.mp3")

// 播放卡音效
func PlayCardSound(name data.CardName, isFemal bool) {
	if name >= data.ZQYS && name <= data.DaWan {
		if equipCardSound == nil {
			equipCardSound = LoadMp3Audio("assets/card/audio/equip.mp3")
		}
		NewAudioPlayer(equipCardSound).Play()
		return
	}
	if isFemal {
		if cardSoundFemal[name-data.Attack] == nil {
			cardSoundFemal[name-data.Attack] = LoadMp3Audio("assets/card/femaleaudio/" + name.String() + ".mp3")
		}
		p := NewAudioPlayer(cardSoundFemal[name-data.Attack])
		p.Play()
	} else {
		if cardSoundMale[name-data.Attack] == nil {
			cardSoundMale[name-data.Attack] = LoadMp3Audio("assets/card/maleaudio/" + name.String() + ".mp3")
		}
		p := NewAudioPlayer(cardSoundMale[name-data.Attack])
		p.Play()
	}
}

// 播放点击卡音效
func PlayClickCard() {
	NewAudioPlayer(clickCard).Play()
}
