package client

import (
	"fmt"
	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/level/block"
)

type Point struct {
	x, y, z int
}

type PlayerPosData struct {
	x,
	y,
	z,
	yaw,
	pitch float64
}

type Movement struct {
	Bot              *FragBot
	MovingForward    bool
	Jumping          bool
	Position         PlayerPosData
	PastSpeed        float64
	PastSlipperiness float64
}

func NewMovement(fb *FragBot) *Movement {
	return &Movement{
		Bot:              fb,
		MovingForward:    false,
		Jumping:          false,
		Position:         PlayerPosData{},
		PastSpeed:        0,
		PastSlipperiness: 0.6,
	}
}

// Called every tick (~20 times a second)
func (m *Movement) move() {

}

func (m *Movement) GetBlock(pos Point) (block.Block, error) {
	w := m.Bot.botWorld
	chunkPos := level.ChunkPos{int32(pos.x) >> 4, int32(pos.z) >> 4}
	if chunk, ok := w.Columns[chunkPos]; ok {
		x, y, z := pos.x, pos.y, pos.z
		y += 64
		if y < 0 || y >= len(chunk.Sections)*16 {
			return block.Air{}, fmt.Errorf("y=%d out of bounds", y)
		}
		if t := chunk.Sections[y>>4]; t.States != nil {
			return block.StateList[t.States.Get(y&15<<8|z&15<<4|x&15)], nil
		} else {
			return block.Air{}, fmt.Errorf("y=%d out of bounds", y)
		}
	} else {
		return block.Air{}, fmt.Errorf("chunk not found")
	}
}
