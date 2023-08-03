package client

import (
	"fmt"
	"github.com/Prince/fragbots/logging"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/level/block"
	pk "github.com/Tnze/go-mc/net/packet"
	"math"
	"strings"
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
	Bot                      *FragBot
	MovingForward            bool
	Jumping                  bool
	Position                 PlayerPosData
	PastSpeed                float64
	PastSlipperiness         float64
	HorizontalMovementTarget Point
	OnGround                 bool
	OnCarpet                 bool
	Sprinting                bool
	PastSprinting            bool
}

func NewMovement(fb *FragBot) *Movement {
	m := &Movement{
		Bot:           fb,
		MovingForward: false,
		Jumping:       false,
		Position: PlayerPosData{
			x:     0,
			y:     0,
			z:     0,
			yaw:   0,
			pitch: 0,
		},
		PastSpeed:        0,
		PastSlipperiness: 0.6,
		OnGround:         true,
		OnCarpet:         false,
		Sprinting:        false,
		PastSprinting:    false,
	}

	//Load player position event
	fb.client.Events.AddListener(
		bot.PacketHandler{
			ID:       packetid.ClientboundPlayerPosition,
			Priority: 64,
			F:        m.synchronizePlayerPosition,
		},
	)

	return m
}

// Called every tick (~20 times a second)
func (m *Movement) move() error {
	//If unable to find block at food then bot is not loaded in
	footBlock, err := m.GetBlockAtFeet()
	if err != nil {
		return nil
	}

	x, y, z, yaw, pitch := m.Position.x, m.Position.y, m.Position.z, m.Position.yaw, m.Position.pitch
	if err := m.Bot.client.Conn.WritePacket(pk.Marshal(
		packetid.ServerboundMovePlayerPosRot,
		pk.Double(x),
		pk.Double(y),
		pk.Double(z),
		pk.Float(yaw),
		pk.Float(pitch),
		pk.Boolean(m.OnGround),
	),
	); err != nil {
		logging.LogWarn("Error sending ServerboundMovePlayer:", err)
		return err
	}

	m.OnGround = !block.IsAir(block.ToStateID[footBlock])

	err = m.handleSprinting()
	if err != nil {
		return err
	}

	m.handleMovement()

	return nil
}

// If returns false then movement is cancelled
func (m *Movement) checkFrontBack() bool {
	playerPos := m.Position

	yawInRadians := m.Position.yaw * 3.14159265358979 / 180

	frontPos := Point{
		x: int(playerPos.x + 0.3*math.Sin(yawInRadians*-1)),
		y: int(playerPos.y),
		z: int(playerPos.z + 0.3*math.Cos(yawInRadians)),
	}
	backPos := Point{
		x: int(playerPos.x - 0.3*math.Sin(yawInRadians*-1)),
		y: int(playerPos.y),
		z: int(playerPos.z - 0.3*math.Cos(yawInRadians)),
	}

	blockInFront, err := m.GetBlock(frontPos)
	if err != nil {
		logging.LogWarn("failed to get block infront:", err)
	}

	blockInBack, err := m.GetBlock(backPos)

	if strings.Contains(blockInFront.ID(), "carpet") || strings.Contains(blockInBack.ID(), "carpet") {
		if !m.OnCarpet {
			m.Position.y += 0.0625
			m.OnCarpet = true
		}
		return true
	} else if m.OnCarpet {
		m.Position.y -= 0.0625
		m.OnCarpet = false
	}

	return blockInFront.ID() == "minecraft:light" || blockInFront.ID() == "minecraft:air"
}

func (m *Movement) handleSprinting() error {
	var err error = nil
	if m.PastSprinting && !m.MovingForward {
		err = m.sendSprintingPacket(false)
		m.PastSprinting = false
		m.Sprinting = false
	}
	if m.Sprinting != m.PastSprinting {
		err = m.sendSprintingPacket(m.Sprinting)
		m.PastSprinting = m.Sprinting
	}

	if err != nil {
		logging.Log("Failed to send sprinting packet:", err)
	}
	return err
}

func (m *Movement) handleMovement() {
	if m.MovingForward {
		motionX, motionZ := m.moveForward()

		if !m.checkFrontBack() {
			m.Position.x -= motionX
			m.Position.z -= motionZ
			m.PastSpeed = 0
		}
	} else {
		m.PastSpeed = 0
	}
}

// https://www.mcpk.wiki/wiki/Horizontal_Movement_Formulas
func (m *Movement) moveForward() (float64, float64) {
	momentum := m.PastSpeed * m.PastSlipperiness * 0.91

	movementMult := 1.0
	if m.Sprinting {
		movementMult = 1.3
	}

	// Ignore effects and slipperiness
	acceleration := 0.1 * movementMult

	if m.Jumping {
		acceleration += 0.2
	} else if !m.OnGround {
		acceleration = 0.02 * movementMult
	}

	speed := momentum + acceleration

	yawInRadians := m.Position.yaw * 3.14159265358979 / 180

	// Positive X is -90 yaw so needs to be reversed
	dx := speed * math.Sin(yawInRadians*-1)
	dz := speed * math.Cos(yawInRadians)

	m.Position.x += dx
	m.Position.z += dz

	m.PastSpeed = speed

	return dx, dz
}

func (m *Movement) synchronizePlayerPosition(packet pk.Packet) error {
	var (
		x,
		y,
		z pk.Double
		yaw,
		pitch pk.Float
	)

	err := packet.Scan(&x, &y, &z, &yaw, &pitch)
	if err != nil {
		logging.LogWarn("Failed to scan packet data", err.Error())
		return err
	}

	logging.Log("teleported to x:", x, ", y:", y, ", z:", z, "yaw:", yaw, ", pitch", pitch)

	m.Position = PlayerPosData{
		x:     float64(x),
		y:     float64(y),
		z:     float64(z),
		yaw:   float64(yaw),
		pitch: float64(pitch),
	}
	return nil
}

func (m *Movement) GetBlockAtFeet() (block.Block, error) {
	playerPos := m.Position
	feetPos := Point{
		x: int(playerPos.x),
		y: int(playerPos.y - 0.2),
		z: int(playerPos.z),
	}
	return m.GetBlock(feetPos)
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

func (m *Movement) sendSprintingPacket(sprinting bool) error {
	action := 3
	if !sprinting {
		action = 4
	}

	return m.Bot.client.Conn.WritePacket(pk.Marshal(
		packetid.ServerboundPlayerCommand,
		pk.VarInt(m.Bot.player.EID),
		pk.VarInt(action),
		pk.VarInt(0),
	),
	)
}
