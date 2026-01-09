package sound

import (
	"os"
)

var deathAudioMap = map[string][]byte{}

func GetDeathAudio(name string) []byte {
	if s, ok := deathAudioMap[name]; ok {
		return s
	}
	path := "assets/deathAudio/" + name + ".mp3"
	_, err := os.Stat(path)
	if err != nil {
		return nil
	}
	s := LoadMp3Audio(path)
	deathAudioMap[name] = s
	return s
}

func PlayDeathSound(name string) {
	a := GetDeathAudio(name)
	if a == nil {
		return
	}
	NewAudioPlayer(a).Play()
}
