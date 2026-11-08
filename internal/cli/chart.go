package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var chartCmd = &cobra.Command{
	Use:   "chart",
	Short: "Package and push Helm charts to OCI registries",
}

var chartPackageCmd = &cobra.Command{
	Use:   "package [path]",
	Short: "Package a Helm chart directory into a versioned chart archive",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chartPath := "charts"
		if len(args) > 0 {
			chartPath = args[0]
		}

		cmd.Printf("📦 Packaging Helm chart from '%s'...\n", chartPath)

		// If helm binary is available in $PATH, use it
		if _, err := exec.LookPath("helm"); err == nil {
			execCmd := exec.Command("helm", "package", chartPath)
			out, err := execCmd.CombinedOutput()
			if err == nil {
				cmd.Print(string(out))
				cmd.Println("✓ Chart packaged successfully via Helm CLI.")
				return nil
			}
		}

		// Fallback to pure Go native packaging
		cmd.Println("ℹ Helm CLI not found. Packaging chart using native MoreOps packager...")
		outPkg, err := packageChartNative(chartPath)
		if err != nil {
			return fmt.Errorf("chart packaging failed: %w", err)
		}

		cmd.Printf("Successfully packaged chart and saved it to: %s\n", outPkg)
		cmd.Println("✓ Chart packaged successfully.")
		return nil
	},
}

func packageChartNative(chartDir string) (string, error) {
	chartYamlPath := filepath.Join(chartDir, "Chart.yaml")
	data, err := os.ReadFile(chartYamlPath)
	if err != nil {
		return "", fmt.Errorf("could not find %s: %w", chartYamlPath, err)
	}

	type chartMeta struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}
	var meta chartMeta
	if err := yaml.Unmarshal(data, &meta); err != nil || meta.Name == "" {
		meta.Name = filepath.Base(chartDir)
		meta.Version = "0.1.0"
	}
	if meta.Version == "" {
		meta.Version = "0.1.0"
	}

	outTgz := fmt.Sprintf("%s-%s.tgz", meta.Name, meta.Version)
	outFile, err := os.Create(outTgz)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	baseDirName := meta.Name
	err = filepath.Walk(chartDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(chartDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(filepath.Join(baseDirName, relPath))

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return outTgz, nil
}

var chartPushCmd = &cobra.Command{
	Use:   "push <archive.tgz> <oci://registry/path>",
	Short: "Push a packaged Helm chart archive to an OCI registry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		archive := args[0]
		registry := args[1]

		cmd.Printf("🚀 Pushing chart '%s' to OCI registry '%s'...\n", archive, registry)
		execCmd := exec.Command("helm", "push", archive, registry)
		out, err := execCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("helm push failed: %w (output: %s)", err, string(out))
		}

		cmd.Print(string(out))
		cmd.Println("✓ Chart pushed successfully to OCI registry.")
		return nil
	},
}

func init() {
	chartCmd.AddCommand(chartPackageCmd)
	chartCmd.AddCommand(chartPushCmd)
	rootCmd.AddCommand(chartCmd)
}
