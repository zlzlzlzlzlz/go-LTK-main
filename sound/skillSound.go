package sound

import (
	"goltk/data"
	"math/rand"
	"os"
	"strconv"
)

var skillAudioMap1 = map[data.SID][]byte{}

var skillAudioMap0 = map[data.SID][]byte{}

func GetSkillAudio(sid data.SID) []byte {
	a := rand.Intn(2)
	if a == 0 {
		if s, ok := skillAudioMap0[sid]; ok {
			return s
		}
		path := "assets/skillAudio/" + sid.String() + strconv.Itoa(a) + ".mp3"
		_, err := os.Stat(path)
		if err != nil {
			return nil
		}
		s := LoadMp3Audio(path)
		skillAudioMap0[sid] = s
		return s
	} else {
		if s, ok := skillAudioMap1[sid]; ok {
			return s
		}
		path := "assets/skillAudio/" + sid.String() + strconv.Itoa(a) + ".mp3"
		_, err := os.Stat(path)
		if err != nil {
			return nil
		}
		s := LoadMp3Audio(path)
		skillAudioMap1[sid] = s
		return s
	}
}

func PlaySkillSound(sid data.SID) {
	a := GetSkillAudio(sid)
	if a == nil {
		return
	}
	NewAudioPlayer(a).Play()
}
