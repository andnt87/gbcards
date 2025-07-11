package main

import (
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 1280
	screenHeight = 780
)

var cardValues = map[string]int{
	"cardSpades1":  1,
	"cardSpades2":  2,
	"cardSpades3":  3,
	"cardSpades4":  4,
	"cardSpades5":  5,
	"cardSpades6":  6,
	"cardSpades7":  7,
	"cardSpades8":  8,
	"cardSpades9":  9,
	"cardSpades10": 10,
	"cardSpades11": 11,
	"cardSpades12": 12,
	"cardSpades13": 13,
	"cardSpades14": 14,

	"cardHearts12": 12,
	"cardHearts13": 13,
	"cardHearts14": 14,
}

type Level struct {
	Name          string
	NumTableDeck  int
	NumTableCards int
}

type Card struct {
	Index   int
	Name    string
	Value   int
	Texture rl.Texture2D
}

type Game struct {
	RT         rl.RenderTexture2D
	Font       rl.Font
	FontMedium rl.Font

	PlayerTextures map[string]rl.Texture2D
	TableTextures  map[string]rl.Texture2D
	BackTextures   map[string]rl.Texture2D

	TableDeck   []Card
	TableCards  []Card
	PlayerDeck  []Card
	PlayerCards []Card

	Scale   float32
	SrcRect rl.Rectangle
	DstRect rl.Rectangle

	Started     bool
	Level       []Level
	LevelSet    int
	CardClicked Card
	GameOver    bool
}

func NewGame() *Game {
	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagMsaa4xHint | rl.FlagVsyncHint | rl.FlagWindowResizable)
	rl.InitWindow(screenWidth, screenHeight, "GB Game")

	g := &Game{
		RT:             rl.LoadRenderTexture(1920, 1056),
		Font:           rl.LoadFontEx("res/fonts/Roboto-Regular.ttf", 40, nil, 0),
		FontMedium:     rl.LoadFontEx("res/fonts/Roboto-Medium.ttf", 40, nil, 0),
		PlayerTextures: loadTextures("res/cards/player/*.png"),
		TableTextures:  loadTextures("res/cards/table/*.png"),
		BackTextures:   loadTextures("res/cards/back/*.png"),
	}
	rl.SetTextureFilter(g.Font.Texture, rl.FilterBilinear)

	g.Level = []Level{
		{Name: "Easy", NumTableDeck: 12, NumTableCards: 3},
		{Name: "Medium", NumTableDeck: 16, NumTableCards: 4},
		{Name: "Hard", NumTableDeck: 20, NumTableCards: 5},
	}

	return g
}

func (g *Game) Shutdown() {
	for _, t := range g.PlayerTextures {
		rl.UnloadTexture(t)
	}
	for _, t := range g.TableTextures {
		rl.UnloadTexture(t)
	}
	for _, t := range g.BackTextures {
		rl.UnloadTexture(t)
	}
	rl.UnloadFont(g.Font)
	rl.UnloadFont(g.FontMedium)
	rl.UnloadRenderTexture(g.RT)
	rl.CloseWindow()
}

func (g *Game) Run() {
	rl.SetTargetFPS(60)
	for !rl.WindowShouldClose() {
		g.Update()
		g.Draw()
	}
}

func (g *Game) Update() {
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	vw := float32(g.RT.Texture.Width)
	vh := float32(g.RT.Texture.Height)

	g.Scale = float32(math.Min(float64(sw/vw), float64(sh/vh)))
	g.SrcRect = rl.Rectangle{Width: vw, Height: -vh}
	g.DstRect = rl.Rectangle{
		X:      (sw - vw*g.Scale) / 2,
		Y:      (sh - vh*g.Scale) / 2,
		Width:  vw * g.Scale,
		Height: vh * g.Scale,
	}
}

func (g *Game) Draw() {
	rl.BeginTextureMode(g.RT)
	rl.ClearBackground(rl.SkyBlue)
	g.TheGame()
	rl.EndTextureMode()

	rl.BeginDrawing()
	rl.ClearBackground(rl.SkyBlue)
	rl.DrawTexturePro(g.RT.Texture, g.SrcRect, g.DstRect, rl.Vector2{}, 0, rl.White)
	rl.EndDrawing()
}

func (g *Game) GetTransformedMousePos() rl.Vector2 {
	mousePos := rl.GetMousePosition()
	mousePos.X = (mousePos.X - g.DstRect.X) / g.Scale
	mousePos.Y = (mousePos.Y - g.DstRect.Y) / g.Scale
	return mousePos
}

func (g *Game) CreateTableDeck() {
	g.TableDeck = make([]Card, 0)
	g.TableCards = make([]Card, 0)
	for i := 0; i < g.Level[g.LevelSet].NumTableDeck; i++ {
		rn := rand.Intn(len(g.TableTextures))
		names := []string{"cardHearts12", "cardHearts13", "cardHearts14"}
		cardTable := Card{
			Name:    names[rn],
			Value:   cardValues[names[rn]],
			Texture: g.TableTextures[names[rn]],
		}
		g.TableDeck = append(g.TableDeck, cardTable)
	}
}

func (g *Game) CreatePlayerDeck() {
	g.PlayerDeck = make([]Card, 0)
	g.PlayerCards = make([]Card, 0)
	generateRandomNumbersWithSumEqualToCardValue := func(cardValue int) (nums []int) {
		numbers := make([]int, 0)
		sum := 0
		i := 0
		for sum < cardValue && i < 2 {
			num := rand.Intn(cardValue - sum)
			if num == 0 {
				num = 1 // Ensure at least one number is added
			}

			numbers = append(numbers, num)
			sum += num
			i++
		}
		if sum > cardValue {
			numbers[len(numbers)-1] -= sum - cardValue
		}

		if sum < cardValue {
			diff := cardValue - sum
			numbers[len(numbers)-1] += diff
		}

		return numbers
	}

	for i := 0; i < len(g.TableDeck); i++ {
		tableCardValue := g.TableDeck[i].Value
		numbers := generateRandomNumbersWithSumEqualToCardValue(tableCardValue)
		for _, num := range numbers {
			// get player texture based on num
			name := fmt.Sprintf("cardSpades%d", num)
			card := Card{
				Name:    name,
				Value:   cardValues[name],
				Texture: g.PlayerTextures[name],
			}
			g.PlayerDeck = append(g.PlayerDeck, card)
		}
	}
}

func (g *Game) ShuffleEveryNCards() {
	numCards := g.Level[g.LevelSet].NumTableCards * 2
	if len(g.PlayerDeck) < numCards {
		return
	}

	for i := 0; i < len(g.PlayerDeck); i += numCards {
		end := i + numCards
		if end > len(g.PlayerDeck) {
			end = len(g.PlayerDeck)
		}
		subSlice := g.PlayerDeck[i:end]
		rand.Shuffle(len(subSlice), func(i, j int) {
			subSlice[i], subSlice[j] = subSlice[j], subSlice[i]
		})
		copy(g.PlayerDeck[i:end], subSlice)
	}
}

func (g *Game) TheGame() {
	vw := float32(g.RT.Texture.Width)
	vh := float32(g.RT.Texture.Height)
	column := vw / 12
	xLeft := column
	xCenter := vw / 2
	xRight := vw - column
	row := vh / 24
	yTop := row
	yCenter := vh / 2
	yBottom := vh - yTop
	yPadding := float32(50)
	mousePos := g.GetTransformedMousePos()

	debugCircles(xLeft, yTop, yCenter, yBottom, xCenter, xRight, 0)

	// -- header
	fpsText := fmt.Sprintf("%d fps, Screen: %.0fx%.0f, Viewport: %.0fx%.0f, Scale: %.2f", rl.GetFPS(),
		float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()), vw, vh, g.Scale)
	rl.DrawTextEx(g.Font, fpsText, rl.Vector2{X: xLeft, Y: yTop}, 40, 0, rl.DarkBlue)

	if !g.Started {
		// -- buttons to start the game
		makeButton := func(text string, raw int, color rl.Color) {
			buttonWidth := float32(200)
			buttonHeight := float32(60)
			buttonX := xCenter - buttonWidth/2
			buttonY := yCenter - buttonHeight/2 - (buttonHeight * 3) + float32(raw)*(buttonHeight+20)
			textWidth := rl.MeasureTextEx(g.Font, text, 40, 0).X

			rl.DrawRectanglePro(rl.Rectangle{X: buttonX, Y: buttonY, Width: buttonWidth, Height: buttonHeight}, rl.Vector2{}, 0, color)
			rl.DrawTextEx(g.Font, text, rl.Vector2{X: buttonX + (buttonWidth-textWidth)/2, Y: buttonY + 10}, 40, 0, rl.White)
			if rl.CheckCollisionPointRec(mousePos, rl.Rectangle{X: buttonX, Y: buttonY, Width: buttonWidth, Height: buttonHeight}) {
				rl.DrawRectangleLinesEx(rl.Rectangle{X: buttonX, Y: buttonY, Width: buttonWidth, Height: buttonHeight}, 2, rl.DarkBlue)
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
					g.Started = true
					g.GameOver = false
					g.TableDeck = nil
					g.TableCards = nil
					g.PlayerDeck = nil
					g.PlayerCards = nil
					switch text {
					case "Easy":
						g.LevelSet = 0
					case "Medium":
						g.LevelSet = 1
					case "Hard":
						g.LevelSet = 2
					case "Win Test":
						g.TableCards = nil
						g.PlayerCards = nil
						return
					case "Lost Test":
						g.PlayerCards = nil
						g.PlayerDeck = nil
						g.TableCards = append(g.TableCards, Card{})
						return
					}
					g.CreateTableDeck()
					g.CreatePlayerDeck()
					g.ShuffleEveryNCards()
				}
			}
		}

		makeButton("Easy", 0, rl.Blue)
		makeButton("Medium", 1, rl.Blue)
		makeButton("Hard", 2, rl.Blue)
		makeButton("Win Test", 3, rl.Orange)
		makeButton("Lost Test", 4, rl.Orange)
	}

	if g.Started {
		cardWidth := float32(140 * 1.5)
		cardHeight := float32(200 * 1.5)
		cardPadding := float32(20)
		cardsNum := g.Level[g.LevelSet].NumTableCards
		xCard := func(i int) float32 {
			return xLeft + float32(i)*(cardWidth+cardPadding)
		}

		// -- footer
		footerText := fmt.Sprintf("Difficulty: %s, Table Deck: %d, Player Deck: %d", g.Level[g.LevelSet].Name, len(g.TableDeck), len(g.PlayerDeck))
		rl.DrawTextEx(g.Font, footerText, rl.Vector2{X: xLeft, Y: yBottom - 40}, 40, 0, rl.DarkBlue)

		// -- button reset right footer
		textWidth := rl.MeasureTextEx(g.Font, "Reset", 40, 0).X
		xButton := xRight - 200
		// Move the button to the top right
		yButton := yTop + 20
		rl.DrawRectanglePro(rl.Rectangle{X: xButton, Y: yButton, Width: 180, Height: 50}, rl.Vector2{}, 0, rl.Orange)
		rl.DrawTextEx(g.Font, "Reset", rl.Vector2{X: xButton + textWidth/2, Y: yButton + 5}, 40, 0, rl.White)
		if rl.CheckCollisionPointRec(mousePos, rl.Rectangle{X: xButton, Y: yButton, Width: 180, Height: 50}) {
			rl.DrawRectangleLinesEx(rl.Rectangle{X: xButton, Y: yButton, Width: 180, Height: 50}, 2, rl.Red)
			if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				g.Started = false
				g.GameOver = false
			}
		}

		// -- table cards
		if len(g.TableCards) == 0 && len(g.TableDeck) >= cardsNum {
			g.TableCards = g.TableDeck[len(g.TableDeck)-cardsNum:]
			g.TableDeck = g.TableDeck[:len(g.TableDeck)-cardsNum]
		}

		if len(g.TableCards) > 0 {
			for i := range g.TableCards {
				card := &g.TableCards[i]
				x := xCard(i)
				y := yTop + yPadding*2
				rl.DrawTexturePro(card.Texture, rl.Rectangle{X: 0, Y: 0, Width: float32(card.Texture.Width), Height: float32(card.Texture.Height)},
					rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}, rl.Vector2{}, 0, rl.White)
				rl.DrawTextEx(g.FontMedium, fmt.Sprintf("%d", card.Value), rl.Vector2{X: x, Y: y - yPadding}, 40, 0, rl.DarkBlue)
				if rl.CheckCollisionPointRec(mousePos, rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}) {
					rl.DrawRectangleLinesEx(rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}, 2, rl.DarkBlue)
					if rl.IsMouseButtonPressed(rl.MouseLeftButton) && g.CardClicked.Value > 0 {
						card.Value -= g.CardClicked.Value
						if card.Value <= 0 {
							g.TableCards = append(g.TableCards[:i], g.TableCards[i+1:]...)
						}
						g.PlayerCards = append(g.PlayerCards[:g.CardClicked.Index], g.PlayerCards[g.CardClicked.Index+1:]...)
						g.CardClicked = Card{}
						break
					}
				}
			}
		}

		// -- table deck
		if len(g.TableDeck) >= cardsNum {
			for i, card := range g.TableDeck {
				x := xCard(cardsNum) + cardPadding + float32(i)*(cardPadding)
				y := yTop + yPadding*2
				rl.DrawTexturePro(g.BackTextures["cardBack_red"], rl.Rectangle{X: 0, Y: 0, Width: float32(card.Texture.Width), Height: float32(card.Texture.Height)},
					rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}, rl.Vector2{}, 0, rl.White)
			}
		}

		// -- player cards
		if len(g.PlayerCards) == 0 && len(g.PlayerDeck) >= cardsNum {
			g.PlayerCards = g.PlayerDeck[len(g.PlayerDeck)-cardsNum:]
			g.PlayerDeck = g.PlayerDeck[:len(g.PlayerDeck)-cardsNum]
		}

		if len(g.PlayerCards) > 0 {
			for i, card := range g.PlayerCards {
				x := xCard(i)
				y := yBottom - cardHeight - yPadding*2
				rl.DrawTexturePro(card.Texture, rl.Rectangle{X: 0, Y: 0, Width: float32(card.Texture.Width), Height: float32(card.Texture.Height)},
					rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}, rl.Vector2{}, 0, rl.White)

				if rl.CheckCollisionPointRec(mousePos, rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}) {
					rl.DrawRectangleLinesEx(rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}, 2, rl.DarkBlue)
					if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
						g.CardClicked = card
						g.CardClicked.Index = i
					}
				}
			}
		}

		// -- player deck
		if len(g.PlayerDeck) >= cardsNum {
			for i, card := range g.PlayerDeck {
				x := xCard(cardsNum) + cardPadding + float32(i)*(4)
				y := yBottom - cardHeight - yPadding*2

				// show cards except last one
				numCards := g.Level[g.LevelSet].NumTableCards
				if i < len(g.PlayerDeck)-numCards {
					rl.DrawTexturePro(g.BackTextures["cardBack_blue"], rl.Rectangle{X: 0, Y: 0, Width: float32(card.Texture.Width), Height: float32(card.Texture.Height)},
						rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}, rl.Vector2{}, 0, rl.White)
				} else {
					// show card rotated incrementally towards right from bottom left corner
					src := rl.Rectangle{X: 0, Y: 0, Width: float32(card.Texture.Width), Height: float32(card.Texture.Height)}
					dst := rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}
					origin := rl.Vector2{X: 0, Y: cardHeight}
					dst.Y += dst.Height
					j := float32(len(g.PlayerDeck) - i)
					k := float32(math.Min(float64(len(g.PlayerDeck)), float64(numCards)))
					rot := (k - j) * 15
					rl.DrawTexturePro(card.Texture, src, dst, origin, rot, rl.White)
					if rl.CheckCollisionPointRec(mousePos, rl.Rectangle{X: x, Y: y, Width: cardWidth, Height: cardHeight}) {
					}
				}
			}
		}

		// -- win/lose condition
		won := len(g.TableCards) == 0 && len(g.PlayerCards) == 0
		lost := len(g.PlayerDeck) == 0 && len(g.PlayerCards) == 0 && len(g.TableCards) > 0
		if won || lost || g.GameOver {
			// Set game over flag if not already set
			g.GameOver = true
			var resultText string
			var textColor rl.Color
			if won {
				resultText = "You won!"
				textColor = rl.DarkBlue
			} else {
				resultText = "You lost!"
				textColor = rl.Red
			}
			textWidth := rl.MeasureTextEx(g.FontMedium, resultText, 80, 0).X
			rl.DrawTextEx(g.Font, resultText, rl.Vector2{X: xCenter - textWidth/2, Y: yCenter - 20}, 80, 0, textColor)
			return // Skip remaining drawing so the win/lose screen stays visible
		}
	}
}

func loadTextures(imagesPath string) map[string]rl.Texture2D {
	files, _ := filepath.Glob(imagesPath)
	textures := make(map[string]rl.Texture2D)
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		img := rl.LoadImage(file)
		textures[name] = rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
	}
	return textures
}

func debugCircles(xLeft float32, yTop float32, yCenter float32, yBottom float32, xCenter float32, xRight float32, radius float32) {
	if radius <= 0 {
		return
	}
	rl.DrawCircle(int32(xLeft), int32(yTop), radius, rl.Red)
	rl.DrawCircle(int32(xLeft), int32(yCenter), radius, rl.Red)
	rl.DrawCircle(int32(xLeft), int32(yBottom), radius, rl.Red)
	rl.DrawCircle(int32(xCenter), int32(yTop), radius, rl.Red)
	rl.DrawCircle(int32(xCenter), int32(yCenter), radius, rl.Red)
	rl.DrawCircle(int32(xCenter), int32(yBottom), radius, rl.Red)
	rl.DrawCircle(int32(xRight), int32(yTop), radius, rl.Red)
	rl.DrawCircle(int32(xRight), int32(yCenter), radius, rl.Red)
	rl.DrawCircle(int32(xRight), int32(yBottom), radius, rl.Red)
}

func main() {
	game := NewGame()
	defer game.Shutdown()
	game.Run()
}
