package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/display"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/repository"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	verbose bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export highlights from Readdeck to Zettelkasten notes",
	Long: `Export highlights from Readdeck to your Zettelkasten system as Markdown notes.

This command:
- Fetches your highlights from Readdeck
- Groups them by their parent document
- Generates or updates structured notes in your Zettelkasten system
- Preserves all metadata such as URLs, publication dates, and authors
- Groups highlights by color

Examples:
  readdeck-highlight-exporter export
  readdeck-highlight-exporter export --verbose
  readdeck-highlight-exporter export --timing`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clear standard log prefix for cleaner output
		log.SetFlags(0)

		startTime := time.Now()

		exporter := getExporter()
		ctx := context.Background()

		fmt.Println("Starting export from Readdeck...")
		results, err := exporter.Export(ctx)

		if err != nil {
			log.Fatalf("Export failed:\n\n%v", err)
		}

		display.PrintSummary(results, true, time.Since(startTime))

		if verbose {
			display.PrintDetails(results)
		}

		fmt.Println("\n✅ Export completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
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
	baseURL := viper.GetString("readdeck.base_url")
	fleetingPath := viper.GetString("export.fleeting_path")

	formatter := repository.NewHighlightFormatter(repository.DefaultColorConfig())
	parser := repository.NewYAMLNoteParser()
	generator := repository.NewYAMLNoteGenerator(formatter, baseURL)
	updater := repository.NewYAMLNoteUpdater(generator, parser)
	noteService := repository.NewCustomNoteService(parser, generator, updater)
	fmt.Printf("Saving to: %s\n", fleetingPath)
	return repository.NewFileNoteRepository(fleetingPath, noteService, verbose)
}
