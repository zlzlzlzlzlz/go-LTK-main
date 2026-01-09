package localclient

import (
	"goltk/data"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// 错误处理
func errHandle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 加载图片
func loadImg(path string) *ebiten.Image {
	img, _, err := ebitenutil.NewImageFromFile(path)
	errHandle(err)
	return img
}

var imgBufMap = map[string]*ebiten.Image{}

// 获取图片，这将会缓存
func getImg(path string) *ebiten.Image {
	if img, ok := imgBufMap[path]; ok {
		return img
	}
	img := loadImg(path)
	imgBufMap[path] = img
	return img
}

// 矩形类
type rect struct {
	x, y            int
	wide, height    int
	hasVirtualClick bool
}

// 检测点是否在矩形内
func (r *rect) isPointInRect(x, y int) bool {
	return (x > r.x && x < r.x+r.wide) && (y > r.y && y < r.y+r.height)
}

// 检测是否被点击
func (r *rect) isClicked() bool {
	if r.hasVirtualClick {
		r.hasVirtualClick = false
		return true
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return false
	}
	return r.isPointInRect(ebiten.CursorPosition())
}

// 虚拟点击
func (r *rect) virtualClick() {
	r.hasVirtualClick = true
}

// 检测list中是否含有item
func isItemInList[T comparable](list []T, item ...T) bool {
	for i := 0; i < len(list); i++ {
		for j := 0; j < len(item); j++ {
			if list[i] == item[j] {
				return true
			}
		}
	}
	return false
}

func findCardById(list []cardI, id data.CID) cardI {
	for _, card := range list {
		if card.getID() == id {
			return card
		}
	}
	return nil
}

func getCardById(list []cardI, id data.CID) ([]cardI, cardI) {
	for i := 0; i < len(list); i++ {
		if list[i].getID() == id {
			c := list[i]
			return append(list[:i], list[i+1:]...), c
		}
	}
	return list, nil
}

func iterators[T any](list []T, fn func(T)) {
	for i := 0; i < len(list); i++ {
		fn(list[i])
	}
}

func delFromList[T any](list []T, ok func(T) bool) []T {
	for i := 0; i < len(list); i++ {
		if ok(list[i]) {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}

// 计算卡牌位置x,y为起始点,gap为默认间距,卡片占据总长达到maxLen后开始折叠,启用tp则会瞬移
func calculateCardPos(x, y, gap, maxLen float64, tp bool, cards ...cardI) {
	if float64(len(cards))*gap > maxLen {
		gap = maxLen / float64(len(cards))
	}
	for _, c := range cards {
		if tp {
			c.setPos(x, y)
		} else {
			c.move2Pos(x, y, nil)
		}
		x += gap
	}
}
