# Ebit Invaders

Ebit Invaders is a simple 2D space-invader-style game built with the [Ebiten](https://ebiten.org/) game engine. The player controls a spaceship and shoots enemies, progressing through levels of increasing difficulty. The game includes multiple sound effects and tracks the player's progress across multiple levels.

## Game Features

- **Player Movement**: Use arrow keys to move left and right.
- **Shooting Mechanics**: Press Space to fire bullets
- **Enemies**: Multiple levels of enemy waves, with enemies moving left, right, and down as they approach the player.
- **Sound Effects**: Realistic shooting, explosion, and victory sounds, providing an immersive experience.
- **Win and Lose Conditions**: Advance through levels by clearing enemies, with game-over on collision with player.

## Controls

- **Left Arrow**: Move player left
- **Right Arrow**: Move player right
- **Space**: Fire a bullet
- **Enter**: Start the game, restart after loss, or proceed to the next level

## Game States

- **Start Screen**: Press Enter to begin.
- **Game Won**: Clear all levels to win; press Enter to restart.
- **Game Lost**: Lose if enemies reach the player's row; press Enter to retry.

## Setup & Installation

1. **Install Ebiten**: This game uses the Ebiten library for Go. To install Ebiten, run:
   ```bash
   go get github.com/hajimehoshi/ebiten/v2
   ```

2. **Run the Game**: To start the game, navigate to the project folder and run:
   ```bash
   go run main.go
   ```

## Project Structure

- **Main Game Loop**: `Update()`, `Draw()`, `Layout()`
- **Player Controls**: `handleKeyPress()` manages input handling for player movement, shooting, and game state transitions.
- **Enemy Movement**: `handleEnemyMove()` updates enemy positions and manages direction changes.
- **Collision Detection**: `handleCollisions()` checks for bullet and enemy collisions.
- **Audio Management**: `SoundPlayer` handles loading and playing of sound effects for various game events.

## Dependencies

- **[Ebiten](https://ebiten.org/)** - Game development library for Go

## Gameplay Preview

Add a screenshot or GIF of gameplay here.

---

## License

This project is licensed under the MIT License.

---

## Acknowledgments

This project was built with Ebiten and is inspired by classic arcade games like Space Invaders.

--- 

