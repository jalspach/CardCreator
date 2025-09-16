package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	//	"strings"
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

// GenerateCard creates a PNG image of a business card that we can use as an email footer.
// It takes paths to a regular font and a bold font.
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

	// Format the phone number before drawing it.
	formattedPhone := cardData.PhoneNumber
	// Remove all non-digit characters from the phone number
	digitsOnly := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, cardData.PhoneNumber)

	if len(digitsOnly) == 10 {
		formattedPhone = fmt.Sprintf("(%s) %s-%s", digitsOnly[0:3], digitsOnly[3:6], digitsOnly[6:10])
	}

	drawText(img, otherFace, fmt.Sprintf("%s", formattedPhone), 70, y, otherTextColor)
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

// cardHandler handles the HTTP request to generate a business card.
func cardHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data from the request body.
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Get the business card details from the form.
	cardData := BusinessCard{
		Name:        r.FormValue("name"),
		Pronouns:    r.FormValue("pronouns"),
		Title:       r.FormValue("title"),
		Company:     r.FormValue("company"),
		Address:     r.FormValue("address"),
		PhoneNumber: r.FormValue("phone_number"),
		Email:       r.FormValue("email"),
	}

	// Set paths to the assets. **You must update these paths.**
	bgImagePath := "../BrandingGuidelines_2025.png"
	regularFontPath := "../Raleway-Regular.ttf"
	boldFontPath := "../Raleway-Bold.ttf"

	// Generate the business card image.
	img, err := GenerateCard(bgImagePath, regularFontPath, boldFontPath, cardData)
	if err != nil {
		log.Printf("Error generating image: %v", err)
		http.Error(w, "Failed to generate business card image", http.StatusInternalServerError)
		return
	}

	// Set the content type header to PNG.
	w.Header().Set("Content-Type", "image/png")

	// Set the Content-Disposition header to suggest a filename.
	w.Header().Set("Content-Disposition", "attachment; filename=\"EmailSignature.png\"")

	// Encode and write the image to the response writer.
	if err := png.Encode(w, img); err != nil {
		log.Printf("Error encoding image: %v", err)
		http.Error(w, "Failed to encode PNG image", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Serve static files from the 'assets' directory.
	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Handle the form submission for generating the card.
	http.HandleFunc("/generate-card", cardHandler)

	// Serve the main HTML file.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./emailfooterserverui.html")
	})

	fmt.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
