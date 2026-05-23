package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/house/goscribe/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	RunE:  runVersion,
}

var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Bump patch version and commit",
	Long:  `Bump the patch version (0.0.x), commit with a Scooby-Doo quote, and optionally tag.`,
	RunE:  runBump,
}

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Tag current commit with version",
	RunE:  runTag,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(bumpCmd)
	rootCmd.AddCommand(tagCmd)
}

func runVersion(cmd *cobra.Command, args []string) error {
	v := version.Get()
	fmt.Printf("goscribe version %s\n", v)
	return nil
}

func runBump(cmd *cobra.Command, args []string) error {
	newVersion, err := version.BumpPatch()
	if err != nil {
		return fmt.Errorf("bump version: %w", err)
	}
	
	quote := version.GetScoobyQuote()
	message := fmt.Sprintf("Bump version to %s\n\n%s", newVersion, quote)
	
	commitCmd := exec.Command("git", "add", "-A")
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("stage changes: %w", err)
	}
	
	commitCmd = exec.Command("git", "commit", "-m", message)
	output, err := commitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("commit failed: %w\n%s", err, string(output))
	}
	
	fmt.Printf("Version bumped to %s\n", newVersion)
	fmt.Printf("Commit message: %s\n", strings.Split(message, "\n")[0])
	fmt.Printf("Scooby quote: %s\n", quote)
	
	return nil
}

func runTag(cmd *cobra.Command, args []string) error {
	v := version.Get()
	
	tagCmd := exec.Command("git", "tag", "-a", v, "-m", fmt.Sprintf("Release %s", v))
	output, err := tagCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tag failed: %w\n%s", err, string(output))
	}
	
	fmt.Printf("Tagged %s\n", v)
	return nil
}
