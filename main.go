package main

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Bullet struct {
	x, y float32
}

type Enemy struct {
	x, y float32
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
}

const (
	gameTitle = "My Space Invaders"

	screenWidth  = 640
	screenHeight = 480

	gameOngoing   = 0
	gameWon       = 1
	gameLost      = 2
	gameNextLevel = 4

	gameLevel1 = 1
	gameLevel2 = 1
	gameLevel3 = 1
	gameLevel4 = 1
	gameLevel5 = 1

	playerSize      = 12
	playerStartX    = screenWidth / 2
	playerStartY    = float32(screenHeight - playerSize - 10)
	playerMoveSpeed = 2

	bulletWidth  = 2
	bulletHeight = 4
	bulletSpeed  = 5
	fireCooldown = 1 * time.Second

	enemySize  = 20
	enemySpace = 20
)

func (g *Game) Update() error {
	// handle game status
	if g.gameStatus != gameOngoing &&
		g.gameStatus != gameNextLevel &&
		ebiten.IsKeyPressed(ebiten.KeyEnter) {
		g.restartGame()
		return nil
	}
	if g.gameStatus == gameNextLevel {
		g.startNextLevel()
		return nil
	}
	if g.gameStatus != gameOngoing {
		return nil
	}
	// other handles which doesn't stop to game
	g.handleKeyPress()
	g.handleBulletMove()
	g.handleEnemyMove()
	g.handleCollisions()
	g.handleGameEnd()
	return nil
}

func (g *Game) handleKeyPress() {
	// Move player
	// move left
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.playerX -= playerMoveSpeed
	}
	// move right
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.playerX += playerMoveSpeed
	}
	// fire bullets
	if ebiten.IsKeyPressed(ebiten.KeySpace) && time.Since(g.lastFireTime) >= fireCooldown {
		// fire a new bullet and update the last fire time
		g.bullets = append(g.bullets, &Bullet{x: g.playerX + playerSize/2 - bulletWidth/2, y: g.playerY})
		g.lastFireTime = time.Now()
	}
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
			enemy.y += enemySize
		}
		activeEnemies = append(activeEnemies, enemy)
	}
	g.enemiesMoveDown = false
	g.enemies = activeEnemies
}

func (g *Game) restartGame() {
	g.playerX = playerStartX
	g.playerY = playerStartY
	g.bullets = []*Bullet{}         // Clear bullets
	g.InitializeEnemies(gameLevel1) // Reset enemies
	g.gameStatus = gameOngoing      // Set game state to ongoing
	g.lastFireTime = time.Now().Add(-fireCooldown)
}

func (g *Game) startNextLevel() {
	g.playerX = playerStartX
	g.playerY = playerStartY
	g.bullets = []*Bullet{} // Clear bullets
	g.gameLevel += 1
	g.InitializeEnemies(g.gameLevel) // Reset enemies
	g.gameStatus = gameOngoing       // Set game state to ongoing
	g.lastFireTime = time.Now().Add(-fireCooldown)
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

func (g *Game) handleGameEnd() {
	// won
	if len(g.enemies) == 0 && g.gameLevel == gameLevel5 {
		g.gameStatus = gameWon
	}
	// next level
	if len(g.enemies) == 0 {
		g.gameStatus = gameNextLevel
	}
	// lost
	for _, enemy := range g.enemies {
		if enemy.y >= playerStartY-playerSize {
			g.gameStatus = gameLost
			break
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.gameStatus == gameNextLevel {
		ebitenutil.DebugPrint(screen, "Level Won! Press Enter to proceed next level")
		return
	}
	if g.gameStatus == gameWon {
		// Display "You Won!" when the game is over
		ebitenutil.DebugPrint(screen, "You Won! Press Enter to Restart")
		return
	}
	if g.gameStatus == gameLost {
		// Display "You Won!" when the game is over
		ebitenutil.DebugPrint(screen, "You Lost! Press Enter to Restart")
		return
	}
	ebitenutil.DebugPrint(screen, gameTitle)
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
	g.enemiesSpeed = float32(1 + (gameLevel / 3))
	enemyCount := 10 * gameLevel
	enemyRows := enemyCount / 10
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

func main() {
	game := &Game{
		playerX:      playerStartX,
		playerY:      playerStartY,
		lastFireTime: time.Now().Add(-fireCooldown),
		gameLevel:    gameLevel1,
	}
	game.InitializeEnemies(gameLevel1)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(gameTitle)
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
