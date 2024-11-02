package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Bullet struct {
	x, y float32
}

type Enemy struct {
	x, y float32
}

type Sound struct {
	name   string
	source []byte
}

type SoundPlayer struct {
	audioContext *audio.Context
	players      map[string]*audio.Player
}

type Game struct {
	playerX, playerY float32
	bullets          []*Bullet
	lastFireTime     time.Time
	enemies          []*Enemy
	enemiesMoveRight bool
	enemiesMoveDown  bool
	enemiesSpeed     float32
	gameStatus       int
	gameLevel        int
	soundPlayer      *SoundPlayer
}

var (
	err error

	//go:embed assets/sound/laserShoot.wav
	shootSoundByte []byte

	//go:embed assets/sound/hitHurt.wav
	hitSoundByte []byte

	//go:embed assets/sound/explosion.wav
	explosionSoundByte []byte

	//go:embed assets/sound/enemyStepDown.wav
	enemyStepDownSoundByte []byte

	//go:embed assets/sound/won.wav
	wonSoundByte []byte

	background *ebiten.Image
)

const (
	// misc
	gameTitle = "Ebit Invaders"

	// screen
	screenWidth  = 640
	screenHeight = 480

	// sounds
	sampleRate         = 44100
	soundShoot         = "shoot"
	soundHit           = "hit"
	soundExplosion     = "explosion"
	soundEnemyStepDown = "stepDown"
	soundWon           = "won"

	// game statuses
	gameOngoing     = 0
	gameWon         = 1
	gameLost        = 2
	gameNextLevel   = 4
	gameStartScreen = 5 // TODO: implement start screen

	// game levels
	gameLevel1 = 1
	gameLevel2 = 2
	gameLevel3 = 3
	gameLevel4 = 4
	gameLevel5 = 5

	// player
	playerSize      = 12
	playerStartX    = screenWidth / 2
	playerStartY    = float32(screenHeight - playerSize - 10)
	playerMoveSpeed = 2

	// player shooting
	bulletWidth  = 2
	bulletHeight = 4
	bulletSpeed  = 5
	fireCooldown = 1 * time.Second

	// enemy
	enemySize  = 20
	enemySpace = 20
)

func init() {
	background, _, err = ebitenutil.NewImageFromFile("assets/img/bgSpace.png")
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update() error {
	g.handleKeyPress()

	if g.gameStatus != gameOngoing {
		return nil
	}

	g.handleGameStatusChange()
	g.handleBulletMove()
	g.handleEnemyMove()
	g.handleCollisions()
	return nil
}

func (g *Game) startScreenKeys() {
	if g.gameStatus != gameStartScreen {
		return
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		g.gameStatus = gameOngoing
	}
}

func (g *Game) onGoingKeys() {
	if g.gameStatus != gameOngoing {
		return
	}
	// player keys
	// move left
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) && g.playerX > 0 {
		g.playerX -= playerMoveSpeed
	}
	// move right
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) && (g.playerX+playerSize) < screenWidth {
		g.playerX += playerMoveSpeed
	}
	// fire bullets
	if ebiten.IsKeyPressed(ebiten.KeySpace) && time.Since(g.lastFireTime) >= fireCooldown {
		// fire a new bullet and update the last fire time
		g.bullets = append(g.bullets, &Bullet{x: g.playerX + playerSize/2 - bulletWidth/2, y: g.playerY})
		g.soundPlayer.PlaySound(soundShoot)
		g.lastFireTime = time.Now()
	}
	// fire chaingun
	if ebiten.IsKeyPressed(ebiten.KeyC) && time.Since(g.lastFireTime) >= time.Millisecond*10 {
		// fire a new bullet and update the last fire time
		g.bullets = append(g.bullets, &Bullet{x: g.playerX + playerSize/2 - bulletWidth/2, y: g.playerY})
		g.soundPlayer.PlaySound(soundShoot)
		g.lastFireTime = time.Now()
	}
}

func (g *Game) lostKeys() {
	if g.gameStatus != gameLost {
		return
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		g.restartGame()
	}
}

func (g *Game) nextLevelKeys() {
	if g.gameStatus != gameNextLevel {
		return
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		g.startNextLevel()
	}
}

func (g *Game) wonKeys() {
	if g.gameStatus != gameWon {
		return
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		// TODO: create end screen with LOST and WON
		g.restartGame()
	}
}

func (g *Game) handleKeyPress() {
	g.startScreenKeys()
	g.onGoingKeys()
	g.lostKeys()
	g.nextLevelKeys()
	g.wonKeys()
}

func (g *Game) handleBulletMove() {
	// Update bullets' position
	var activeBullets []*Bullet
	for _, bullet := range g.bullets {
		bullet.y -= bulletSpeed
		// Keep bullets that are on screen
		if bullet.y+bulletHeight > 0 {
			activeBullets = append(activeBullets, bullet)
		}
	}
	g.bullets = activeBullets
}

func (g *Game) handleEnemyMove() {
	var activeEnemies []*Enemy

	if len(g.enemies) == 0 {
		return
	}

	if g.enemiesMoveRight && g.enemies[len(g.enemies)-1].x+enemySize >= screenWidth {
		g.enemiesMoveRight = false
		g.enemiesMoveDown = true
	}

	if !g.enemiesMoveRight && g.enemies[0].x < 0 {
		g.enemiesMoveRight = true
		g.enemiesMoveDown = true
	}

	for _, enemy := range g.enemies {

		// move enemies X
		if g.enemiesMoveRight {
			enemy.x += g.enemiesSpeed
		} else if !g.enemiesMoveRight {
			enemy.x -= g.enemiesSpeed
		}
		// move enemies Y
		if g.enemiesMoveDown {
			g.soundPlayer.PlaySound(soundEnemyStepDown)
			enemy.y += enemySize
		}
		activeEnemies = append(activeEnemies, enemy)
	}
	g.enemiesMoveDown = false
	g.enemies = activeEnemies
}

func (g *Game) gameReset() {
	g.playerX = playerStartX
	g.playerY = playerStartY
	g.bullets = []*Bullet{}
	g.gameStatus = gameOngoing
	g.lastFireTime = time.Now().Add(-fireCooldown)
}

func (g *Game) restartGame() {
	g.gameReset()
	g.gameLevel = gameLevel1
	g.InitializeEnemies(g.gameLevel)
}

func (g *Game) startNextLevel() {
	g.gameReset()
	g.gameLevel += 1
	g.InitializeEnemies(g.gameLevel)
}

func (g *Game) handleCollisions() {
	var remainingBullets []*Bullet
	var remainingEnemies []*Enemy

	for _, enemy := range g.enemies {
		hit := false
		for _, bullet := range g.bullets {
			if bullet.x < enemy.x+enemySize &&
				bullet.x+bulletWidth > enemy.x &&
				bullet.y < enemy.y+enemySize &&
				bullet.y+bulletHeight > enemy.y {
				// Collision detected; mark this enemy as hit and don't add it to remainingEnemies
				hit = true
				g.soundPlayer.PlaySound(soundHit)
				break
			}
		}
		// Only keep the enemy if it wasn’t hit
		if !hit {
			remainingEnemies = append(remainingEnemies, enemy)
		}
	}

	// Keep bullets that haven’t hit any enemies
	for _, bullet := range g.bullets {
		hit := false
		for _, enemy := range g.enemies {
			if bullet.x < enemy.x+enemySize &&
				bullet.x+bulletWidth > enemy.x &&
				bullet.y < enemy.y+enemySize &&
				bullet.y+bulletHeight > enemy.y {
				hit = true
				break
			}
		}
		if !hit {
			remainingBullets = append(remainingBullets, bullet)
		}
	}

	// Update bullets and enemies after removing hit objects
	g.bullets = remainingBullets
	g.enemies = remainingEnemies
}

func (g *Game) handleGameStatusChange() {
	// next level
	if len(g.enemies) == 0 && g.gameLevel != gameLevel5 {
		fmt.Println("Current game level: " + strconv.Itoa(g.gameLevel))
		fmt.Println("Level cleared! More to come")
		g.soundPlayer.PlaySound(soundWon)
		g.gameStatus = gameNextLevel
	}

	// player won if cleared all the stages
	if len(g.enemies) == 0 && g.gameLevel == gameLevel5 {
		fmt.Println("all levels cleared. Player won!")
		g.soundPlayer.PlaySound(soundWon)
		g.gameStatus = gameWon
	}

	// lost when enemy is in the same row as player
	for _, enemy := range g.enemies {
		if enemy.y >= playerStartY-playerSize {
			fmt.Println("YOU LOST!")
			g.soundPlayer.PlaySound(soundExplosion)
			g.gameStatus = gameLost
			break
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// bg
	bgOptions := &ebiten.DrawImageOptions{}
	bgOptions.GeoM.Translate(0, 0)
	screen.DrawImage(background, bgOptions)

	if g.gameStatus == gameStartScreen {
		ebitenutil.DebugPrint(screen, "This is the start screen. Press Enter to play the game")
		return
	}

	// Level won text
	levelString := strconv.Itoa(g.gameLevel)
	if g.gameStatus == gameNextLevel {
		ebitenutil.DebugPrint(screen, "Level "+levelString+" cleared! Press Enter to proceed to the next level")
		return
	}
	// Game won text
	if g.gameStatus == gameWon {
		// Display "You Won!" when the game is over
		ebitenutil.DebugPrint(screen, "You Won! Press Enter to Restart")
		return
	}
	// Game lost text
	if g.gameStatus == gameLost {
		ebitenutil.DebugPrint(screen, "You Lost! Press Enter to Restart")
		return
	}
	ebitenutil.DebugPrint(screen, gameTitle+": level "+levelString)

	// draw player
	vector.DrawFilledRect(screen, g.playerX, g.playerY, playerSize, playerSize, color.White, false)

	// draw bullets
	for _, bullet := range g.bullets {
		vector.DrawFilledRect(screen, bullet.x, bullet.y, bulletWidth, bulletHeight, color.White, false)
	}

	// draw enemies
	for _, enemy := range g.enemies {
		vector.DrawFilledRect(screen, enemy.x, enemy.y, enemySize, enemySize, color.White, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenW, screenH int) {
	return screenWidth, screenHeight
}

func (g *Game) InitializeEnemies(gameLevel int) {
	g.enemies = []*Enemy{}
	g.enemiesMoveRight = true
	g.enemiesMoveDown = false
	g.enemiesSpeed = float32(1 + (gameLevel / 5))
	enemyCount := 10 * gameLevel
	enemyRows := int(math.Ceil(float64(enemyCount) / 10))
	enemyColumns := enemyCount / enemyRows
	enemyStartX := (screenWidth - enemyColumns*enemySize - (enemyColumns-1)*enemySpace) / 2 // top middle of the screen
	enemyStartY := 40

	for row := 0; row < enemyRows; row++ {
		for col := 0; col < enemyColumns; col++ {
			x := float32(enemyStartX + col*(enemySize+enemySpace))
			y := float32(enemyStartY + row*(enemySize+enemySpace))
			g.enemies = append(g.enemies, &Enemy{x: x, y: y})
		}
	}
}

func NewSoundPlayer(ctx *audio.Context) *SoundPlayer {
	return &SoundPlayer{
		audioContext: ctx,
		players:      make(map[string]*audio.Player),
	}
}

// LoadSound loads a .wav sound from byte data and assigns it a name.
func (sp *SoundPlayer) LoadSound(name string, wavData []byte) error {
	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(wavData))
	if err != nil {
		fmt.Println(err)
		return err
	}
	player, err := sp.audioContext.NewPlayer(stream)
	if err != nil {
		fmt.Println(err)
		return err
	}
	sp.players[name] = player
	return nil
}

// PlaySound plays the sound by name, rewinding it to the beginning each time.
func (sp *SoundPlayer) PlaySound(name string) {
	if player, ok := sp.players[name]; ok {
		player.Rewind()
		player.Play()
	}
}

func loadSounds(soundPlayer *SoundPlayer) error {
	fmt.Println("Start loading sounds")

	sounds := []*Sound{
		{soundShoot, shootSoundByte},
		{soundHit, hitSoundByte},
		{soundExplosion, explosionSoundByte},
		{soundEnemyStepDown, enemyStepDownSoundByte},
		{soundWon, wonSoundByte},
	}

	for _, sound := range sounds {
		fmt.Println("Sound: " + sound.name + " loaded.")
		if err := soundPlayer.LoadSound(sound.name, sound.source); err != nil {
			log.Fatal(err)
			return err
		}
	}

	fmt.Println("All sounds loaded success!")
	return nil
}

func main() {
	audioContext := audio.NewContext(sampleRate)
	soundPlayer := NewSoundPlayer(audioContext)

	loadSounds(soundPlayer)

	game := &Game{
		playerX:      playerStartX,
		playerY:      playerStartY,
		lastFireTime: time.Now().Add(-fireCooldown),
		gameLevel:    gameLevel1,
		soundPlayer:  soundPlayer,
		gameStatus:   gameStartScreen,
	}

	game.InitializeEnemies(gameLevel1)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(gameTitle)

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
