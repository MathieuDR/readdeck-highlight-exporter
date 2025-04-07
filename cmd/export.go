/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"net/http"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/repository"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		exporter := getExporter()
		ctx := context.Background()

		log.Println("Starting export from Readdeck...")
		notes, err := exporter.Export(ctx)

		if err != nil {
			// Detailed error in debug mode
			log.Fatalf("Export failed:\n\n%v", err)
		}

		// Log export summary
		log.Printf("Export completed successfully! Exported %d notes", len(notes))

		// Count total highlights
		totalHighlights := 0
		for _, note := range notes {
			totalHighlights += len(note.Highlights)
		}

		// Provide a summary of what was exported
		log.Printf("Total highlights: %d", totalHighlights)

		// Show a brief summary of each exported note
		if len(notes) < 20 {
			log.Println("\nExported notes:")
			for i, note := range notes {
				log.Printf("%d. %s (%d highlights) -> %s",
					i+1,
					note.Bookmark.Title,
					len(note.Highlights),
					note.Path)
			}
		}

		log.Println("Export completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Update command information
	exportCmd.Short = "Export highlights from Readdeck to Zettelkasten notes"
	exportCmd.Long = `Export highlights from Readdeck to your Zettelkasten system as Markdown notes.

This command:
- Fetches your highlights from Readdeck
- Groups them by their parent document
- Generates or updates structured notes in your Zettelkasten system
- Preserves all metadata such as URLs, publication dates, and authors
- Groups highlights by color

Examples:
  readdeck-highlight-exporter export
  readdeck-highlight-exporter export --debug`
}

func getExporter() *service.Exporter {
	client := getClient()
	repo := getRepository()
	return service.NewExporter(client, repo)
}

func getClient() readdeck.Client {
	timeout := viper.GetDuration("readdeck.request_timeout")
	baseURL := viper.GetString("readdeck.base_url")
	token := viper.GetString("readdeck.token")

	httpClient := http.Client{
		Timeout: timeout,
	}
	return readdeck.NewHttpClient(httpClient, baseURL, token, 100)
}

func getRepository() repository.NoteRepository {
	formatter := repository.NewHighlightFormatter(repository.DefaultColorConfig())
	parser := repository.NewYAMLNoteParser()
	generator := repository.NewYAMLNoteGenerator(formatter)
	updater := repository.NewYAMLNoteUpdater(generator, parser)
	noteService := repository.NewCustomNoteService(parser, generator, updater)
	fleetingPath := viper.GetString("export.fleeting_path")
	log.Printf("Saving to: %s", fleetingPath)
	return repository.NewFileNoteRepository(fleetingPath, noteService)
}
