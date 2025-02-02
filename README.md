# Cloud Generation App

A perpetual motion cloud generation application written in Go using the [Ebiten](https://github.com/hajimehoshi/ebiten) game engine.

## Features

- Dynamic cloud density control
- "Realistic" cloud rendering with varying sizes and opacity levels
- Adjustable tree density and shadow intensity

## Controls

- **M**: Toggle Environment Controls
- **LMB**: Drag Sun or Trees
- **ESC**: Exit the application

When environment controls are active:
- **Up Arrow**: Increase tree density
- **Down Arrow**: Decrease tree density
- **Left Arrow**: Decrease cloud count
- **Right Arrow**: Increase cloud count
- **S**: Decrease tree shadow intensity
- **D**: Increase tree shadow intensity

When environment controls are hidden:
- **Up Arrow**: Increase cloud density
- **Down Arrow**: Decrease cloud density

## Requirements

- Go 1.16 or higher
- Ebiten v2 game engine (automatically installed via go modules)

## Installation

```bash
# Clone the repo,
# Install dependencies with
go mod tidy
```

## Running the Application

```bash
go run main.go
```

The application will open in a new window with an initial cloud density of 20%. Use the arrow keys to adjust the density to your preference.

## Demo
![Cloud Preview](./preview1.gif)
