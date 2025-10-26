package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	jargon_stemmer "search-index/jargon-stemmer"
	"sort"
	"strings"
	"time"
)

func generateSVGIconsData(ctx context.Context) ([]SVGIconData, error) {
	fmt.Println("🎨 Generating SVG icons data...")

	// Path to cluster.json file
	clusterPath := "../frontend/data/cluster_svg.json"

	content, err := ioutil.ReadFile(clusterPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cluster.json: %w", err)
	}

	var cluster SVGCluster
	if err := json.Unmarshal(content, &cluster); err != nil {
		return nil, fmt.Errorf("failed to parse cluster.json: %w", err)
	}

	var svgIconsData []SVGIconData
	categoryCount := 0
	iconCount := 0

	fmt.Println("Processing categories:")

	for _, clusterEntry := range cluster.Clusters {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		categoryCount++

		// Process each icon in the cluster
		for _, fileName := range clusterEntry.FileNames {
			iconCount++

			// Remove leading underscore if present and get the name without extension
			iconName := strings.TrimPrefix(fileName.FileName, "_")
			iconName = strings.TrimSuffix(iconName, ".svg")

			// Format the display name to be more user-friendly
			displayName := formatIconName(iconName)

			// Create the path (similar to Python logic)
			iconPath := fmt.Sprintf("/freedevtools/svg_icons/%s/%s/", clusterEntry.SourceFolder, iconName)

			// Generate ID from path (similar to Python logic)
			iconID := generateIconIDFromPath(iconPath)

			// Use description from fileName if available, otherwise create default
			description := fileName.Description
			if description == "" {
				description = fmt.Sprintf("SVG icon for %s", displayName)
			}

			// Generate icon data
			iconData := SVGIconData{
				ID:          iconID,
				Name:        displayName,
				Description: description,
				Path:        iconPath,
				Image:   fmt.Sprintf("/svg_icons/%s/%s", clusterEntry.SourceFolder, fileName.FileName),
				Category:    "svg_icons",
			}

			svgIconsData = append(svgIconsData, iconData)
		}
	}

	// Sort by ID
	sort.Slice(svgIconsData, func(i, j int) bool {
		return svgIconsData[i].ID < svgIconsData[j].ID
	})

	fmt.Printf("🎨 Processed %d categories with %d icons total\n", categoryCount, iconCount)
	return svgIconsData, nil
}

func generateIconIDFromPath(path string) string {
	// Remove the base path (similar to Python logic)
	cleanPath := strings.Replace(path, "/freedevtools/svg_icons/", "", 1)
	
	// Remove trailing slash if present
	cleanPath = strings.TrimSuffix(cleanPath, "/")
	
	// Replace remaining slashes with hyphens
	cleanPath = strings.Replace(cleanPath, "/", "-", -1)
	
	// Replace any invalid characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	cleanPath = reg.ReplaceAllString(cleanPath, "_")
	
	// Add prefix with hyphen and sanitize
	return fmt.Sprintf("svg-icons-%s", sanitizeID(cleanPath))
}

func formatIconName(iconName string) string {
	// Replace underscores and hyphens with spaces
	name := strings.Replace(iconName, "_", " ", -1)
	name = strings.Replace(name, "-", " ", -1)

	// Title case
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}


func RunSVGIconsOnly(ctx context.Context, start time.Time) {
	fmt.Println("🎨 Generating SVG icons data only...")

	icons, err := generateSVGIconsData(ctx)
	if err != nil {
		log.Fatalf("❌ SVG icons data generation failed: %v", err)
	}

	// Save to JSON
	if err := saveToJSON("svg_icons.json", icons); err != nil {
		log.Fatalf("Failed to save SVG icons data: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n🎉 SVG icons data generation completed in %v\n", elapsed)
	fmt.Printf("📊 Generated %d SVG icons\n", len(icons))

	// Show sample data
	fmt.Println("\n📝 Sample SVG icons:")
	for i, icon := range icons {
		if i >= 10 { // Show first 10
			fmt.Printf("  ... and %d more icons\n", len(icons)-10)
			break
		}
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, icon.Name, icon.ID)
		if icon.Description != "" {
			fmt.Printf("     Description: %s\n", truncateString(icon.Description, 80))
		}
		fmt.Printf("     Image: %s\n", icon.Image)
		fmt.Printf("     Path: %s\n", icon.Path)
		fmt.Println()
	}

	fmt.Printf("💾 Data saved to output/svg_icons.json\n")
	
	// Automatically run stem processing
	fmt.Println("\n🔍 Running stem processing...")
	if err := jargon_stemmer.ProcessJSONFile("output/svg_icons.json"); err != nil {
		log.Fatalf("❌ Stem processing failed: %v", err)
	}
	fmt.Println("✅ Stem processing completed!")
}