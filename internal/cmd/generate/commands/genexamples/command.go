// Licensed to Elasticsearch B.V. under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.

package genexamples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/tools/imports"

	"github.com/elastic/go-elasticsearch/v8/internal/cmd/generate/commands"
	"github.com/elastic/go-elasticsearch/v8/internal/cmd/generate/utils"
)

var (
	inputSrc  *string
	outputSrc *string
	debugSrc  *bool
	colorSrc  *bool
	gofmtSrc  *bool

	inputDoc  *string
	outputDoc *string
	debugDoc  *bool
)

func init() {
	inputSrc = genexamplesSrcCmd.Flags().StringP("input", "i", "", "Path to a file with specification for examples")
	outputSrc = genexamplesSrcCmd.Flags().StringP("output", "o", "", "Path to a folder for generated output")
	debugSrc = genexamplesSrcCmd.Flags().BoolP("debug", "d", false, "Print the generated source to terminal")
	colorSrc = genexamplesSrcCmd.Flags().BoolP("color", "c", true, "Syntax highlight the debug output")
	gofmtSrc = genexamplesSrcCmd.Flags().BoolP("gofmt", "f", true, "Format generated output with 'gofmt'")

	genexamplesSrcCmd.MarkFlagRequired("input")
	genexamplesSrcCmd.MarkFlagRequired("output")
	genexamplesSrcCmd.Flags().SortFlags = false

	inputDoc = genexamplesDocCmd.Flags().StringP("input", "i", "", "Path to a file with specification for examples")
	outputDoc = genexamplesDocCmd.Flags().StringP("output", "o", "", "Path to a folder for generated output")
	debugDoc = genexamplesDocCmd.Flags().BoolP("debug", "d", false, "Print the generated source to terminal")

	genexamplesDocCmd.MarkFlagRequired("input")
	genexamplesDocCmd.MarkFlagRequired("output")
	genexamplesDocCmd.Flags().SortFlags = false

	genexamplesCmd.AddCommand(genexamplesSrcCmd)
	genexamplesCmd.AddCommand(genexamplesDocCmd)

	commands.RegisterCmd(genexamplesCmd)
}

var genexamplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Generate the Go examples for documentation",
}

var genexamplesSrcCmd = &cobra.Command{
	Use:   "src",
	Short: "Generate the Go sources for examples",
	Run: func(cmd *cobra.Command, args []string) {
		command := &SrcCommand{
			Input:          *inputSrc,
			Output:         *outputSrc,
			DebugSource:    *debugSrc,
			ColorizeSource: *colorSrc,
			GofmtSource:    *gofmtSrc,
		}
		if err := command.Execute(); err != nil {
			utils.PrintErr(err)
			os.Exit(1)
		}
	},
}

var genexamplesDocCmd = &cobra.Command{
	Use:   "doc",
	Short: "Generate the ASCIIDoc examples",
	Run: func(cmd *cobra.Command, args []string) {
		command := &DocCommand{Input: *inputDoc, Output: *outputDoc, DebugSource: *debugDoc}
		if err := command.Execute(); err != nil {
			utils.PrintErr(err)
			os.Exit(1)
		}
	},
}

// SrcCommand represents the command for generating Go source code.
//
type SrcCommand struct {
	Input          string
	Output         string
	DebugSource    bool
	ColorizeSource bool
	GofmtSource    bool
}

// DocCommand represents the command for generating ASCIIDoc examples.
//
type DocCommand struct {
	Input       string
	Output      string
	DebugSource bool
}

// Execute runs the command.
//
func (cmd *SrcCommand) Execute() error {
	var (
		processed int
		skipped   int
		start     = time.Now()
	)

	if cmd.Output != "-" {
		outputDir := filepath.Join(cmd.Output, "src")
		if err := os.MkdirAll(outputDir, 0775); err != nil {
			return fmt.Errorf("error creating output directory %q: %s", outputDir, err)
		}
	}

	f, err := os.Open(cmd.Input)
	if err != nil {
		return fmt.Errorf("error reading input: %s", err)
	}
	defer f.Close()

	var examples []Example
	if err := json.NewDecoder(f).Decode(&examples); err != nil {
		return fmt.Errorf("error decoding input: %s", err)
	}

	for _, e := range examples {
		if e.Enabled() && e.Executable() {
			if utils.IsTTY() {
				fmt.Fprint(os.Stderr, "\x1b[2m")
			}
			fmt.Fprintln(os.Stderr, strings.Repeat("━", utils.TerminalWidth()))
			fmt.Fprintf(os.Stderr, "Processing example %q @ %s\n", e.ID(), e.Digest)
			if utils.IsTTY() {
				fmt.Fprint(os.Stderr, "\x1b[0m")
			}
			if err := cmd.processExample(e); err != nil {
				return fmt.Errorf("error processing example %s: %v", e.ID(), err)
			}
			processed++
		} else {
			skipped++
		}
	}

	if utils.IsTTY() {
		fmt.Fprint(os.Stderr, "\x1b[2m")
	}
	fmt.Fprintln(os.Stderr, strings.Repeat("━", utils.TerminalWidth()))
	fmt.Fprintf(os.Stderr, "Processed %d examples, skipped %d examples in %s\n", processed, skipped, time.Since(start).Truncate(time.Millisecond))
	if utils.IsTTY() {
		fmt.Fprint(os.Stderr, "\x1b[0m")
	}

	return nil
}

func (cmd *SrcCommand) processExample(e Example) error {
	g := SrcGenerator{Example: e}
	fName := filepath.Join(cmd.Output, "src", g.Filename())
	out := g.Output()

	if cmd.GofmtSource {
		var buf bytes.Buffer
		buf.ReadFrom(out)

		bout, err := imports.Process(
			"example_test.go",
			buf.Bytes(),
			&imports.Options{
				AllErrors:  true,
				Comments:   true,
				FormatOnly: false,
				TabIndent:  true,
				TabWidth:   1,
			})
		if err != nil {
			if cmd.DebugSource {
				utils.PrintSourceWithErr(&buf, err)
			}
			return err
		}

		out = bytes.NewBuffer(bout)
	}

	if cmd.DebugSource {
		var (
			err error
			buf bytes.Buffer
			src io.Reader
			tee = io.TeeReader(out, &buf)
		)

		if utils.IsTTY() {
			fmt.Fprint(os.Stderr, "\x1b[2m")
		}
		fmt.Fprintln(os.Stderr, strings.Repeat("━", utils.TerminalWidth()))
		if utils.IsTTY() {
			fmt.Fprint(os.Stderr, "\x1b[0m")
		}

		if cmd.ColorizeSource {
			src, err = utils.Chromatize(tee)
			if err != nil {
				return fmt.Errorf("error syntax highligting the output: %s", err)
			}

			_, err = io.Copy(os.Stderr, src)
			if err != nil {
				return fmt.Errorf("error copying output: %s", err)
			}
		}

		if _, err = io.Copy(os.Stderr, tee); err != nil {
			return fmt.Errorf("error copying output: %s", err)
		}

		fmt.Fprintf(os.Stderr, "\n\n")

		out = &buf
	}

	if cmd.Output == "-" {
		if _, err := io.Copy(os.Stdout, out); err != nil {
			return fmt.Errorf("error copying output: %s", err)
		}
	} else {
		f, err := os.OpenFile(fName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("error creating file: %s", err)
		}
		if _, err = io.Copy(f, out); err != nil {
			return fmt.Errorf("error copying output: %s", err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("error closing file: %s", err)
		}
	}

	return nil
}

// Execute runs the command.
//
func (cmd *DocCommand) Execute() error {
	return nil
}
