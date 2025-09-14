package main

import (
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// BusinessCard holds the data for the card.
type BusinessCard struct {
	Name        string
	Pronouns    string
	Title       string
	Company     string
	Address     string
	PhoneNumber string
	Email       string
}

// GenerateCard creates a PNG image of a business card.
// It now takes paths to a regular font and a bold font.
func GenerateCard(bgImagePath, regularFontPath, boldFontPath string, cardData BusinessCard) (image.Image, error) {
	// Open and decode the background image file.
	bgFile, err := os.Open(bgImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open background image: %w", err)
	}
	defer bgFile.Close()
	bgImage, _, err := image.Decode(bgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode background image: %w", err)
	}

	// Create a new RGBA image with the same bounds as the background.
	bounds := bgImage.Bounds()
	img := image.NewRGBA(bounds)

	// Draw the background image onto the new image.
	draw.Draw(img, bounds, bgImage, image.Point{}, draw.Src)

	// Load the regular font.
	regularFontBytes, err := ioutil.ReadFile(regularFontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read regular font file: %w", err)
	}
	regularFont, err := opentype.Parse(regularFontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse regular font: %w", err)
	}

	// Load the bold font.
	boldFontBytes, err := ioutil.ReadFile(boldFontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bold font file: %w", err)
	}
	boldFont, err := opentype.Parse(boldFontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bold font: %w", err)
	}

	// Set up text drawing properties.
	nameFontSize := 20.0
	pronounsFontSize := 13.0
	titleFontSize := 15.0 // at this size there is a 36 char limit
	companyFontSize := 15.0
	otherFontSize := 13.0
	midFontSize := 14.0

	// Create font faces for the three different styles.
	nameFace, err := opentype.NewFace(boldFont, &opentype.FaceOptions{
		Size:    nameFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create name font face: %w", err)
	}

	midFace, err := opentype.NewFace(boldFont, &opentype.FaceOptions{
		Size:    midFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create mid font face: %w", err)
	}

	titleFace, err := opentype.NewFace(boldFont, &opentype.FaceOptions{
		Size:    titleFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create mid font face: %w", err)
	}

	otherFace, err := opentype.NewFace(regularFont, &opentype.FaceOptions{
		Size:    otherFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create other font face: %w", err)
	}

	pronounsFace, err := opentype.NewFace(regularFont, &opentype.FaceOptions{
		Size:    pronounsFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pronouns font face: %w", err)
	}

	// Define colors for the text.
	nameColor := color.RGBA{R: 242, G: 103, B: 34, A: 255}         // main Orange
	titleTextColor := color.RGBA{R: 0, G: 48, B: 71, A: 255}       // Main Blue
	companyTextColor := color.RGBA{R: 109, G: 110, B: 113, A: 255} // Main Gray
	otherTextColor := color.RGBA{R: 109, G: 110, B: 113, A: 255}   // Main Gray

	// Draw the text dynamically.
	// We'll use a y-coordinate variable that gets updated after each line.
	y := 25 // Starting y position
	drawText(img, nameFace, cardData.Name, 70, y, nameColor)
	y += int(nameFontSize) + 0

	if cardData.Pronouns != "" {
		drawText(img, pronounsFace, cardData.Pronouns, 70, y, otherTextColor)
		y += int(pronounsFontSize) + 1
	}

	drawText(img, titleFace, fmt.Sprintf("%s", cardData.Title), 70, y, titleTextColor)
	y += int(titleFontSize) + 0

	y += 10 // Add a little more space after the name/pronouns/title block

	drawText(img, midFace, fmt.Sprintf("%s", cardData.Company), 70, y, companyTextColor)
	y += int(companyFontSize) + 0
	drawText(img, otherFace, fmt.Sprintf("%s", cardData.Address), 70, y, otherTextColor)
	y += int(otherFontSize) + 0
	drawText(img, otherFace, fmt.Sprintf("%s", cardData.PhoneNumber), 70, y, otherTextColor)
	y += int(otherFontSize) + 0
	drawText(img, otherFace, fmt.Sprintf("%s", cardData.Email), 70, y, otherTextColor)

	return img, nil
}

// Helper function to draw text on an image.
func drawText(img *image.RGBA, face font.Face, text string, x, y int, c color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}

// readCardsFromCSV reads business card data from a CSV file.
func readCardsFromCSV(csvPath string) ([]BusinessCard, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read and discard the header row.
	if _, err := reader.Read(); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("CSV file is empty")
		}
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var cards []BusinessCard
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		if len(record) < 7 {
			log.Printf("Skipping invalid record: %v (expected 7 fields, got %d)", record, len(record))
			continue
		}

		card := BusinessCard{
			Name:        record[0],
			Pronouns:    record[1],
			Title:       record[2],
			Company:     record[3],
			Address:     record[4],
			PhoneNumber: record[5],
			Email:       record[6],
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func main() {
	// --- Replace these values with your own details ---
	bgImagePath := "./BrandingGuidelines_2025.png"
	regularFontPath := "./Raleway-Regular.ttf"
	boldFontPath := "./Raleway-Bold.ttf"
	csvPath := "./cards.csv"
	// --------------------------------------------------

	// Check if the background image file exists.
	if _, err := os.Stat(bgImagePath); os.IsNotExist(err) {
		log.Fatalf("Background image file does not exist: %s", bgImagePath)
	}

	// Check if the regular font file exists.
	if _, err := os.Stat(regularFontPath); os.IsNotExist(err) {
		log.Fatalf("Regular font file does not exist: %s", regularFontPath)
	}

	// Check if the bold font file exists.
	if _, err := os.Stat(boldFontPath); os.IsNotExist(err) {
		log.Fatalf("Bold font file does not exist: %s", boldFontPath)
	}

	// Check if the CSV file exists.
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		log.Fatalf("CSV file does not exist: %s", csvPath)
	}

	// Read card data from the CSV file.
	cards, err := readCardsFromCSV(csvPath)
	if err != nil {
		log.Fatalf("Error reading cards from CSV: %v", err)
	}

	if len(cards) == 0 {
		log.Println("No cards found in the CSV file.")
		return
	}

	for _, card := range cards {
		// Generate the email signature image.
		cardImage, err := GenerateCard(bgImagePath, regularFontPath, boldFontPath, card)
		if err != nil {
			log.Fatalf("Error generating email signature for %s: %v", card.Name, err)
		}

		// Sanitize the name for the filename.
		sanitizedName := strings.ReplaceAll(card.Name, " ", "_")
		outputFileName := fmt.Sprintf("%s_email_signature.png", sanitizedName)

		// Create the output file.
		outFile, err := os.Create(outputFileName)
		if err != nil {
			log.Fatalf("Failed to create output file %s: %v", outputFileName, err)
		}
		defer outFile.Close()

		// Encode and save the image as a PNG.
		if err := png.Encode(outFile, cardImage); err != nil {
			log.Fatalf("Failed to encode image to PNG for %s: %v", card.Name, err)
		}

		fmt.Printf("Successfully generated business card: %s\n", outputFileName)
	}
}
