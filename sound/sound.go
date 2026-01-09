package sound

import (
	"io"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

var audioContext = audio.NewContext(44100)

type Player = audio.Player

// 错误处理
func errHandle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 加载mp3音频,这将不会缓存
func LoadMp3Audio(path string) []byte {
	fd, err := os.Open(path)
	errHandle(err)
	defer fd.Close()
	dec, err := mp3.DecodeWithoutResampling(fd)
	errHandle(err)
	out, err := io.ReadAll(dec)
	errHandle(err)
	return out
}

var audioMap = map[string][]byte{}

// 获取音频,这将会缓存
func GetMp3Audio(path string) []byte {
	if s, ok := audioMap[path]; ok {
		return s
	}
	s := LoadMp3Audio(path)
	audioMap[path] = s
	return s
}

// 播放音频,这将会缓存
func PlayAudio(path string) {
	p := NewAudioPlayer(GetMp3Audio(path))
	p.Play()
}

var bgmMap = map[string]*os.File{}

// 新建bgm用于应对长音乐
func NewBgm(path string) *Player {
	if fd, ok := bgmMap[path]; ok {
		fd.Close()
	}
	fd, err := os.Open(path)
	errHandle(err)
	dec, err := mp3.DecodeWithoutResampling(fd)
	errHandle(err)
	bgmMap[path] = fd
	return newInfLoop(dec)
}

// 新建无限循环
func newInfLoop(s *mp3.Stream) *Player {
	infLoop := audio.NewInfiniteLoop(s, s.Length())
	p, err := audioContext.NewPlayer(infLoop)
	errHandle(err)
	return p
}

// 新建音乐播放器
func NewAudioPlayer(s []byte) *Player {
	p := audioContext.NewPlayerFromBytes(s)
	return p
}
